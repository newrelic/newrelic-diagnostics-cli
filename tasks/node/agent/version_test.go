package agent

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logtask "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
)

func TestNodeAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Node/Agent test suite")
}

var _ = Describe("Node/Agent/Version", func() {

	Describe("searchLogs", func() {
		var (
			logElement logtask.LogElement
			output     string
			err        error
		)

		JustBeforeEach(func() {
			output, err = searchLogs(logElement)
		})

		Context("When given a logElement containing a Node Agent Version", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic_agent.log",
					FilePath: "../../fixtures/node/",
				}
			})
			It("Should return the expected agent version and a nil error", func() {
				Expect(output).To(Equal("1.38.2"))
				Expect(err).To(BeNil())

			})
		})

		Context("When given a malformed logElement path ", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic_agent.log",
					FilePath: "../../../NonExistentDirectory/node/",
				}
			})
			It("Should return an empty agent version and an error", func() {
				Expect(output).To(Equal(""))
				Expect(err).To(Not(BeNil()))
			})
		})

		Context("When given a file with no Node Agent Version ", func() {
			BeforeEach(func() {
				logElement = logtask.LogElement{
					FileName: "newrelic-python-agent.log",
					FilePath: "../../fixtures/python/root/tmp/",
				}
			})
			It("Should return an empty agent version and no error", func() {
				Expect(output).To(Equal(""))
				Expect(err).To(BeNil())
			})
		})

	})

	Describe("getNodeVerFromLog()", func() {
		var (
			incomingLogs []logtask.LogElement
			output       []logNodeAgentVersion
		)

		JustBeforeEach(func() {
			output = getNodeVerFromLog(incomingLogs)
		})

		Context("When given a logElement containing a Node Agent version", func() {
			BeforeEach(func() {
				incomingLogs = []logtask.LogElement{
					logtask.LogElement{
						FileName: "newrelic_agent.log",
						FilePath: "../../fixtures/node/",
					},
				}
			})
			It("Should return the expected agent version, log file location, and true", func() {
				expectedReturn := []logNodeAgentVersion{
					logNodeAgentVersion{
						Logfile:      "../../fixtures/node/newrelic_agent.log",
						AgentVersion: "1.38.2",
						MatchFound:   true,
					},
				}
				Expect(output).To(Equal(expectedReturn))
			})
		})

		Context("When given a logElement which does not contain a Node Agent Version", func() {
			BeforeEach(func() {
				incomingLogs = []logtask.LogElement{
					logtask.LogElement{
						FileName: "newrelic-python-agent.log",
						FilePath: "../../fixtures/python/root/tmp/",
					},
				}
			})
			It("Should return an empty Node agent version, the log file location, and false", func() {
				expectedReturn := []logNodeAgentVersion{
					logNodeAgentVersion{
						Logfile:      "../../fixtures/python/root/tmp/newrelic-python-agent.log",
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
						FilePath: "../../fixtures/node/",
					},
				}
			})
			It("Should return an empty agent version, log file location, and false", func() {
				expectedReturn := []logNodeAgentVersion{
					logNodeAgentVersion{
						Logfile:      "../../fixtures/node/newrelic_agent.BADFILE.log",
						AgentVersion: "",
						MatchFound:   false,
					},
				}
				Expect(output).To(Equal(expectedReturn))
			})

		})
	})

})
