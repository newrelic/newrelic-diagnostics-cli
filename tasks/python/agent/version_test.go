package agent

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logtask "github.com/newrelic/NrDiag/tasks/base/log"
)

func TestPythonAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Python/Agent test suite")
}

var _ = Describe("Python/Agent/Verson", func() {

	Describe("searchLogs", func() {
		var (
			logElement logtask.LogElement
			output     string
			err        error
		)

		JustBeforeEach(func() {
			output, err = searchLogs(logElement)
		})

		Context("When given a logElement containing a Python Agent Version", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic-python-agent.log",
					FilePath: "../../fixtures/python/root/tmp/",
				}
			})
			It("Should return the expected agent version and a nil error", func() {
				Expect(output).To(Equal("2.86.3.70"))
				Expect(err).To(BeNil())

			})
		})

		Context("When given a malformed logElement path ", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic_agent.log",
					FilePath: "../../../NonExistentDirectory/python/",
				}
			})
			It("Should return an empty agent version and an error", func() {
				Expect(output).To(Equal(""))
				Expect(err).To(Not(BeNil()))
			})
		})

		Context("When given a file with no Python Agent Version ", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic_agent.log",
					FilePath: "../../fixtures/node/",
				}
			})
			It("Should return an empty agent version and no error", func() {
				Expect(output).To(Equal(""))
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("getPythonVersionFromLog()", func() {
		var (
			incomingLogs []logtask.LogElement
			output       []LogPythonAgentVersion
		)

		JustBeforeEach(func() {
			output = getPythonVersionFromLog(incomingLogs)
		})

		Context("When given a logElement containing a Python Agent version", func() {
			BeforeEach(func() {
				incomingLogs = []logtask.LogElement{
					logtask.LogElement{
						FileName: "newrelic-python-agent.log",
						FilePath: "../../fixtures/python/root/tmp/",
					},
				}
			})
			It("Should return the expected agent version, log file location, and true", func() {
				expectedReturn := []LogPythonAgentVersion{
					LogPythonAgentVersion{
						Logfile:      "../../fixtures/python/root/tmp/newrelic-python-agent.log",
						AgentVersion: "2.86.3.70",
						MatchFound:   true,
					},
				}
				Expect(output).To(Equal(expectedReturn))
			})
		})

		Context("When given a logElement which does not contain a Python Agent Version", func() {
			BeforeEach(func() {
				incomingLogs = []logtask.LogElement{
					logtask.LogElement{
						FileName: "newrelic_agent.log",
						FilePath: "../../fixtures/node/",
					},
				}
			})
			It("Should return an empty Python agent version, the log file location, and false", func() {
				expectedReturn := []LogPythonAgentVersion{
					LogPythonAgentVersion{
						Logfile:      "../../fixtures/node/newrelic_agent.log",
						AgentVersion: "",
						MatchFound:   false,
					},
				}
				Expect(output).To(Equal(expectedReturn))
			})
		})

		Context("When given a logElement with a malformed filepath", func() {
			BeforeEach(func() {
				incomingLogs = []logtask.LogElement{
					logtask.LogElement{
						FileName: "newrelic_agent.BADFILE.log",
						FilePath: "../../fixtures/python/",
					},
				}
			})
			It("Should return an empty agent version, log file location, and false", func() {
				expectedReturn := []LogPythonAgentVersion{
					LogPythonAgentVersion{
						Logfile:      "../../fixtures/python/newrelic_agent.BADFILE.log",
						AgentVersion: "",
						MatchFound:   false,
					},
				}
				Expect(output).To(Equal(expectedReturn))
			})

		})
	})

})
