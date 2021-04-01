package output

import (
	"archive/zip"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const permissionsError = "\n------Error creating output files.------\nEnsure you have rights for creating files in the local directory or specify a different output directory with -output-path\nA 'permission denied' error may be solved by re-running this program prefixed by the command 'sudo -E'. The '-E' option will help preserve the environment variables needed for running this program."

type resultsOutput struct {
	RunDate       time.Time
	NRDiagVersion string
	Configuration interface{}
	Results       []registration.TaskResult
}

// wee bit of a hack for testing
var OutputNow = time.Now

//getResultsJSON takes in array of Result structs along with bool for indentation to be users. Outputs JSON of Results array -- if indented is true, output is nicely formatted.
func getResultsJSON(data []registration.TaskResult) string {

	outputData := resultsOutput{
		RunDate:       OutputNow(),
		NRDiagVersion: config.Version,
		Configuration: config.Flags,
		Results:       data,
	}

	output, err := json.MarshalIndent(outputData, "", "	")
	if err != nil {
		log.Info("Couldn't save JSON output: ", err)
	}
	return string(output)
}

//getEpoch returns epoch time (at time of fn call) as an int64
func getEpoch() int64 {
	now := time.Now().UTC().UnixNano()
	return (now / 1000000)
}

func outputJSON(json string) {
	jsonFile := filepath.Clean(config.Flags.OutputPath + "/nrdiag-output.json")
	log.Debug("Creating json file:", jsonFile)
	err := os.MkdirAll(config.Flags.OutputPath, 0777)
	if err != nil {
		log.Info("Error creating directory", err)
		log.Info(permissionsError)
	}
	ioutil.WriteFile(jsonFile, []byte(json), 0644)
}

func CreateZip() *zip.Writer {
	err := os.MkdirAll(config.Flags.OutputPath, 0777)
	if err != nil {
		log.Info("Error creating directory", err)
		log.Info(permissionsError)
	}
	zipfile, err := os.Create(config.Flags.OutputPath + "/nrdiag-output.zip")
	if err != nil {
		log.Info("Error creating zip file", err)
		log.Info(permissionsError)
	}

	// Create a new zip archive.
	w := zip.NewWriter(zipfile)
	return w
}

func CloseZip(zipfile *zip.Writer) {
	// All done, now close the zip file
	log.Debug("Done executing tasks, closing zip file")
	zipErr := zipfile.Close()
	if zipErr != nil {
		log.Info("error closing zip file: ", zipErr)
	}
}

func mapContains(set map[string]struct{}, item string) bool {
	_, ok := set[item]
	return ok
}

// CopyFilesToZip - Copies files to the zip archive
func copyFilesToZip(dst *zip.Writer, filesToZip []tasks.FileCopyEnvelope) {

	for _, envelope := range filesToZip {

		if envelope.Stream != nil {
			header := zip.FileHeader{
				Name:   filepath.ToSlash("nrdiag-output/" + envelope.StoreName()),
				Method: zip.Deflate,
			}

			writer, _ := dst.CreateHeader(&header)
			for s := range envelope.Stream {
				io.WriteString(writer, s)
			}
		} else {
			log.Debug("adding " + envelope.Path + " to zip")
			// Get file info from file
			stat, err := os.Stat(envelope.Path)
			if err != nil {
				log.Info("Error adding file to Diagnostics CLI zip file: ", err)
				return
			}
			// open file handle
			fileHandle, err := os.Open(envelope.Path)
			defer fileHandle.Close()

			if err != nil {
				log.Info("Error opening file handle", err)
				return
			}

			header, err := zip.FileInfoHeader(stat)
			if err != nil {
				log.Info("Error copying file", err)
				return
			}
			// Setting filename to deduplicated file name
			header.Name = envelope.StoreName()
			header.Name = filepath.ToSlash("nrdiag-output/" + header.Name) //Add folder to filename to unzip into a folder
			log.Debug("storing name", header.Name)

			// Change to deflate to gain better compression
			// see http://golang.org/pkg/archive/zip/#pkg-constants
			header.Method = zip.Deflate

			// write zip file header
			writer, err := dst.CreateHeader(header)

			if err != nil {
				log.Info("Error writing results to zip file: ", err)
				return
			}

			_, err = io.Copy(writer, fileHandle)
			if err != nil {
				log.Info("Error writing file into zip: ", err)
			}
		}

		//Add filepath and name to text file
		addFileToFileList(envelope)
	}
}

// This takes the fileToCopy item and appends the values to a text file to be included in the zip file to preserve filepaths
func addFileToFileList(file tasks.FileCopyEnvelope) {
	f, err := os.OpenFile(config.Flags.OutputPath+"/nrdiag-filelist.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		log.Info("Error writing output file", err)
		log.Info(permissionsError)
	}
	defer f.Close()

	if _, err = f.WriteString("\nStored file name:" + file.StoreName() + "\nOriginal path:" + file.Path + "\r\n"); err != nil {
		log.Info("Error writing output file", err)
	}
}

//filteredResult - Checks if a result status (e.g. "Warning") was passed in the -filter flag. Returns a boolean
func filteredResult(result string) bool {
	lowercaseFilter := strings.Replace(strings.ToLower(config.Flags.Filter), " ", "", -1)

	if lowercaseFilter == "all" {
		return true
	}

	filters := strings.Split(lowercaseFilter, ",")

	for _, filter := range filters {
		if filter == strings.ToLower(result) {
			return true
		}
	}
	return false
}

//filteredToString - Takes an array of ints corresponding to the 5 statuses, with a counter for each: array[status] = status count
// returns a string summary of instances:
// IN: [3,1,0,0,2]
// OUT: 3 Success, 1 Warning, 2 None
func filteredToString(filtered [6]int) string {
	var outputStrings []string
	for i, value := range filtered {
		if value != 0 {
			outputStrings = append(outputStrings, strconv.Itoa(value)+" "+tasks.Status(i).StatusToString())
		}
	}
	return strings.Join(outputStrings, ", ")
}
