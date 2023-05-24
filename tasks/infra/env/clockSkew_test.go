package env

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
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
					"Infra/Agent/Connect": {
						Status:  tasks.Success,
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

					"Infra/Agent/Connect": {
						Status:  tasks.Failure,
						Payload: []string{},
					},
				}
				p.httpGetter = mockInvalidDateHeader
			})

			It("It should a return task.Error and a Summary", func() {
				Expect(result.Status).To(Equal(tasks.Error))
				Expect(result.Summary).To(Equal("Unable to determine New Relic collector URL from Infra/Agent/Connect task"))
			})
		})

		Context("When upstream returns tasks.None result", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": {
						Status: tasks.None,
					},
				}
				p.httpGetter = mockInvalidDateHeader
			})

			It("It should a return task.Error and a Summary", func() {
				Expect(result.Status).To(Equal(tasks.None))
				Expect(result.Summary).To(Equal("Unable to retrieve urls from Infra/Agent/Connect. This task did not run"))
			})
		})

		Context("When clock is not in sync", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": {
						Status: tasks.Success,
						Payload: []string{
							"https://infra-api.newrelic.com",
						},
					},
				}
				p.httpGetter = mockValidDateHeader
				p.checkForClockSkew = func(time.Time, time.Time) (bool, int) {
					return true, 61
				}
			})

			Context("in windows environments", func() {
				BeforeEach(func() {
					p.runtimeOS = "windows"
				})

				It("It should a return a failure Status and Summary", func() {
					expectedSummary := "Detected clock skew of 61 seconds between host and New Relic collector[.] This could lead to chart irregularities:\n\tHost time: .+\n\tCollector time: .+\nYour host may be affected by clock skew. Please consider using NTP to keep your host clocks in sync."
					Expect(result.Status).To(Equal(tasks.Failure))
					Expect(result.Summary).To(MatchRegexp(expectedSummary))
				})
			})

		})

		Context("When clock is off by a couple of seconds but still in sync", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{

					"Infra/Agent/Connect": {
						Status: tasks.Success,
						Payload: []string{
							"https://infra-api.newrelic.com",
						},
					},
				}
				p.httpGetter = mockValidDateHeader
				p.checkForClockSkew = func(time.Time, time.Time) (bool, int) {
					return false, 2
				}

			})

			It("It should a return a successful Status and Summary", func() {
				expectedSummary := "System clock in sync with New Relic collector"
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})

		Context("When infra-api is not in the collector URLs", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Agent/Connect": {
						Status: tasks.Success,
						Payload: []string{
							"https://metric-api.newrelic.com",
						},
					},
				}
				p.httpGetter = mockValidDateHeader
				p.checkForClockSkew = func(time.Time, time.Time) (bool, int) {
					return false, 2
				}

			})

			It("It should a return a successful Status and Summary", func() {
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

	Describe("checkForClockSkew()", func() {
		var (
			isClockDiffRelevant bool
			diffSeconds         int
			hostTime            time.Time
			serverTime          time.Time
		)

		JustBeforeEach(func() {
			isClockDiffRelevant, diffSeconds = checkForClockSkew(serverTime, hostTime)
		})

		Context("checkForClockSkew successfully detects skew under 60 seconds", func() {
			BeforeEach(func() {
				hostTime, _ = time.Parse(time.RFC1123, "Wed, 24 May 2023 12:00:00 GMT")
				serverTime, _ = time.Parse(time.RFC1123, "Wed, 24 May 2023 12:00:02 GMT")
			})

			It("should return false and 2", func() {
				expectedIsClockDiffRelevant := false
				expectedDiffSeconds := 2
				Expect(isClockDiffRelevant).To(Equal(expectedIsClockDiffRelevant))
				Expect(diffSeconds).To(Equal(expectedDiffSeconds))
			})
		})

		Context("checkForClockSkew successfully detects skew over 60 seconds", func() {
			BeforeEach(func() {
				hostTime, _ = time.Parse(time.RFC1123, "Wed, 24 May 2023 12:00:00 GMT")
				serverTime, _ = time.Parse(time.RFC1123, "Wed, 24 May 2023 12:01:01 GMT")
			})

			It("should return true and 61", func() {
				expectedIsClockDiffRelevant := true
				expectedDiffSeconds := 61
				Expect(isClockDiffRelevant).To(Equal(expectedIsClockDiffRelevant))
				Expect(diffSeconds).To(Equal(expectedDiffSeconds))
			})
		})

	})

})
