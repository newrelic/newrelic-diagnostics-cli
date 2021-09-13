package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/newrelic/newrelic-diagnostics-cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

type Client struct {
	cli *http.Client
	URL string
}

var _ = Describe("GetTicketFile", func() {
	var (
		attachmentKey  string
		timestamp      string
		expectedResult uploadFiles
	)
	Describe("Testing proper ticket upload file", func() {
		BeforeEach(func() {
			attachmentKey = "12345678912345678912345678912345"
			timestamp = "2021-07-28T22:49:34Z"
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			expectedResult = uploadFiles{
				path:        "./",
				filename:    "nrdiag-output.json",
				newFilename: "nrdiag-output-2021-07-28T22:49:34Z.json",
				filesize:    0,
				URL:         "http://localhost:3000/attachments/upload",
				key:         "",
			}
			config.Flags.OutputPath = "./"
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
		})
		Context("With proper JSON format", func() {
			It("Should return a JSON", func() {
				result := getTicketUploadFile(attachmentKey, timestamp, "json")
				Expect(result).To(Equal(expectedResult))
			})
		})
	})
})

var _ = Describe("GetAttachmentEndpoints", func() {
	var (
		expectedResult string
	)
	Context("With no attachmentEndpointsSet", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			expectedResult = "http://localhost:3000/attachments"
		})
		It("Should return default attachment endpoint (localhost)", func() {
			result := getAttachmentsEndpoint()
			Expect(result).To(Equal(expectedResult))
		})
	})
	Context("With flag attachment endpoints set", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = "http://diag.datanerd.us/attachments"
			config.AttachmentEndpoint = ""
			expectedResult = "http://diag.datanerd.us/attachments"
		})
		It("Should return flag attachment endpoint", func() {
			result := getAttachmentsEndpoint()
			Expect(result).To(Equal(expectedResult))
		})
	})
	Context("With attachment endpoint set", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = "http://diag.datanerd.us/attachments"
			expectedResult = "http://diag.datanerd.us/attachments"
		})
		It("Should return attachment endpoint", func() {
			result := getAttachmentsEndpoint()
			Expect(result).To(Equal(expectedResult))
		})
	})

})

var _ = Describe("buildGetRequestURL", func() {
	var (
		expectedResult string
	)
	Context("with filename, attachmentKey, and filesize set", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			expectedResult = "http://localhost:3000/attachments/upload_url?attachment_key=1Q3OS5O1ffsd2345678901t56789014&filename=nrdiag-output.json&filesize=23647"
		})
		It("Should return default attachment endpoint (localhost)", func() {
			result := buildGetRequestURL("nrdiag-output.json", "1Q3OS5O1ffsd2345678901t56789014", 23647)
			Expect(result).To(Equal(expectedResult))
		})
	})

})

func TestGoodAttachmentUploadURL(t *testing.T) {
	r := thisRouter(SuccessGetRequestAttachmentKey)

	s := httptest.NewServer(r)
	defer s.Close()

	thisClient := &Client{
		cli: s.Client(),
		URL: s.URL,
	}
	testAPIEndpoint := thisClient.URL + "/attachments/upload_url?attachment_key=1Q3OS5O1ffsd2345678901t5F789014R&filename=nrdiag-output.zip&filesize=23456"
	expectedResponse := jsonResponse{
		URL: "https://mock-bucket.s3.amazonaws.com/staging/tickets/543210/attachments/12345678-abcd-9876-fedc-abcdefabcdef/nrdiag-output-2021-04-29T05:15:00Z.json",
		Key: "tickets/543210/12345678-abcd-9876-fedc-abcdefabcdef/nrdiag-output-2021-04-29T05:15:00Z.json",
	}

	req, err := http.NewRequest("GET", testAPIEndpoint, nil)
	if err != nil {
		panic(err)
	}

	res, err := thisClient.cli.Do(req)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		panic(err)
	}
	var MyResult jsonResponse
	err = json.Unmarshal(buf.Bytes(), &MyResult)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, expectedResponse, MyResult)
}

func TestBadAttachmentUploadURL(t *testing.T) {
	r := thisRouter(FailureGetRequest)

	s := httptest.NewServer(r)
	defer s.Close()

	thisClient := &Client{
		cli: s.Client(),
		URL: s.URL,
	}
	testAPIEndpoint := thisClient.URL + "/attachments/upload_url?attachment_key=badattachmentkey&filename=nrdiag-output.json&filesize=34256"

	req, err := http.NewRequest("GET", testAPIEndpoint, nil)
	if err != nil {
		panic(err)
	}

	res, err := thisClient.cli.Do(req)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestGoodLicenseUploadURL(t *testing.T) {
	r := thisRouter(SuccessGetRequestLicenseKey)

	s := httptest.NewServer(r)
	defer s.Close()

	thisClient := &Client{
		cli: s.Client(),
		URL: s.URL,
	}
	testAPIEndpoint := thisClient.URL + "/attachments/upload_url?attachment_key=1Q3OS5O1ffsd2345678901t5F789014RFf87AsS0&filename=nrdiag-output.json&filesize=23456"
	expectedResponse := jsonResponse{
		URL: "https://mock-bucket.s3.amazonaws.com/staging/accounts/123456789/attachments/12345678-abcd-9876-fedc-abcdefabcdef/nrdiag-output-2021-04-29T05:15:00Z.json",
		Key: "accounts/123456789/12345678-abcd-9876-fedc-abcdefabcdef/nrdiag-output-2021-04-29T05:15:00Z.json",
	}

	req, err := http.NewRequest("GET", testAPIEndpoint, nil)
	if err != nil {
		panic(err)
	}

	res, err := thisClient.cli.Do(req)
	if err != nil {
		panic(err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(res.Body)
	if err != nil {
		panic(err)
	}
	var MyResult jsonResponse
	err = json.Unmarshal(buf.Bytes(), &MyResult)
	if err != nil {
		panic(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, expectedResponse, MyResult)
}

func thisRouter(thisRequest func(w http.ResponseWriter, r *http.Request)) *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Path("/attachments/upload_url").
		Queries("attachment_key", "{attachment_key:[a-zA-Z0-9]{32}|[a-zA-Z0-9]{40}}", "filename", "{filename:nrdiag-output.json|nrdiag-output.zip}", "filesize", "{filesize:[0-9]+}").
		HandlerFunc(thisRequest).
		Name("/attachments/upload_url")

	return r
}

func SuccessGetRequestAttachmentKey(w http.ResponseWriter, r *http.Request) {
	rawResponse, err := ioutil.ReadFile("./mocks/hdash-response-good-attachment.json")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(rawResponse))
}
func SuccessGetRequestLicenseKey(w http.ResponseWriter, r *http.Request) {
	rawResponse, err := ioutil.ReadFile("./mocks/hdash-response-good-license.json")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(rawResponse))
}

func FailureGetRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "application/json")
}
