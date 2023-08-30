package attach

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/output/color"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

type UploadFiles struct {
	Path        string
	Filename    string
	NewFilename string
	Filesize    int64
	URL         string
	Key         string
}

type IAttachDeps interface {
	GetFileSize(file string) int64
	GetReader(file string) (*bytes.Reader, error)
	GetWrapper(endpoint string, file *bytes.Reader, fileSize int64, filename string, attachmentKey string) httpHelper.RequestWrapper
	GetUrlsToReturn(res *http.Response) (*string, error)
}

type AttachResponse struct {
	URL     string `json:"url"`
	Success bool   `json:"success"`
}

type AttachDeps struct{}

const awsUploadTimeoutSeconds = 7200
const defaultAttachmentEndpoint = "http://localhost:3000/attachments"

// Upload - takes the license key from ValidateLicenseKey
// and uploads the output to account
func Upload(endpoint string, identifyingKey string, timestamp string, dependencies IAttachDeps) {
	log.Debugf("Attempting to attach file with key: %s\n", identifyingKey)
	var filesToUpload []UploadFiles

	zipfile := getFilesForUpload(identifyingKey, timestamp, "zip", dependencies)
	jsonfile := getFilesForUpload(identifyingKey, timestamp, "json", dependencies)

	filesToUpload = append(filesToUpload, zipfile)
	filesToUpload = append(filesToUpload, jsonfile)

	if len(filesToUpload) == 0 {
		log.Debug("No files to upload.")
		return
	}

	log.Info(color.ColorString(color.White, "Uploading results to New Relic"))
	urls, err := uploadFilesToAccount(endpoint, filesToUpload, identifyingKey, dependencies)
	if err != nil {
		log.Fatalf("Error uploading large file: %s", err.Error())
	}
	printedUrls := make(map[string]bool)
	log.Info("Successfully uploaded to account!! Find your latest run here: ")
	for _, url := range urls {
		if !printedUrls[url] {
			infoStr := fmt.Sprintf("\t%v\n", url)
			filteredOutput := color.ColorString(color.LightBlue, infoStr)
			log.Infof(filteredOutput)
		}
		printedUrls[url] = true
	}

	log.Debug("Successfully uploaded to account")

}

func getFilesForUpload(identifyingKey string, timestamp string, filetype string, deps IAttachDeps) UploadFiles {
	thisFileName := "nrdiag-output." + filetype
	thisFile := UploadFiles{Path: config.Flags.OutputPath, Filename: thisFileName}
	thisFile.Filesize = deps.GetFileSize(thisFile.Path + "/" + thisFile.Filename)
	extension := filepath.Ext(thisFileName)
	shortName := thisFileName[0 : len(thisFileName)-len(extension)]
	thisFile.NewFilename = shortName + "-" + timestamp + extension
	log.Debug("Renamed file from", thisFileName, " to ", thisFile.NewFilename)
	return thisFile
}

func uploadFilesToAccount(endpoint string, filesToUpload []UploadFiles, attachmentKey string, deps IAttachDeps) ([]string, error) {
	var urlsToReturn []string
	for _, files := range filesToUpload {
		newUrl, err := uploadFile(endpoint, files, attachmentKey, deps)
		if err != nil {
			return nil, err
		}
		if !strings.Contains(files.Filename, ".zip") {
			urlsToReturn = append(urlsToReturn, *newUrl)
		}
	}
	return urlsToReturn, nil
}

func uploadFile(endpoint string, files UploadFiles, attachmentKey string, deps IAttachDeps) (*string, error) {
	log.Debug("Opening", filepath.Join(files.Path, files.Filename), "for upload")
	reader, err := deps.GetReader(filepath.Join(files.Path, files.Filename))
	if err != nil {
		log.Info("Error uploading", err)
		return nil, err
	}

	wrapper := deps.GetWrapper(endpoint, reader, files.Filesize, files.NewFilename, attachmentKey)

	log.Debug("Starting upload")
	res, err := makeRequest(wrapper)

	if err != nil {
		log.Info("Error uploading file", err)
		return nil, err
	}
	if res.StatusCode != 200 {
		log.Info("Error uploading, status code was", res.Status)
		body, _ := io.ReadAll(res.Body)
		log.Debug("Body was", string(body))
		log.Debug("headers were", res.Header)
		return nil, errors.New(res.Status)
	}
	log.Debug("Upload finished with status:  ", res.Status)
	newUrl, urlError := deps.GetUrlsToReturn(res)
	if urlError != nil {
		return nil, urlError
	}
	return newUrl, nil
}

func getAttachmentsEndpoint() string {
	if config.Flags.AttachmentEndpoint != "" { //If local development flag is supplied
		return config.Flags.AttachmentEndpoint
	} else if config.AttachmentEndpoint != "" { //Else if its a binary build
		return config.AttachmentEndpoint
	}
	log.Debug("No attachments endpoint supplied! Defaulting to localhost.") //This case should only be local dev without attachment flag
	return defaultAttachmentEndpoint
}

func makeRequest(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return httpHelper.MakeHTTPRequest(wrapper)
}

func (a AttachDeps) GetFileSize(file string) int64 {
	stat, err := os.Stat(file)
	if err != nil {
		log.Fatalf("Error getting filesize: %s", err.Error())
	}

	return stat.Size()
}

func (a AttachDeps) GetReader(file string) (*bytes.Reader, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		log.Info("Error uploading", err)
		return nil, err
	}
	return bytes.NewReader(data), err
}

func (a AttachDeps) GetWrapper(endpoint string, file *bytes.Reader, fileSize int64, filename string, attachmentKey string) httpHelper.RequestWrapper {
	wrapper := httpHelper.RequestWrapper{
		Method:         "POST",
		URL:            getAttachmentsEndpoint() + "/" + endpoint,
		Payload:        file,
		Length:         fileSize,
		TimeoutSeconds: awsUploadTimeoutSeconds,
	}
	wrapper.Params = url.Values{}
	wrapper.Params.Add("filename", filename)
	wrapper.Headers = make(map[string]string)
	wrapper.Headers["Attachment-Key"] = attachmentKey
	return wrapper
}

func (a AttachDeps) GetUrlsToReturn(res *http.Response) (*string, error) {
	bodyBytes, _ := io.ReadAll(res.Body)
	var bodyJson AttachResponse
	marshallErr := json.Unmarshal(bodyBytes, &bodyJson)
	if marshallErr != nil {
		return nil, marshallErr
	}
	return &bodyJson.URL, nil
}
