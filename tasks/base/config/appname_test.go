package config

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
				"Base/Env/CollectEnvVars",
				"Base/Env/CollectSysProps",
			}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("Execute()", func() {
		Context("If upstream is not successful", func() {

			It("Should return expected result", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": {
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
					"Base/Config/Validate": {
						Payload: []string{"test", "another one"},
						Status:  tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.Error,
					Summary: tasks.AssertionErrorSummary,
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns a config file containing a default appname", func() {
			It("Should return a payload with a formatted message and Warning status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": {
						Payload: []ValidateElement{
							{
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/nrdiag/fixtures/java/newrelic/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.appname",
									Path:     "",
									RawValue: "My Application",
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.Warning,
					Summary: "One or more of your applications is using a default appname: \n\t\"My Application\" as specified in /nrdiag/fixtures/java/newrelic/newrelic.yml \nMultiple applications with the same default appname will all report to the same source. Consider changing to a unique appname and review the recommended documentation",
					URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns a config file containing a customized appname", func() {
			It("Should return a payload with a formatted message and success status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Env/CollectEnvVars": {
						Status: tasks.Info,
						Payload: map[string]string{
							"NEW_RELIC_LICENSE_KEY": "my-license-key",
						},
					},
					"Base/Env/CollectSysProps": {
						Status: tasks.Info,
						Payload: []tasks.ProcIDSysProps{
							{
								ProcID: 40149,
								SysPropsKeyToVal: map[string]string{
									"-Dnewrelic.config.appserver_dispatcher": "Apache Tomcat",
								},
							},
						},
					},
					"Base/Config/Validate": {
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
					Summary: "1 unique application name(s) found: Custom AppName",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns a config file containing multiple appnames", func() {
			It("Should return a payload with a formatted message and success status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": {
						Payload: []ValidateElement{
							{
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/nrdiag/fixtures/java/newrelic/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: "",
									Children: []tasks.ValidateBlob{
										{
											Key:      "app_name",
											Path:     "common",
											RawValue: "Custom AppName",
										},
										{
											Key:      "app_name",
											Path:     "development",
											RawValue: "My Application (Development)",
										},
										{
											Key:      "app_name",
											Path:     "production",
											RawValue: "My Application (Production)",
										},
									},
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
					Summary: "1 unique application name(s) found: Custom AppName",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream finds the new relic env var app name", func() {
			It("Should return a payload with the env var value and ignore values found in config files", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{

					"Base/Env/CollectEnvVars": {
						Status: tasks.Info,
						Payload: map[string]string{
							"NEW_RELIC_APP_NAME": "mysolid-appname",
						},
					},
					"Base/Config/Validate": {
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
							Name:     "mysolid-appname",
							FilePath: "NEW_RELIC_APP_NAME",
						},
					},
					Summary: "A unique application name was found through the New Relic App name environment variable: mysolid-appname",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream finds a system property for app name", func() {
			It("Should return a payload with the env var value and ignore values found in config files", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{

					"Base/Env/CollectEnvVars": {
						Status: tasks.Info,
						Payload: map[string]string{
							"NEW_RELIC_LICENSE_KEY": "my-license-key",
						},
					},
					"Base/Env/CollectSysProps": {
						Status: tasks.Info,
						Payload: []tasks.ProcIDSysProps{
							{
								ProcID: 40149,
								SysPropsKeyToVal: map[string]string{
									"-Dnewrelic.config.app_name": "my-appname",
								},
							},
						},
					},
					"Base/Config/Validate": {
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
							Name:     "my-appname",
							FilePath: "-Dnewrelic.config.app_name",
						},
					},
					Summary: "An application name was found through a New Relic system property: my-appname",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("", func() {
			It("Should return a payload with a formatted message and Warning status", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Base/Config/Validate": {
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
					Summary: "No New Relic app names were found. Please ensure an app name is set in your New Relic agent configuration file or as a New Relic environment variable (NEW_RELIC_APP_NAME). Ignore this warning if you are troubleshooting for a non APM Agent.",
					URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
	})
})
