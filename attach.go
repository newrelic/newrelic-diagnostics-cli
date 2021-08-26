package main

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

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

type uploadFiles struct {
	path        string
	filename    string
	newFilename string
	filesize    int64
	URL         string
	key         string
}

type jsonResponse struct {
	URL   string `json:"url"`
	Key   string `json:"key"`
	Error string `json:"error"`
}

const ticketAttachmentUploadTimeoutSeconds = 600
const awsUploadTimeoutSeconds = 7200
const defaultAttachmentEndpoint = "http://localhost:3000/attachments"
const awsS3Domain = "s3.amazonaws.com"

func getAttachmentsEndpoint() string {
	if config.Flags.AttachmentEndpoint != "" { //If local development flag is supplied
		return config.Flags.AttachmentEndpoint
	} else if config.AttachmentEndpoint != "" { //Else if its a binary build
		return config.AttachmentEndpoint
	}
	log.Debug("No attachments endpoint supplied! Defaulting to localhost.") //This case should only be local dev without attachment flag
	return defaultAttachmentEndpoint
}

func addFileToForm(originalFilename string, newFilename string, i int, w *multipart.Writer) {
	f, err := os.Open(originalFilename)
	if err != nil {
		log.Debug("error", err)
		return
	}
	defer f.Close()

	log.Debugf("uploading %s as %s...\n", originalFilename, newFilename)
	fw, err := w.CreateFormFile("file"+fmt.Sprint(i), newFilename)
	if err != nil {
		log.Debug("error", err)
		return
	}
	if _, err = io.Copy(fw, f); err != nil {
		log.Debug("error", err)
		return
	}
}

// Upload - takes the attachment key from a ticket OR license key from ValidateLicenseKey
// and uploads the output to that ticket/s3
func Upload(identifyingKey string, timestamp string) {
	log.Debugf("Attempting to attach file with key: %s\n", identifyingKey)
	log.Debugf("argument zero: %s\n", os.Args[0])
	// look at our command name, should be 'nrdiag' in production
	var filesToUpload []uploadFiles

	s3zipfile := getS3UploadFiles(identifyingKey, timestamp, "zip")
	s3jsonfile := getS3UploadFiles(identifyingKey, timestamp, "json")
	//ticketUploadFile will upload the JSON file to the support ticket
	//currently it uploads all 3 of these files because Haberdasher returns the default URL
	//as a https://diagnostics... as opposed to https://s3.amazonaws...
	ticketUploadFile := getTicketUploadFile(identifyingKey, timestamp)
	// Calculate the filename just once

	filesToUpload = append(filesToUpload, s3zipfile)
	filesToUpload = append(filesToUpload, s3jsonfile)
	filesToUpload = append(filesToUpload, ticketUploadFile)

	uploadFilelist(identifyingKey, filesToUpload)
}

func getS3UploadFiles(identifyingKey string, timestamp string, filetype string) uploadFiles {
	thisFileName := "nrdiag-output." + filetype
	thisFile := uploadFiles{path: config.Flags.OutputPath, filename: thisFileName}
	thisFile.path = config.Flags.OutputPath
	thisFile.filename = thisFileName
	stat, err := os.Stat(thisFile.path + "/" + thisFile.filename)
	if err != nil {
		log.Fatalf("Error getting filesize: %s", err.Error())
	}
	thisFile.filesize = stat.Size()
	thisFile.newFilename = datestampFile(thisFileName, timestamp)

	// Get upload URL for file
	requestURL := buildGetRequestURL(thisFile.newFilename, identifyingKey, thisFile.filesize)
	jsonResponse, err := getUploadURL(requestURL)
	if err != nil {
		log.Fatalf("Unable to retrieve upload URL: %s\nIf you can see the nrdiag output in your directory, consider manually uploading it to your support ticket\n", err.Error())
	}
	thisFile.URL = jsonResponse.URL
	if jsonResponse.Key != "" {
		thisFile.key = jsonResponse.Key
	}
	log.Debug("This file upload URL is ", thisFile.URL)

	return thisFile
}

func getTicketUploadFile(attachmentKey string, timestamp string) uploadFiles {
	jsonfile := uploadFiles{path: config.Flags.OutputPath, filename: "nrdiag-output.json"}
	jsonfile.path = config.Flags.OutputPath
	jsonfile.filename = "nrdiag-output.json"
	jsonfile.newFilename = datestampFile("nrdiag-output.json", timestamp)

	jsonfile.URL = getAttachmentsEndpoint() + "/upload"

	return jsonfile
}

