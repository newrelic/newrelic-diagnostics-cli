package template

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func TestExampleTemplate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Example/Template/* test suite")
}

var _ = Describe("Example/Template/FullTask", func() {

	var p ExampleTemplateFullTask

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Example",
				Subcategory: "Template",
				Name:        "FullTask",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Explanatory help text displayed for this task"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Base/Config/Validate"}))
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

		Context("when upstream results are empty", func() {

			// BeforeEach establishes context in which these tests will run
			// In this case our context will be empty upstream results
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{},
					},
				}
			})

			It("should return an expected none result", func() {
				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "There were no config files from which to pull the log level",
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

		Context("when upstream results type assertion fails", func() {

			// BeforeEach establishes context in which these tests will run
			// In this case our context will be upstream results payload not
			// containing expected type
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Warning,
						Payload: "I should be a slice of validate elements, but I'm not",
					},
				}
			})

			It("should return a result with None Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return a result the expected Summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("when expected key does not appear in upstream config elements", func() {
			// BeforeEach establishes context in which these tests will run
			// In this case our context will be upstream results payload not
			// containing expected type
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "license_key",
									RawValue: "Schnauzer12",
								},
							},
						},
					},
				}
			})

			It("should return a result with Failure Status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return a result with the expected Summary", func() {
				Expect(result.Summary).To(Equal("Config file doesn't contain log_level"))
			})
		})

		Context("when log level is set to finest in upstream results", func() {
			// BeforeEach establishes context in which these tests will run
			// In this case our context will be upstream results payload not
			// containing expected type
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "log_level",
									RawValue: "finest",
								},
							},
						},
					},
				}
			})

			It("should return a result with Success Status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return a result with the expected Summary", func() {
				Expect(result.Summary).To(Equal("Log level is finest"))
			})
		})

		Context("when log level is set to info in upstream results", func() {
			// BeforeEach establishes context in which these tests will run
			// In this case our context will be upstream results payload not
			// containing expected type
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "log_level",
									RawValue: "info",
								},
							},
						},
					},
				}
			})

			It("should return a result with Warning Status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return a result with the expected Summary", func() {
				Expect(result.Summary).To(Equal("Log level is info, you may want to consider updating the log level to finest before uploading logs to support"))
			})
		})

		Context("When log level is set to non-info, non-finest value in upstream results", func() {
			// BeforeEach establishes context in which these tests will run
			// In this case our context will be upstream results payload not
			// containing expected type
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "log_level",
									RawValue: "dachshund",
								},
							},
						},
					},
				}
			})

			It("should return a result with Error Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return a result with the expected Summary", func() {
				Expect(result.Summary).To(Equal("We were unable to determine the log level"))
			})
		})

	})
})
