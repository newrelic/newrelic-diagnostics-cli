package attach

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/output/color"
	"gopkg.in/cheggaaa/pb.v1"

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

type jsonResponse struct {
	URL   string `json:"url"`
	Key   string `json:"key"`
	Error string `json:"error"`
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

	log.Debug("Uploading to account")
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
		newUrl, err := uploadFileMultipart(endpoint, files, attachmentKey, deps)
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
	log.Debug("Opening", files.Path+"/"+files.Filename, "for upload")
	reader, err := deps.GetReader(files.Path + "/" + files.Filename)
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
		body, _ := ioutil.ReadAll(res.Body)
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

func uploadFileMultipart(endpoint string, files UploadFiles, attachmentKey string, deps IAttachDeps) (*string, error) {
	path := files.Path + "/" + files.Filename
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pipeOut, pipeIn := io.Pipe()
	bar := pb.New(int(files.Filesize)).SetUnits(pb.U_BYTES)
	writer := multipart.NewWriter(pipeIn)
	done := make(chan error)
	var res *http.Response
	go func() {
		url := getAttachmentsEndpoint() + "/" + endpoint + "?filename=" + files.NewFilename
		req, err := http.NewRequest(http.MethodPost, url, pipeOut)
		if err != nil {
			done <- err
			return
		}

		req.Header.Add("Content-Type", writer.FormDataContentType())
		req.Header.Add("Attachment-Key", attachmentKey)
		req.Header.Set("User-Agent", "Nrdiag_/"+config.Version)

		bar.Start()

		res, err = http.DefaultClient.Do(req)
		if err != nil {
			log.Info("Error uploading file", err)
			done <- err
			return
		}

		done <- nil
	}()

	partWriter, err := writer.CreateFormFile("file", path)
	if err != nil {
		return nil, err
	}
	out := io.MultiWriter(partWriter, bar)
	_, err = io.Copy(out, f)
	if err != nil {
		return nil, err
	}

	if err = writer.Close(); err != nil {
		return nil, err
	}

	if err = pipeIn.Close(); err != nil {
		return nil, err
	}

	if err = <-done; err != nil {
		return nil, err
	}

	bar.Finish()
	if res.StatusCode != 200 {
		log.Info("Error uploading, status code was", res.Status)
		body, _ := ioutil.ReadAll(res.Body)
		log.Debug("Body was", string(body))
		log.Debug("headers were", res.Header)
		return nil, errors.New(res.Status)
	}
	defer res.Body.Close()

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
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Info("Error uploading", err)
		return nil, err
	}
	return bytes.NewReader(data), err
}

func (a AttachDeps) GetWrapper(endpoint string, file *bytes.Reader, fileSize int64, filename string, attachmentKey string) httpHelper.RequestWrapper {
	headers := make(map[string]string)
	headers["Attachment-Key"] = attachmentKey

	wrapper := httpHelper.RequestWrapper{
		Method:         "POST",
		URL:            getAttachmentsEndpoint() + "/" + endpoint + "?filename=" + filename,
		Payload:        file,
		Length:         fileSize,
		TimeoutSeconds: awsUploadTimeoutSeconds,
	}
	wrapper.Headers = headers
	return wrapper
}

func (a AttachDeps) GetUrlsToReturn(res *http.Response) (*string, error) {
	bodyBytes, _ := ioutil.ReadAll(res.Body)
	var bodyJson AttachResponse
	marshallErr := json.Unmarshal(bodyBytes, &bodyJson)
	if marshallErr != nil {
		return nil, marshallErr
	}
	return &bodyJson.URL, nil

}

// All the functions below are legacy and may be removed at any time

// UploadLegacy - takes the license key from ValidateLicenseKey
// and uploads the output to s3. This is the legacy method of uploading
func UploadLegacy(identifyingKey string, timestamp string) {
	log.Debugf("Attempting to attach file with key: %s\n", identifyingKey)
	log.Debugf("argument zero: %s\n", os.Args[0])
	// look at our command name, should be 'nrdiag' in production
	var filesToUpload []UploadFiles

	//files to be uploaded to s3
	s3zipfile := getS3UploadFiles(identifyingKey, timestamp, "zip")
	s3jsonfile := getS3UploadFiles(identifyingKey, timestamp, "json")

	filesToUpload = append(filesToUpload, s3zipfile)
	filesToUpload = append(filesToUpload, s3jsonfile)

	uploadFilelist(identifyingKey, filesToUpload)
}

