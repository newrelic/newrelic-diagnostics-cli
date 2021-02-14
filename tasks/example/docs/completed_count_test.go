package docs

// This is an example task test file referenced in /docs/unit-testing.md

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
)

func TestBaseLogCount(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Base/Log/Count test suite")
}

var _ = Describe("Base/Log/Count", func() {
	var p BaseLogCount

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Log",
				Name:        "Count",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExaplanation := "Count log files collected."
			Expect(p.Explain()).To(Equal(expectedExaplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Base/Log/Collect"}
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

		Context("upstream dependency task failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Collect": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("There were no log files to count"))
			})
		})

		Context("upstream dependency returned unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Collect": tasks.Result{
						Status:  tasks.Success,
						Payload: "This should be of type []LogElement, but it's a string instead",
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("one or more log files found by upstream dependency", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Collect": tasks.Result{
						Status: tasks.Info,
						Payload: []log.LogElement{
							{
								FileName: "newrelic.log",
								FilePath: "/",
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("1 log file(s) collected"))
			})
		})

		Context("no log files found by upstream dependency", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Collect": tasks.Result{
						Status:  tasks.Info,
						Payload: []log.LogElement{},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No log files collected"))
			})
		})
	})
})
