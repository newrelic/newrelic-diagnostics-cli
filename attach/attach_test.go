package attach

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

var testServer *httptest.Server

func setup() {
	testServer = httptest.NewServer((http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/success" {
			w.WriteHeader(200)
			w.Write([]byte{})
		}
		if r.URL.Path == "/error" {
			w.WriteHeader(400)
			w.Write([]byte{})
		}
	})))
}

func teardown() {
	testServer.Close()
}

func TestUpload(t *testing.T) {
	setup()
	defer teardown()
	type args struct {
		identifyingKey string
		timestamp      string
		dependencies   IAttachDeps
	}
	tests := []struct {
		name     string
		args     args
		endpoint string
	}{
		{
			name: "Test successful Upload",
			args: args{
				identifyingKey: "TestKey",
				timestamp:      "TimestampTest",
				dependencies:   mocks.MAttachDeps{},
			},
			endpoint: "/success",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AttachmentEndpoint = testServer.URL + tt.endpoint
			Upload(tt.args.identifyingKey, tt.args.timestamp, tt.args.dependencies)
		})
	}
}

func Test_uploadFilesToAccount(t *testing.T) {
	setup()
	defer teardown()

	var files []UploadFiles
	zipfile := UploadFiles{
		Path:        "/",
		Filename:    "file1.zip",
		NewFilename: "file1-timestamp.zip",
		Filesize:    1000,
		URL:         "",
		Key:         "testKey",
	}
	jsonfile := UploadFiles{
		Path:        "/",
		Filename:    "file1.json",
		NewFilename: "file1-timestamp.json",
		Filesize:    1000,
		URL:         "",
		Key:         "testKey",
	}
	files = append(files, zipfile)
	files = append(files, jsonfile)

	type args struct {
		filesToUpload []UploadFiles
		attachmentKey string
		deps          IAttachDeps
	}
	tests := []struct {
		name     string
		args     args
		wantErr  bool
		endpoint string
	}{
		{
			name: "Test successful uploadFilesToAccount",
			args: args{
				filesToUpload: files,
				attachmentKey: "testKey",
				deps:          mocks.MAttachDeps{},
			},
			wantErr:  false,
			endpoint: "/success",
		},
		{
			name: "Test failed uploadFilesToAccount",
			args: args{
				filesToUpload: files,
				attachmentKey: "testKey",
				deps:          mocks.MAttachDeps{},
			},
			wantErr:  true,
			endpoint: "/error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.AttachmentEndpoint = testServer.URL + tt.endpoint
			if err := uploadFilesToAccount(tt.args.filesToUpload, tt.args.attachmentKey, tt.args.deps); (err != nil) != tt.wantErr {
				t.Errorf("uploadFilesToAccount() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAttachDeps_GetAttachmentsEndpoint(t *testing.T) {
	tests := []struct {
		name                    string
		attachmentEndpoint      string
		flagsAttachmentEndpoint string
		want                    string
	}{
		{
			name:                    "Test GetAttachmentEndpoint none set",
			attachmentEndpoint:      "",
			flagsAttachmentEndpoint: "",
			want:                    "http://localhost:3000/attachments",
		},
		{
			name:                    "Test GetAttachmentEndpoint config.Flags.AttachmentEndpoint set",
			attachmentEndpoint:      "",
			flagsAttachmentEndpoint: "http://diag.datanerd.us/attachments",
			want:                    "http://diag.datanerd.us/attachments",
		},
		{
			name:                    "Test GetAttachmentEndpoint config.AttachmentEndpoint set",
			attachmentEndpoint:      "http://diag.datanerd.us/attachments",
			flagsAttachmentEndpoint: "",
			want:                    "http://diag.datanerd.us/attachments",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Flags.AttachmentEndpoint = tt.flagsAttachmentEndpoint
			config.AttachmentEndpoint = tt.attachmentEndpoint
			a := AttachDeps{}
			if got := a.GetAttachmentsEndpoint(); got != tt.want {
				t.Errorf("AttachDeps.GetAttachmentsEndpoint() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttachDeps_GetWrapper(t *testing.T) {
	type args struct {
		file          *bytes.Reader
		fileSize      int64
		attachmentKey string
	}
	mockAttachDeps := mocks.MAttachDeps{}
	mockFile, _ := mockAttachDeps.GetReader("")
	tests := []struct {
		name string
		args args
		want httpHelper.RequestWrapper
	}{
		{
			name: "Test GetWrapper",
			args: args{
				file:          mockFile,
				fileSize:      mockFile.Size(),
				attachmentKey: "testKey",
			},
			want: httpHelper.RequestWrapper{
				Method:         "POST",
				URL:            "http://localhost:3000/attachments/upload_s3",
				Headers:        map[string]string{"Attachment-Key": "testKey"},
				Payload:        mockFile,
				Length:         mockFile.Size(),
				TimeoutSeconds: 7200,
				BypassProxy:    false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			a := AttachDeps{}
			if got := a.GetWrapper(tt.args.file, tt.args.fileSize, tt.args.attachmentKey); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AttachDeps.GetWrapper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAttachDeps_MakeRequest(t *testing.T) {
	setup()
	defer teardown()
	mockAttachDeps := mocks.MAttachDeps{}
	mockFile, _ := mockAttachDeps.GetReader("")
	type args struct {
		wrapper httpHelper.RequestWrapper
	}
	tests := []struct {
		name           string
		args           args
		wantStatusCode int
		wantErr        bool
	}{
		{
			name: "Test MakeRequest success",
			args: args{
				wrapper: httpHelper.RequestWrapper{
					Method:         "POST",
					URL:            testServer.URL + "/success",
					Headers:        map[string]string{"Attachment-Key": "testKey"},
					Payload:        mockFile,
					Length:         mockFile.Size(),
					TimeoutSeconds: 7200,
					BypassProxy:    false,
				},
			},
			wantStatusCode: 200,
			wantErr:        false,
		},
		{
			name: "Test MakeRequest fail",
			args: args{
				wrapper: httpHelper.RequestWrapper{
					Method:         "POST",
					URL:            testServer.URL + "/error",
					Headers:        map[string]string{"Attachment-Key": "testKey"},
					Payload:        mockFile,
					Length:         mockFile.Size(),
					TimeoutSeconds: 7200,
					BypassProxy:    false,
				},
			},
			wantStatusCode: 500,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := AttachDeps{}
			got, err := a.MakeRequest(tt.args.wrapper)
			if (err != nil) != tt.wantErr {
				t.Errorf("AttachDeps.MakeRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil {
				if !reflect.DeepEqual(got.StatusCode, tt.wantStatusCode) {
					t.Errorf("AttachDeps.MakeRequest() = %v, want %v", got, tt.wantStatusCode)
					return
				}
			}
		})
	}
}

// Legacy tests below

type Client struct {
	cli *http.Client
	URL string
}

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
			deps := AttachDeps{}
			result := deps.GetAttachmentsEndpoint()
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
			deps := AttachDeps{}
			result := deps.GetAttachmentsEndpoint()
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
			deps := AttachDeps{}
			result := deps.GetAttachmentsEndpoint()
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
	rawResponse, err := ioutil.ReadFile("../mocks/hdash-response-good-attachment.json")
	if err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(rawResponse))
}
func SuccessGetRequestLicenseKey(w http.ResponseWriter, r *http.Request) {
	rawResponse, err := ioutil.ReadFile("../mocks/hdash-response-good-license.json")
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
