package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node/Env/VersionCompatibility", func() {
	var p NodeEnvVersionCompatibility

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Node",
				Subcategory: "Env",
				Name:        "VersionCompatibility",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct string", func() {
			expectedString := "Check Nodejs version compatibility with New Relic Nodejs agent"

			Expect(p.Explain()).To(Equal(expectedString))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return correct slice", func() {
			expectedDependencies := []string{"Node/Env/Version", "Node/Agent/Version"}

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

		Context("when Node.js version does not return anything", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.None,
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "6.0.0.0",
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Node is not installed"))
			})
		})

		Context("when Node.js Agent version does not return anything", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 10,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Node Agent Version not detected. This task didn't run."))
			})
		})

		Context("When Node Version returns a wrong type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "10.6.7.456",
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "6.0.0.0",
					},
				}
			})
			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected None for not meeting requirements", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("When Node Agent Version returns a wrong type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 10,
							Minor: 6,
							Patch: 7,
							Build: 456,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 6,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
				}
			})
			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected None for not meeting requirements", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("When an odd version of Node.js is used", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 11,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "6.0.0.0",
					},
				}
			})

			It("should return an expected Warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return an expected Warning for an odd version", func() {
				Expect(result.Summary).To(Equal("Your 11 Node Version is not officially supported by the Node Agent because odd versions are considered unstable/experimental releases"))
			})
		})

		Context("When an invalid version of Node.js is used", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 8,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "6.0.0.0",
					},
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Warning for an odd version", func() {
				Expect(result.Summary).To(Equal("Your 8.0.0.0 Node version is not in the list of supported versions by the Node Agent. Please review our documentation on version requirements"))
			})
		})

		Context("When VersionIsCompatible returns an error", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 10,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "potato",
					},
				}
			})

			It("should return an expected Error result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected Error for the string being passed in", func() {
				Expect(result.Summary).To(Equal("There was an issue when checking for Node.js Version compatibility: Unable to parse version: potato"))
			})
		})

		Context("When VersionIsCompatible returns false", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 12,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "4.6.0.0",
					},
				}
			})

			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Failure for the incompatible versions", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 12.0.0.0, is not compatible with New Relic's Node.js agent"))
			})
		})

		Context("When an even valid version of Node.js is used", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 12,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "6.0.0.0",
					},
				}
			})

			It("should return an expected success for an odd version", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 12.0.0.0, is compatible with New Relic's Node.js agent"))
			})
		})

		Context("When version 10 of Node.js is used with a valid Agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 10,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "7.0.0.0",
					},
				}
			})

			It("should return an expected success for a compatible version", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected successful message", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 10.0.0.0, is compatible with New Relic's Node.js agent"))
			})
		})

		Context("When version 10 of Node.js is used with an incompatible Agent version", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Node/Env/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 10,
							Minor: 0,
							Patch: 0,
							Build: 0,
						},
					},
					"Node/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "8.0.0.0",
					},
				}
			})

			It("should return an expected failure for a compatible version", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected failure message", func() {
				Expect(result.Summary).To(Equal("Your current Node.js version, 10.0.0.0, is not compatible with New Relic's Node.js agent"))
			})
		})
	})

})
