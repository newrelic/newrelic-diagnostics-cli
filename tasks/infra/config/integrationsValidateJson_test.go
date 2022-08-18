package config

// Tests for Infra/Config/IntegrationsValidateJson

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Infra/Config/IntegrationsValidateJson", func() {
	var p InfraConfigIntegrationsValidateJson

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Config",
				Name:        "IntegrationsValidateJson",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Validate json of New Relic Infrastructure on-host integration configuration and definition file(s)"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Infra/Config/IntegrationsValidate"}
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
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: no validated integrations"))
			})
		})

		Context("upstream dependency task returned unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status:  tasks.Success,
						Payload: "I should be a []config.ValidateElement",
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

		Context("found invalid json in yml", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "redis-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								Status: 0,
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: interface{}(nil),
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "keys",
											Path:     "",
											RawValue: `{"key":}`,
											Children: []tasks.ValidateBlob(nil),
										},
									},
								},
							},
						},
					},
				}
			})

			It("should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Found invalid JSON values in following yml files:\n\t/etc/newrelic-infra/integrations.d/redis-config.yml: 'keys' field contains invalid JSON\n"))
			})
		})

		Context("found valid json in yml", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "redis-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								Status: 0,
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: interface{}(nil),
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "keys",
											Path:     "",
											RawValue: `{"key": "value"}`,
											Children: []tasks.ValidateBlob(nil),
										},
									},
								},
							},
						},
					},
				}
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected success result summary", func() {
				Expect(result.Summary).To(Equal("Found and validated 1 `*-config.yml` files."))
			})
		})

		Context("found no `*-config.yml` files to validate", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "redis.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								Status: 0,
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: interface{}(nil),
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "keys",
											Path:     "",
											RawValue: `{"true"}`,
											Children: []tasks.ValidateBlob(nil),
										},
									},
								},
							},
						},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Unable to locate *-config.yml files with known JSON fields to validate."))
			})
		})
	})

	Describe("isValidJson()", func() {
		var (
			result  bool
			rawJson string
		)
		JustBeforeEach(func() {
			result = isValidJson(rawJson)
		})
		Context("when given a string containing the raw JSON value for a given yaml/key", func() {
			rawJson = `{"keyName":"This is a valid JSON. This should pass!"}`
			It("should return true when given a valid JSON structured value", func() {
				expectedResult := true
				Expect(result).To(Equal(expectedResult))
			})
		})
	})

})