func getS3UploadFiles(identifyingKey string, timestamp string, filetype string) UploadFiles {
	thisFileName := "nrdiag-output." + filetype
	thisFile := UploadFiles{Path: config.Flags.OutputPath, Filename: thisFileName}
	thisFile.Path = config.Flags.OutputPath
	thisFile.Filename = thisFileName
	stat, err := os.Stat(thisFile.Path + "/" + thisFile.Filename)
	if err != nil {
		log.Fatalf("Error getting filesize: %s", err.Error())
	}
	thisFile.Filesize = stat.Size()
	thisFile.NewFilename = datestampFile(thisFileName, timestamp)

	// Get upload URL for file
	requestURL := buildGetRequestURL(thisFile.NewFilename, identifyingKey, thisFile.Filesize)
	jsonResponse, err := getUploadURL(requestURL)
	if err != nil {
		log.Fatalf("Unable to retrieve upload URL: %s\nIf you can see the nrdiag output in your directory, consider manually uploading it to your support ticket\nIf you want to upload it to your account, use the -a option", err.Error())
	}
	thisFile.URL = jsonResponse.URL
	if jsonResponse.Key != "" {
		thisFile.Key = jsonResponse.Key
	}
	log.Debug("This file upload URL is ", thisFile.URL)

	return thisFile
}

func uploadFilelist(attachmentKey string, filesForAWS []UploadFiles) {
	if len(filesForAWS) != 0 {
		log.Debug("Uploading to AWS")
		AWSErr := uploadAWS(filesForAWS, attachmentKey)
		if AWSErr != nil {
			log.Fatalf("Error uploading large file: %s", AWSErr.Error())
		}
		log.Debug("Successfully uploaded to AWS")
	}
}

func uploadAWS(filesToUpload []UploadFiles, attachmentKey string) error {
	for _, files := range filesToUpload {
		log.Debug("opening", files.Path+"/"+files.Filename, "to upload to S3")

		data, err := ioutil.ReadFile(files.Path + "/" + files.Filename)
		if err != nil {
			log.Info("Error uploading", err)
			return err
		}
		reader := bytes.NewReader(data)

		log.Debug("Starting upload to ", files.URL)
		wrapper := httpHelper.RequestWrapper{
			Method:         "PUT",
			URL:            files.URL,
			Payload:        reader,
			Length:         files.Filesize,
			TimeoutSeconds: awsUploadTimeoutSeconds,
		}

		res, err := httpHelper.MakeHTTPRequest(wrapper)

		if err != nil {
			log.Info("Error uploading file", err)
			return err
		}
		if res.StatusCode != 200 {
			log.Info("Error uploading to AWS, status code was", res.Status)
			body, _ := ioutil.ReadAll(res.Body)
			log.Debug("Body was", string(body))
			log.Debug("headers were", res.Header)
			return errors.New("error uploading, status code was " + res.Status)
		}
		log.Debug(res.Status, "was status code to AWS upload")

	}
	return nil
}

func getUploadURL(requestURL string) (jsonResponse, error) {
	log.Debug("Making http request to ", requestURL)

	wrapper := httpHelper.RequestWrapper{
		Method: "GET",
		URL:    requestURL,
	}

	res, err := httpHelper.MakeHTTPRequest(wrapper)
	if err != nil {
		return jsonResponse{}, err
	}

	if res.StatusCode != http.StatusOK {
		return jsonResponse{}, fmt.Errorf("got %v status code from %s", res.StatusCode, requestURL)
	}

	bodyBytes, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		return jsonResponse{}, readErr
	}

	var data jsonResponse
	jsonParseError := json.Unmarshal(bodyBytes, &data)

	log.Debugf("Response is: %s\n", string(bodyBytes))
	if res.StatusCode != http.StatusOK {
		if data.Error != "" {
			return jsonResponse{}, fmt.Errorf("(%v status) %s", res.StatusCode, data.Error)
		}
		return jsonResponse{}, fmt.Errorf("received %v response code", res.StatusCode)
	}

	//Checking for json parse error after checking response code since we want to give
	//errors surfaced to the user (after network/body read errors) in the following priority:
	// 	1. Error message returned from server (data.Error)
	//  2. !200 response code
	//  3. json parsing error
	if jsonParseError != nil {
		return jsonResponse{}, fmt.Errorf("error parsing response json: %s", jsonParseError.Error())
	}

	_, err = url.ParseRequestURI(data.URL)
	if err != nil {
		return jsonResponse{}, fmt.Errorf("invalid URL: '%s'", data.URL)
	}

	return data, nil
}

func buildGetRequestURL(filename, attachmentKey string, filesize int64) string {
	requestURL := getAttachmentsEndpoint() + "/upload_url"
	log.Debug("Making call to get zip file endpoint")

	// Now add the parameters to the URL
	requestURL += "?attachment_key=" + attachmentKey
	requestURL += "&filename=" + filename
	requestURL += "&filesize=" + strconv.FormatInt(filesize, 10)

	return requestURL
}

func datestampFile(originalFile, timestamp string) string {
	extension := filepath.Ext(originalFile)
	shortName := originalFile[0 : len(originalFile)-len(extension)]
	newName := shortName + "-" + timestamp + extension

	log.Debug("Renamed file from", originalFile, " to ", newName)
	return newName
}
