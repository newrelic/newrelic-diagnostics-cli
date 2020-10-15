package requirements

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

var _ = Describe("Python/Requirements/PythonVersion", func() {
	var p PythonRequirementsPythonVersion //instance of our task struct to be used in tests

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Python",
				Subcategory: "Requirements",
				Name:        "PythonVersion",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Check Python version compatibility with New Relic Python agent"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Python/Env/Version",
				"Python/Agent/Version"}
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

		Context("when upstream dependency does not return successful status", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status: tasks.Error,
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "5.2.3.131",
					},
				}
			})

			It("Should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Python version not detected. This task didn't run."))
			})
		})

		Context("when tasks.VersionIsCompatible throws an error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.8",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "Peruvian potatoes",
					},
				}
			})

			It("should return an expected summary", func() {
				fmt.Println("MY SUMMARY: ", result.Summary)
				Expect(result.Status).To(Equal(tasks.Error))
				Expect(result.Summary).To(Equal("We ran into an error while parsing your current agent version Peruvian potatoes. Unable to parse version: Peruvian potatoes"))
			})
		})

		Context("when python version is supported by the agent", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.8.1",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "5.2.3.131",
					},
				}
			})

			It("should return an expected summary", func() {
				Expect(result.Status).To(Equal(tasks.Success))
				Expect(result.Summary).To(Equal("Your Python version is supported by the Python Agent."))
			})
		})

		Context("when python version is supported but they need to upgrade their agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.8",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.4.0.95",
					},
				}
			})

			It("should return an expected summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("Your 3.8 Python version is not supported by this specific Python Agent Version. You'll have to use a different version of the Python Agent, 5.2.3.131 as the minimum, to ensure the agent works as expected."))
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic"))
			})
		})

		Context("when python version is supported but they need to downgrade their agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "3.3",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "5.10.0.138",
					},
				}
			})

			It("should return an expected summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("Your 3.3 Python version is not supported by this specific Python Agent Version. You'll have to use a different version of the Python Agent, 2.42.0.35 as the minimum, to ensure the agent works as expected."))
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic"))
			})
		})
		Context("when python version is not supported by the agent", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Python/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "2.5",
					},
					"Python/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "5.10.0.138",
					},
				}
			})

			It("should return an expected summary", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
				Expect(result.Summary).To(Equal("Your 2.5 Python version is not in the list of supported versions by the Python Agent. Please review our documentation on version requirements"))
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic"))
			})
		})

	})
})
