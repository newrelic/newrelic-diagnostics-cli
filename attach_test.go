package main

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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

var _ = Describe("TestS3Upload", func() {
	var (
		filesToUpload  []uploadFiles
		attachmentKey  string
		expectedResult string
	)
	Context("With non s3 URL given", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			filesToUpload = []uploadFiles{
				uploadFiles{
					path:        "./",
					filename:    "nrdiag-output.json",
					newFilename: "nrdiag-output-2021-07-28T22:49:34Z.json",
					filesize:    0,
					URL:         "http://localhost:3000/attachments/upload",
					key:         "",
				},
			}
		})
		It("Should return a 500 internal server error", func() {
			result := uploadAWS(filesToUpload, attachmentKey)
			Expect(result).To(Equal(errors.New("Error uploading, status code was 500 Internal Server Error")))
		})
	})

	Context("With non s3 URL given", func() {
		BeforeEach(func() {
			config.Flags.AttachmentEndpoint = ""
			config.AttachmentEndpoint = ""
			filesToUpload = []uploadFiles{
				uploadFiles{
					path:        "./",
					filename:    "",
					newFilename: "",
					filesize:    0,
					URL:         "http://localhost:3000/attachments/upload",
					key:         "",
				},
			}
			expectedResult = "read " + filesToUpload[0].path + "/: is a directory"
		})
		It("Should return default attachment endpoint (localhost)", func() {
			result := uploadAWS(filesToUpload, attachmentKey)
			Expect(result.Error()).To(Equal(expectedResult))
		})
	})

})
