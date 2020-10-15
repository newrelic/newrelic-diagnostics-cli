package config

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

var _ = Describe("Base/Config/AppName", func() {

	var p BaseConfigAppName

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Config",
				Name:        "AppName",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Check for default application names in New Relic agent configuration."))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return a slice with dependencies", func() {
			expectedDependencies := []string{
				"Base/Config/Validate",
			}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("Execute()", func() {
		Context("If upstream is not successful", func() {

			It("Should return expected result", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Payload: []ValidateElement{},
						Status:  tasks.None,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "Task did not meet requirements necessary to run: no validated config files to check",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns an unexpected type", func() {

			It("Should return expected result", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Payload: []string{"test", "another one"},
						Status:  tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.Error,
					Summary: "Task did not meet requirements necessary to run: type assertion failure",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns a config file containing a default appname", func() {
			It("Should return a payload with a formatted message and Warning status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Payload: []ValidateElement{
							{
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/nrdiag/fixtures/java/newrelic/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.appname",
									Path:     "",
									RawValue: "My Application (Staging)",
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status: tasks.Warning,
					Summary: "One or more of your applications is using a default appname: " +
						"\n\t\"My Application (Staging)\" as specified in /nrdiag/fixtures/java/newrelic/newrelic.yml " +
						"\nMultiple applications with the same default appname will all report to the same source. " +
						"You may want to consider changing to a unique appname. Note that this will cause the application to report to " +
						"a new heading in the New Relic user interface, with a total discontinuity of data. If you are overriding the " +
						"default appname with environment variables, you can ignore this warning.\n--",
					URL: "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns a config file containing a customized appname", func() {
			It("Should return a payload with a formatted message and Warning status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Payload: []ValidateElement{
							{
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/nrdiag/fixtures/java/newrelic/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.appname",
									Path:     "",
									RawValue: "Custom AppName",
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status: tasks.Success,
					Payload: []AppNameInfo{
						{
							Name:     "Custom AppName",
							FilePath: "/nrdiag/fixtures/java/newrelic/newrelic.yml",
						},
					},
					Summary: "1 unique application name(s) found.",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("", func() {
			It("Should return a payload with a formatted message and Warning status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Payload: []ValidateElement{
							{
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/nrdiag/fixtures/java/newrelic/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.appname",
									Path:     "",
									RawValue: "",
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.Warning,
					Summary: "No New Relic app names were found. Please ensure an app name is set in your New Relic agent configuration.",
					URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
	})
})
