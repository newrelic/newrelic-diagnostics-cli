package agent

import (
	"errors"
	"runtime"
	"sort"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/newrelic/newrelic-diagnostics-cli/suites"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestBaseAgentEOL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Agent/EOL test suite")
}

var _ = Describe("Base/Agent/EOL", func() {
	format.TruncatedDiff = false
	var p BaseAgentEOL

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Agent",
				Name:        "EOL",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Detect end of life (EOL) New Relic agents"))
		})
	})

	Describe("Dependencies()", func() {
		defaultDependencies := []string{
			"Node/Agent/Version",
			"Java/Agent/Version",
			"Python/Agent/Version",
			"Ruby/Agent/Version",
			"PHP/Agent/Version",
		}

		if runtime.GOOS == "windows" {
			defaultDependencies = append(defaultDependencies, "DotNet/Agent/Version")
		}
		Context("When one agent suite is selected", func() {
			JustBeforeEach(func() {
				p.suiteManager = &suites.SuiteManager{
					SelectedSuites: []suites.Suite{
						{Identifier: "node"},
					},
				}
			})
			It("Should expected slice of dependencies", func() {
				Expect(p.Dependencies()).To(Equal([]string{"Node/Agent/Version"}))
			})
		})
		Context("When multiple agent suite is selected", func() {
			JustBeforeEach(func() {
				p.suiteManager = &suites.SuiteManager{
					SelectedSuites: []suites.Suite{
						{Identifier: "node"},
						{Identifier: "php"},
					},
				}
			})
			It("Should expected slice of dependencies", func() {
				Expect(p.Dependencies()).To(Equal([]string{"Node/Agent/Version", "PHP/Agent/Version"}))
			})
		})
		Context("When then 'all' suite is selected", func() {
			JustBeforeEach(func() {
				p.suiteManager = &suites.SuiteManager{
					SelectedSuites: []suites.Suite{
						{Identifier: "all"},
						{Identifier: "php"},
					},
				}
			})
			It("Should expected slice of dependencies", func() {
				Expect(p.Dependencies()).To(Equal(defaultDependencies))
			})
		})
		Context("When no suite is selected", func() {
			JustBeforeEach(func() {
				p.suiteManager = &suites.SuiteManager{}
			})
			It("Should expected slice of dependencies", func() {
				Expect(p.Dependencies()).To(Equal(defaultDependencies))
			})
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

		Context("When agent not installed and version is not available", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"Java/Agent/Version": tasks.Result{
						Status: tasks.Failure,
					},
					"Python/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"PHP/Agent/Version": tasks.Result{
						Status: tasks.Error,
					},
					"Dotnet/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}

			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No New Relic agent versions detected. This task did not run"))
			})
		})

		Context("When all agents found are supported", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "2.13.2",
					},
					"Java/Agent/Version": tasks.Result{
						Status:  tasks.Success,
						Payload: "3.6.2",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.0.0",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "4.0",
					},
					"PHP/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 5,
							Minor: 0,
							Build: 1,
							Patch: 212,
						},
					},
					"Dotnet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.0",
					},
				}

			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected success result summary", func() {
				expectedSummary := []string{
					"We detected 6 New Relic agent(s) running on your system:",
					"6 New Relic agent(s) whose version is within the scope of support",
				}
				Expect(result.Summary).To(Equal(strings.Join(expectedSummary, "\n")))
			})
		})

		Context("When there in an unexpected payload type from upstream dependencies", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: []string{"I should not be in a slice"},
					},
					"Java/Agent/Version": tasks.Result{
						Status:  tasks.Success,
						Payload: "3.6.2",
					},
				}

			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected success result summary", func() {
				expectedSummary := []string{
					"We detected 2 New Relic agent(s) running on your system:",
					"1 New Relic agent(s) whose version is within the scope of support",
					"1 New Relic agent(s) whose EOL status could not be determined:",
					"\tNode agent ",
				}
				Expect(result.Summary).To(Equal(strings.Join(expectedSummary, "\n")))
			})
		})

		Context("When agent not installed and version is not available", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"Java/Agent/Version": tasks.Result{
						Status: tasks.Failure,
					},
					"Python/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
					"PHP/Agent/Version": tasks.Result{
						Status: tasks.Error,
					},
					"Dotnet/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}

			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No New Relic agent versions detected. This task did not run"))
			})
		})

		Context("When there is an error parsing an agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "___#_3_#3",
					},
					"Java/Agent/Version": tasks.Result{
						Status:  tasks.Success,
						Payload: "3.6.2",
					},
				}

			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected success result summary", func() {
				expectedSummary := []string{
					"We detected 2 New Relic agent(s) running on your system:",
					"1 New Relic agent(s) whose version is within the scope of support",
					"1 New Relic agent(s) whose EOL status could not be determined:",
					"\tNode agent ___#_3_#3",
				}
				Expect(result.Summary).To(Equal(strings.Join(expectedSummary, "\n")))
			})
		})

		Context("When some agents are EOL versions", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "1.0",
					},
					"Java/Agent/Version": tasks.Result{
						Status:  tasks.Success,
						Payload: "3.6.2",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "1.0.5",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "4.0",
					},
					"PHP/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 5,
							Minor: 0,
							Build: 1,
							Patch: 212,
						},
					},
					"Dotnet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.0",
					},
				}

			})

			It("should return a success status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected success result summary", func() {
				expectedSummary := []string{
					"We detected 6 New Relic agent(s) running on your system:",
					"4 New Relic agent(s) whose version is within the scope of support",
					"2 New Relic agent(s) whose version has reached EOL:",
					"\tNode agent 1.0",
					"\tPython agent 1.0.5",
				}

				sort.Strings(expectedSummary)
				summarySplit := strings.Split(result.Summary, "\n")
				sort.Strings(summarySplit)

				Expect(summarySplit).To(Equal(expectedSummary))
			})
		})

		Context("When upstream returns a slice of vers", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{3, 2, 3, 2},
							tasks.Ver{9, 8, 7, 4},
						},
					},
				}

			})

			It("should return a failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected failure result summary", func() {
				expectedSummary := []string{
					"We detected 2 New Relic agent(s) running on your system:",
					"1 New Relic agent(s) whose version has reached EOL:",
					"\tRuby agent 3.2.3.2",
					"1 New Relic agent(s) whose version is within the scope of support",
				}
				sort.Strings(expectedSummary)
				summarySplit := strings.Split(result.Summary, "\n")
				sort.Strings(summarySplit)

				Expect(summarySplit).To(Equal(expectedSummary))
			})
		})

		Context("When upstream returns an empty slice of vers", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: []tasks.Ver{},
					},
				}

			})

			It("should return a error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected Error result summary", func() {
				expectedSummary := []string{
					"\tRuby agent ",
					"1 New Relic agent(s) whose EOL status could not be determined:",
					"We detected 1 New Relic agent(s) running on your system:"}

				sort.Strings(expectedSummary)
				summarySplit := strings.Split(result.Summary, "\n")
				sort.Strings(summarySplit)

				Expect(summarySplit).To(Equal(expectedSummary))
			})
		})

	})

	Describe("isItEOL()", func() {
		var (
			//input
			version   string
			agentName string
			//output
			isItUnsupported bool
			err             error
		)
		JustBeforeEach(func() {
			isItUnsupported, err = isItEOL(version, agentName)
		})
		Context("With supported version 5.10.0", func() {
			BeforeEach(func() {
				version = "5.10.0"
				agentName = "Node"
			})
			It("Should return supported", func() {
				Expect(isItUnsupported).To(BeFalse())
				Expect(err).To(BeNil())
			})
		})
		Context("With unsupported version 4.23.4.113", func() {
			BeforeEach(func() {
				version = "4.23.4.113"
				agentName = "PHP"
			})
			It("Should return supported", func() {
				Expect(isItUnsupported).To(BeTrue())
				Expect(err).To(BeNil())
			})
		})
		Context("With unparseable string version", func() {
			BeforeEach(func() {
				version = "llama"
				agentName = "DotNet"
			})
			It("Should return supported", func() {
				Expect(isItUnsupported).To(BeFalse())
				Expect(err).To(Equal(errors.New("unable to parse version: " + version)))
			})
		})
	})
	Describe("genSummaryStatus()", func() {
		var (
			//input
			successes []agentVersion
			errors    []agentVersion
			failures  []agentVersion

			//output
			status  tasks.Status
			summary string
		)

		JustBeforeEach(func() {
			status, summary = genSummaryStatus(successes, errors, failures)
		})

		Context("When provided 0 total count of successes, errors and failures", func() {
			BeforeEach(func() {
				successes = []agentVersion{}
				errors = []agentVersion{}
				failures = []agentVersion{}
			})

			It("Should return error status and expected summary", func() {
				Expect(status).To(Equal(tasks.None))
				Expect(summary).To(Equal("No New Relic agent versions detected. This task did not run"))
			})
		})

		Context("When provided successes > 0", func() {
			BeforeEach(func() {

				successes = []agentVersion{
					agentVersion{
						name:    "Node",
						version: "5.10.0",
					},
					agentVersion{
						name:    "Java",
						version: "4.11.0",
					},
				}
				errors = []agentVersion{}
				failures = []agentVersion{}
			})

			It("Should return success status and expected summary", func() {
				lines := []string{
					"We detected 2 New Relic agent(s) running on your system:",
					"2 New Relic agent(s) whose version is within the scope of support",
				}
				Expect(status).To(Equal(tasks.Success))
				Expect(summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

		Context("When provided failures > 0", func() {
			BeforeEach(func() {

				successes = []agentVersion{}
				errors = []agentVersion{}
				failures = []agentVersion{
					agentVersion{
						name:    "Python",
						version: "2.40.0.34",
					},
					agentVersion{
						name:    "Java",
						version: "2.21.4",
					},
				}
			})

			It("Should return failure status and expected summary", func() {
				lines := []string{
					"We detected 2 New Relic agent(s) running on your system:",
					"2 New Relic agent(s) whose version has reached EOL:",
					"	Python agent 2.40.0.34",
					"	Java agent 2.21.4",
				}
				sort.Strings(lines)
				summarySplit := strings.Split(summary, "\n")
				sort.Strings(summarySplit)

				Expect(status).To(Equal(tasks.Failure))
				Expect(summarySplit).To(Equal(lines))
			})
		})

		Context("When provided with errors", func() {
			BeforeEach(func() {
				successes = []agentVersion{}
				errors = []agentVersion{
					agentVersion{
						name:    "Python",
						version: "",
					},
				}
				failures = []agentVersion{}
			})

			It("Should return error status and expected summary", func() {
				lines := []string{
					"We detected 1 New Relic agent(s) running on your system:",
					"1 New Relic agent(s) whose EOL status could not be determined:",
					"	Python agent ",
				}
				Expect(status).To(Equal(tasks.Error))
				Expect(summary).To(Equal(strings.Join(lines, "\n")))
			})
		})

	})

})
