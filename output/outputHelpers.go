package output

import (
	"archive/zip"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/output/color"
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

type (
	writeFunction func(path string, info os.FileInfo, zipfile *zip.Writer) error
)

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

func outputJSON(json string) {
	jsonFile := filepath.Clean(config.Flags.OutputPath + "/nrdiag-output.json")
	log.Debug("Creating json file:", jsonFile)
	err := os.MkdirAll(config.Flags.OutputPath, 0777)
	if err != nil {
		log.Info("Error creating directory", err)
		log.Info(permissionsError)
	}
	_ = ioutil.WriteFile(jsonFile, []byte(json), 0644)
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
		//Add filepath and name to text file
		addFileToFileList(envelope)

		if envelope.Stream != nil {
			header := zip.FileHeader{
				Name:   filepath.ToSlash("nrdiag-output/" + envelope.StoreName()),
				Method: zip.Deflate,
			}

			writer, _ := dst.CreateHeader(&header)
			for s := range envelope.Stream {
				_, _ = io.WriteString(writer, s)
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
			if err != nil {
				log.Info("Error opening file handle", err)
				return
			}
			defer fileHandle.Close()

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

//CreateFileList - Create output file and wipe out if it already exists
func CreateFileList() error {
	err := ioutil.WriteFile(config.Flags.OutputPath+"/nrdiag-filelist.txt", []byte(""), 0644)
	if err != nil {
		return err
	} else {
		_ = ioutil.WriteFile(config.Flags.OutputPath+"/nrdiag-filelist.txt", []byte("List of files in zipfile"), 0644)
	}
	return nil
}

func WalkSizeFunction(info os.FileInfo) int64 {
	return info.Size()
}
func WalkCopyFunction(path string, info os.FileInfo, err error, zipfile *zip.Writer, writeToZip writeFunction) error {
	if err != nil {
		log.Infof(color.ColorString(color.Yellow, "Could not add %s to zip.\n"), path)
		return err
	}
	if info.IsDir() {
		return nil
	}
	if strings.HasSuffix(info.Name(), ".exe") {
		log.Infof(color.ColorString(color.Yellow, "Could not add %s to zip.\n"), path)
		log.Info(color.ColorString(color.Yellow, "Executable files are not allowed to be included in the zip.\n"))
		return errors.New("cannot add executable files")
	}

	ok := writeToZip(path, info, zipfile)
	if ok != nil {
		return ok
	}

	log.Infof("Adding file to Diagnostics CLI zip file: %s\n", path)
	addFileToFileList(tasks.FileCopyEnvelope{
		Path: path,
	})

	return nil
}

func WriteFileToZip(path string, info os.FileInfo, zipfile *zip.Writer) error {
	file, ok := os.Open(path)
	if ok != nil {
		return ok
	}
	defer file.Close()

	header, ok := zip.FileInfoHeader(info)
	if ok != nil {
		log.Info("Error copying file", ok)
		return ok
	}

	header.Name = filepath.ToSlash("nrdiag-output/Include/" + path)
	header.Method = zip.Deflate
	writer, ok := zipfile.CreateHeader(header)
	if ok != nil {
		log.Info("Error writing results to zip file: ", ok)
		return ok
	}
	_, ok = io.Copy(writer, file)
	if ok != nil {
		return ok
	}
	return nil
}
