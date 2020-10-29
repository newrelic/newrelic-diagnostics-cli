package agent

import (
	"testing"

	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type mockReader struct{}

func (p mockReader) Read([]byte) (int, error) {
	return 0, errors.New("Banana")
}

func (p mockReader) Close() error {
	return errors.New("Banana")
}

func TestInfraAgentConnect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra/Agent/Connect test suite")
}

var _ = Describe("Infra/Agent/Connect", func() {

	var p InfraAgentConnect

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Agent",
				Name:        "Connect",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Check network connection to New Relic Infrastructure collector endpoint"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return a slice ", func() {
			expectedDependencies := []string{
				"Base/Config/ProxyDetect",
				"Base/Config/RegionDetect",
				"Infra/Config/Agent",
			}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
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

		Context("If Infrastructure agent config is not present", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Failure,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Payload: []string{"us01"},
					},
				}
			})
			It("Should fail and return expected result", func() {

				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "Infrastructure Agent config not present.",
				}

				Expect(result).To(Equal(expectedResult))

			})
		})

		Context("If there is a network error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Payload: []string{"us01"},
					},
				}

				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{Body: mockReader{}}, errors.New("Failed request (timeout)")
				}
			})
			It("Should return a failed status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should have a summary with network error message", func() {
				Expect(result.Summary).To(ContainSubstring("Failed request (timeout)"))
			})
		})

		Context("If there was an error reading the response body", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Payload: []string{"us01"},
					},
				}

				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						Body:       mockReader{},
						StatusCode: 404,
						Status:     "404 Not Found",
					}, nil
				}
			})
			It("Should return a failed status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should have a summary with body read error message", func() {
				Expect(result.Summary).To(ContainSubstring("Banana"))
			})
		})

		Context("If status code is not expected", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/ProxyDetect": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Payload: []string{"us01"},
					},
				}

				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						StatusCode: 403,
						Status:     "Forbidden",
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
					}, nil
				}
			})

			It("Should return a failed status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should have a summary with unexpected response status and code", func() {
				Expect(result.Summary).To(ContainSubstring("Unexpected Response: 403 Forbidden"))
			})
		})
		Context("If region detect provides 0 regions", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/ProxyDetect": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Payload: []string{},
					},
				}
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Status:     "404 Not Found",
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
					}, nil
				}
			})
			It("Should have checked 2 region endpoints", func() {
				payload, _ := result.Payload.(map[string]string)
				Expect(len(payload)).To(Equal(2))
			})
			It("Should return a successful status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should have a summary with success message for region us01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to us01 Infrastructure API endpoint."))
			})
		})
		Context("If multiple regions provided", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/ProxyDetect": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Status:  tasks.Success,
						Payload: []string{"us01", "eu01"},
					},
				}
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Status:     "404 Not Found",
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
					}, nil
				}
			})
			It("Should return a successful status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should have checked 2 region endpoints", func() {
				payload, _ := result.Payload.(map[string]string)
				Expect(len(payload)).To(Equal(2))
			})
			It("Should have a summary with success message for region us01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to us01 Infrastructure API endpoint."))
			})
			It("Should have a summary with success message for region eu01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to eu01 Infrastructure API endpoint."))
			})
		})

		Context("If one region was provided", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/ProxyDetect": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Status:  tasks.Success,
						Payload: []string{"eu01"},
					},
				}
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Status:     "404 Not Found",
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
					}, nil
				}
			})
			It("Should return a successful status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should have checked 2 region endpoints", func() {
				payload, _ := result.Payload.(map[string]string)
				Expect(len(payload)).To(Equal(1))
			})
			It("Should have a summary with success message for region us01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to eu01 Infrastructure API endpoint."))
			})
			It("Should have a summary with success message for region eu01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to eu01 Infrastructure API endpoint."))
			})
		})

		Context("If expected status code", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/ProxyDetect": tasks.Result{
						Status: tasks.Success,
					},
					"Base/Config/RegionDetect": tasks.Result{
						Status:  tasks.Success,
						Payload: []string{"us01"},
					},
				}
				p.httpGetter = func(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
					return &http.Response{
						StatusCode: 404,
						Status:     "404 Not Found",
						Body:       ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
					}, nil
				}
			})
			It("Should return a successful status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should have a summary with success message for region us01", func() {
				Expect(result.Summary).To(ContainSubstring("Successfully connected to us01 Infrastructure API endpoint."))
			})
		})
	})
})