func uploadFilelist(attachmentKey string, filelist []uploadFiles) {
	var filesForTicketAttachment, filesForAWS []uploadFiles

	for _, upload := range filelist {
		if strings.Contains(upload.URL, awsS3Domain) {
			filesForAWS = append(filesForAWS, upload)
		} else {
			filesForTicketAttachment = append(filesForTicketAttachment, upload)
		}
	}

	log.Debug("AWS files found", len(filesForAWS))
	log.Debug("Ticket attachment files found", len(filesForTicketAttachment))

	if len(filesForAWS) != 0 {
		log.Debug("Uploading to AWS")
		log.Info("Uploading to AWS")
		AWSErr := uploadAWS(filesForAWS, attachmentKey)
		if AWSErr != nil {
			log.Fatalf("Error uploading large file: %s", AWSErr.Error())
		}
		log.Debug("Successfully uploaded to AWS, adding completed files for ticket attachment upload")
	}

	//length of an attachment key is 32 and if both attach and attachment key are provided, then this will check if what is begin passed through is an attachment key
	if len(filesForTicketAttachment) != 0 && len(attachmentKey) == 32 {
		log.Debug("Uploading to Haberdasher for ticket attachment")
		log.Info("Uploading to Haberdasher for ticket attachment")
		attachErr := uploadTicketAttachments(filesForTicketAttachment, attachmentKey)
		if attachErr != nil {
			log.Fatalf("Error uploading file to New Relic Support: %s", attachErr.Error())
		}
	}
}

func uploadTicketAttachments(filesToUpload []uploadFiles, attachmentKey string) error {

	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add the other fields
	fw, err := w.CreateFormField("attachment_key")
	if err != nil {
		log.Debug("Error creating form field")
		return err
	}
	if _, err = fw.Write([]byte(attachmentKey)); err != nil {
		log.Debug("Error creating form field")
		return err
	}
	var filelist string
	for i, upload := range filesToUpload {
		// First check to see if key exists
		if upload.key != "" {
			s3key, err := w.CreateFormField("S3key")
			if err != nil {
				log.Debug("Error creating form field")
				return err
			}
			if _, err = s3key.Write([]byte(upload.key)); err != nil {
				log.Debug("Error creating form field")
				return err
			}
		} else {
			addFileToForm(upload.path+"/"+upload.filename, upload.newFilename, i, w)
			filelist += upload.newFilename + ","
		}
	}

	fl, err := w.CreateFormField("filelist")
	if err != nil {
		log.Debug("Error creating form field")
		return err
	}

	if _, err = fl.Write([]byte(filelist)); err != nil {
		log.Debug("Error creating form field for filelist")
		return err
	}

	// Don't forget to set the content type, this will contain the boundary.
	headers := make(map[string]string)
	headers["Content-Type"] = w.FormDataContentType()
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()
	url := getAttachmentsEndpoint() + "/upload"
	log.Debug("URL is", url)
	// Submit the request

	wrapper := httpHelper.RequestWrapper{
		Method:         "POST",
		URL:            url,
		Payload:        &b,
		Headers:        headers,
		TimeoutSeconds: ticketAttachmentUploadTimeoutSeconds,
	}

	res, err := httpHelper.MakeHTTPRequest(wrapper)

	if err != nil {
		log.Info("Failed upload: " + err.Error())
		return err
	}

	bodyBytes, _ := ioutil.ReadAll(res.Body)
	bodyString := string(bodyBytes)
	log.Debugf("Reponse: %s\n", bodyString)

	var data jsonResponse
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		//Not returning here as this doesn't necessarily mean the upload failed.
		log.Debugf("Error parsing json response when attempting to upload ticket attachments: %s\n", err.Error())
	}

	if res.StatusCode != http.StatusOK {
		if data.Error != "" {
			return fmt.Errorf("(%v status) %s", res.StatusCode, data.Error)
		}
		return fmt.Errorf("received %v response code", res.StatusCode)
	}
	return nil

}

func uploadAWS(filesToUpload []uploadFiles, attachmentKey string) error {

	for _, files := range filesToUpload {
		log.Debug("opening", files.path+"/"+files.filename, "to upload to S3")

		data, err := ioutil.ReadFile(files.path + "/" + files.filename)
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
			Length:         files.filesize,
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
			return errors.New("Error uploading, status code was " + res.Status)
		}
		log.Debug(res.Status, "was status code to AWS upload")

	}
	return nil
}

func datestampFile(originalFile, timestamp string) string {
	extension := filepath.Ext(originalFile)
	shortName := originalFile[0 : len(originalFile)-len(extension)]
	newName := shortName + "-" + timestamp + extension

	log.Debug("Renamed file from", originalFile, " to ", newName)
	return newName
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
		return jsonResponse{}, fmt.Errorf("Got %v status code from %s", res.StatusCode, requestURL)
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
		return jsonResponse{}, fmt.Errorf("Invalid URL: '%s'", data.URL)
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
