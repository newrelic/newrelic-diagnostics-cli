package env

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func mockValidDateHeader(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		Header: http.Header{
			"Date": []string{"Wed, 26 Feb 2020 10:45:17 GMT"},
		},
		Body: ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockInvalidDateHeader(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{
		Header: http.Header{},
		Body:   ioutil.NopCloser(bytes.NewReader([]byte("test body"))),
	}, nil
}

func mockUnSuccessfulRequestErr(wrapper httpHelper.RequestWrapper) (*http.Response, error) {
	return &http.Response{}, errors.New("Error! This is an error")
}
func TestInfraCheckClockSkew(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra/Env/ClockSkew test suite")
}

var _ = Describe("Infra/Env/ClockSkew", func() {

	var p InfraEnvClockSkew
	format.TruncatedDiff = false
	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Env",
				Name:        "ClockSkew",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Detect if host has clock skew from New Relic collector"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return expected dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Infra/Agent/Connect", "Base/Config/ProxyDetect"}))
		})
	})

	Describe("Execute()", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When payload returns type assertion failure", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Agent/Connect": tasks.Result{
						Payload: "Incorrect payload",
					},
				}
			})

			It("It should a return task.None and a Summary", func() {
				Expect(result.Status).To(Equal(tasks.Error))
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("When API endpoint is not able to be parsed for collector time", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": tasks.Result{
						Payload: map[string]string{
							"requestURLs": "",
						},
					},
				}
				p.httpGetter = mockInvalidDateHeader
			})

			It("It should a return task.Error and a Summary", func() {
				Expect(result.Status).To(Equal(tasks.Error))
				Expect(result.Summary).To(Equal("Unable to determine New Relic collector time"))
			})
		})

		Context("When clock is not in sync", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": tasks.Result{
						Payload: map[string]string{
							"requestURLs": "https://infra-api.newrelic.com",
						},
					},
				}
				p.httpGetter = mockValidDateHeader
				hostTime, _ := time.Parse(time.RFC1123, "Wed, 26 Feb 2020 11:45:17 GMT")
				p.checkForClockSkew = func(time.Time) (bool, int, time.Time) {
					return true, 61, hostTime.In(time.UTC)
				}

			})

			Context("in windows environments", func() {
				BeforeEach(func() {
					p.runtimeOS = "windows"
				})

				It("It should a return a failure Status and Summary", func() {
					expectedSummary := "Detected clock skew of 61 seconds between host and New Relic collector. This could lead to chart irregularities:\n\tHost time:      2020-02-26 11:45:17 +0000 UTC\n\tCollector time: 2020-02-26 10:45:17 +0000 UTC\nYour host may be affected by clock skew. Please consider using NTP to keep your host clocks in sync."
					Expect(result.Status).To(Equal(tasks.Failure))
					Expect(result.Summary).To(Equal(expectedSummary))
					Expect(result.URL).To(Equal(troubleshootingURLwindows))
				})
			})

			Context("in non-windows environments", func() {
				BeforeEach(func() {
					p.runtimeOS = "linux"
				})

				It("It should a return a failure Status and Summary", func() {
					expectedSummary := "Detected clock skew of 61 seconds between host and New Relic collector. This could lead to chart irregularities:\n\tHost time:      2020-02-26 11:45:17 +0000 UTC\n\tCollector time: 2020-02-26 10:45:17 +0000 UTC\nYour host may be affected by clock skew. Please consider using NTP to keep your host clocks in sync."
					Expect(result.Status).To(Equal(tasks.Failure))
					Expect(result.Summary).To(Equal(expectedSummary))
					Expect(result.URL).To(Equal(troubleshootingURLlinux))
				})
			})

		})
		Context("When clock is off by a couple of seconds but still in sync", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": tasks.Result{
						Payload: map[string]string{
							"requestURLs": "https://infra-api.newrelic.com",
						},
					},
				}
				p.httpGetter = mockValidDateHeader
				hostTime, _ := time.Parse(time.RFC1123, "Wed, 26 Feb 2020 10:45:19 GMT")
				p.checkForClockSkew = func(time.Time) (bool, int, time.Time) {
					return false, 2, hostTime
				}

			})

			It("It should a return a failure Status and Summary", func() {
				expectedSummary := "System clock in sync with New Relic collector"
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})
	})

	Describe("getCollectorTime()", func() {
		var (
			apiEndpoint string
			serverTime  time.Time
			err         error
		)

		JustBeforeEach(func() {
			serverTime, err = p.getCollectorTime(apiEndpoint)
		})

		Context("getCollectorTime successfully parses endpoint", func() {

			BeforeEach(func() {
				apiEndpoint = "https://infra-api.newrelic.com"
				p.httpGetter = mockValidDateHeader
			})

			It("should return a server time and nil", func() {
				expectedTime, _ := time.Parse(time.RFC1123, "Wed, 26 Feb 2020 10:45:17 GMT")
				Expect(serverTime).To(Equal(expectedTime.In(time.UTC)))
				Expect(err).To(BeNil())
			})

		})

		Context("getCollectorTime unsuccessfully parses endpoint", func() {

			BeforeEach(func() {
				apiEndpoint = "https://infra-api.newrelic.com"
				p.httpGetter = mockInvalidDateHeader
			})

			It("should return a an error", func() {
				Expect(err).To(Equal(errResponseMissingDateHeader))
			})

		})
	})
})
