package env

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/helpers/httpHelper"
	"github.com/newrelic/NrDiag/tasks"
)

func TestBaseEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Env/* test suite")
}

func mockSuccessfulRequest200(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockUnsuccessfulRequest400(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		StatusCode: 400,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockUnSuccessfulRequestErr(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{}, errors.New("Error! This is an error")
}

var _ = Describe("Base/Env/DetectAWS", func() {
	var p BaseEnvDetectAWS //instance of our task struct to be used in tests

	//Tests go here!
	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Env",
				Name:        "DetectAWS",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explanation of task", func() {
			Expect(p.Explain()).To(Equal("Detect if running in AWS environment"))
		})
	})

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("200 response received", func() {

			BeforeEach(func() {
				p.httpGetter = mockSuccessfulRequest200
			})

			It("Should return an expected result status of success", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("Should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Successfully detected AWS."))
			})
		})

		Context("400 response received", func() {

			BeforeEach(func() {
				p.httpGetter = mockUnsuccessfulRequest400
			})

			It("Should return an expected result status of none", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("Should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("AWS metadata endpoint timeout"))
			})
		})

		Context("Request error received", func() {

			BeforeEach(func() {
				p.httpGetter = mockUnSuccessfulRequestErr
			})

			It("Should return an expected result status of none", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("Should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("AWS metadata endpoint timeout"))
			})
		})
	})

	Describe("checkAWSMetaData()", func() {

		var metaDataResult bool

		JustBeforeEach(func() {
			metaDataResult = p.checkAWSMetaData()
		})

		Context("when AWS is detected", func() {
			BeforeEach(func() {
				p.httpGetter = mockSuccessfulRequest200
			})

			It("Should return true", func() {
				Expect(metaDataResult).To(Equal(true))
			})
		})

		Context("when AWS is not detected", func() {
			BeforeEach(func() {
				p.httpGetter = mockUnsuccessfulRequest400
			})

			It("Should return false", func() {
				Expect(metaDataResult).To(Equal(false))
			})
		})

		Context("Request error received", func() {

			BeforeEach(func() {
				p.httpGetter = mockUnSuccessfulRequestErr
			})

			It("Should return false", func() {
				Expect(metaDataResult).To(Equal(false))
			})
		})
	})

})
