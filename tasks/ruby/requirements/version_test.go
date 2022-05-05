package requirements

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRequirements(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ruby/Requirements Suite")
}

var _ = Describe("Ruby/Requirements/Version", func() {
	var p RubyRequirementsVersion //instance of our task struct to be used in tests

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Ruby",
				Subcategory: "Requirements",
				Name:        "Version",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Check Ruby version compatibility with New Relic Ruby agent"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Ruby/Env/Version",
				"Ruby/Agent/Version"}
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
					"Ruby/Env/Version": tasks.Result{
						Status: tasks.Error,
					},
					"Ruby/Agent/Version": tasks.Result{
						Status:  tasks.Success,
						Payload: "6.11.0.365",
					},
				}
			})

			It("Should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Ruby version not detected. This task didn't run."))
			})
		})
		Context("when upstream dependency ruby agent version does not return successful status", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby 2.4.0p0 (2016-12-24 revision 57164) [x86_64-linux]", //changed to just 2.4 and still worked with the test --> ruby 2.4.0p0 (2016-12-24 revision 57164) [x86_64-linux]
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}
			})
			It("Should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Ruby Agent version not detected. This task didn't run."))
			})
		})

		Context("when upstream dependency ruby env version is not compatible with ruby agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby 1.8.6p0 (2016-12-24 revision 57164) [x86_64-linux]",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{3, 9, 9, 275},
						},
					},
				}
			})
			It("Should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none failure summary", func() {
				Expect(result.Summary).To(Equal("Your ruby 1.8.6p0 (2016-12-24 revision 57164) [x86_64-linux] Ruby version is not in the list of supported versions by the Ruby Agent. Please review our documentation on version requirements"))
			})
		})

		Context("when upstream dependency ruby parses and encounters an error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby ",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{3, 9, 9, 275},
						},
					},
				}
			})
			It("Should return an expected error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none error summary", func() {
				Expect(result.Summary).To(Equal("While parsing the Ruby Version, we encountered an error: no found result for Ruby Version when parsing for payload ruby "))
			})
		})

		Context("when upstream dependency ruby agent version is compatible", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby 2.4.0p0 (2016-12-24 revision 57164) [x86_64-linux]",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{4, 8, 0, 2},
							tasks.Ver{4, 8, 0, 3},
						},
					},
				}
			})
			It("Should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none success summary", func() {
				Expect(result.Summary).To(Equal("Compatible Version detected: 4.8.0.2\nCompatible Version detected: 4.8.0.3\n"))
			})
		})

		Context("when upstream dependency ruby agent is incompatible", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby 2.4.0p0 (2016-12-24 revision 57164) [x86_64-linux]",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{3, 9, 9, 275},
							tasks.Ver{3, 9, 6, 257},
						},
					},
				}
			})
			It("Should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none failure summary", func() {
				Expect(result.Summary).To(Equal("Incompatible Version detected: 3.9.9.275\nIncompatible Version detected: 3.9.6.257\n"))
			})
		})

		Context("when upstream dependency ruby agents are compatible and incompatible ", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Ruby/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "ruby 1.9.2p0 (2016-12-24 revision 57164) [x86_64-linux]",
					},
					"Ruby/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: []tasks.Ver{
							tasks.Ver{3, 9, 5, 275},
							tasks.Ver{3, 9, 6, 257},
						},
					},
				}
			})
			It("Should return an expected warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return an expected none warning summary", func() {
				Expect(result.Summary).To(Equal("Incompatible Version detected: 3.9.5.275\nCompatible Version detected: 3.9.6.257\n"))
			})
		})

	})

})
