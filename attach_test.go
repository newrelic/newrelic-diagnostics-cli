package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/newrelic/newrelic-diagnostics-cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

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
				result := getTicketUploadFile(attachmentKey, timestamp)
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

func TestGetUploadURLPath(t *testing.T) {
	t.Parallel()
	// setup()
	// defer teardown()
	testAPIEndpoint := "/attachments/upload_url?attachment_key=1Q3OS5O1ffsd2345678901t56789014&filename=nrdiag-output.zip&filesize=234567"

	req, _ := http.NewRequest("GET", testAPIEndpoint, nil)
	writer := httptest.NewRecorder()

	thisRouter().ServeHTTP(writer, req)
	fmt.Println(req.URL)
	assert.Equal(t, http.StatusOK, writer.Code)
}

func thisRouter() *mux.Router {
	r := mux.NewRouter().StrictSlash(true)
	r.Path("/attachments/upload_url").
		Queries("attachment_key", "{[a-zA-Z0-9]+}").
		Queries("filename", "{nrdiag-output.zip|nrdiag-output.json}").
		Queries("filesize", "{[0-9]+}").
		HandlerFunc(GetRequest).
		Name("/attachments/upload_url")
	r.HandleFunc("/attachments/upload_url/{attachment_key:{[a-zA-Z0-9]+}}{filename:{nrdiag-output.zip|nrdiag-output.json}}{filesize:{[0-9]+}}", GetRequest).Name("/attachments/upload_url").Methods("GET")
	http.Handle("/", r)
	return r
}

func GetRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	myString := vars["mystring"]

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(myString))
}
