package config

// Tests for Infra/Config/IntegrationsCollect

import (
	"errors"
	"os"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Infra/Config/IntegrationsValidate", func() {
	var p InfraConfigIntegrationsValidate

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Config",
				Name:        "IntegrationsValidate",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Validate yml formatting of New Relic Infrastructure on-host integration configuration and definition file(s)"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Infra/Config/IntegrationsCollect"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("validateYamls()", func() {
		var (
			yamlLocations    []config.ConfigElement
			validatedYamls   []config.ValidateElement
			validationErrors []validationError
		)

		JustBeforeEach(func() {
			validatedYamls, validationErrors = p.validateYamls(yamlLocations)
		})
		Context("when unable to parse the provided yaml file", func() {
			BeforeEach(func() {
				p.fileReader = os.Open
				yamlLocations = []config.ConfigElement{
					config.ConfigElement{
						FileName: "newrelic-infra.yml",
						FilePath: "fixtures/invalid_infra_config/",
					},
				}
			})
			It("should return the expected validation error", func() {
				expectedResult := []validationError{
					validationError{
						fileLocation: "fixtures/invalid_infra_config/newrelic-infra.yml",
						errorText:    "Unable to parse yaml: yaml: line 14: did not find expected key.\nThis can mean that you either have incorrect spacing/indentation around this line or that you have a syntax error, such as a missing/invalid character",
					},
				}
				Expect(validationErrors).To(Equal(expectedResult))
			})

			It("should return an empty slice of validated yamls", func() {
				expectedResult := []config.ValidateElement{}
				Expect(validatedYamls).To(Equal(expectedResult))
			})
		})

		Context("when there is an error reading the provided yaml file", func() {
			BeforeEach(func() {
				p.fileReader = func(fileName string) (*os.File, error) {
					return nil, errors.New("read error!")
				}
				yamlLocations = []config.ConfigElement{
					config.ConfigElement{
						FileName: "config.yml",
						FilePath: "fake/path/",
					},
				}
			})
			It("should return the expected validation error", func() {
				expectedResult := []validationError{
					validationError{
						fileLocation: "fake/path/config.yml",
						errorText:    "Unable to read yaml: read error!",
					},
				}
				Expect(validationErrors).To(Equal(expectedResult))
			})

			It("should return an empty slice of validated yamls", func() {
				expectedResult := []config.ValidateElement{}
				Expect(validatedYamls).To(Equal(expectedResult))
			})
		})

		Context("when provided correct yaml", func() {
			BeforeEach(func() {
				p.fileReader = os.Open
				yamlLocations = []config.ConfigElement{
					config.ConfigElement{
						FileName: "correct.yml",
						FilePath: "fixtures/test_yml/",
					},
				}
			})
			It("should return the slice of config.ValidateElements with expected values", func() {
				expectedResult := []config.ValidateElement{
					{
						Config: config.ConfigElement{
							FileName: "correct.yml",
							FilePath: "fixtures/test_yml/",
						},
						Status: 0,
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "group1",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "key",
											Path:     "/group1",
											RawValue: "value",
											Children: nil,
										},
									},
								},
							},
						},
					},
				}
				Expect(validatedYamls).To(Equal(expectedResult))
			})

			It("should return an empty slice of validation errors yamls", func() {
				expectedResult := []validationError{}
				Expect(validationErrors).To(Equal(expectedResult))
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

		Context("upstream dependency task failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsCollect": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No On-host Integration config and definitions files were collected. Task not executed."))
			})
		})

		Context("upstream dependency task returned unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsCollect": tasks.Result{
						Status:  tasks.Success,
						Payload: "I should be a []config.ConfigElement",
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

		Context("no integration configurations were are present in upstream payload", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsCollect": tasks.Result{
						Status:  tasks.Success,
						Payload: []config.ConfigElement{},
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No On-host Integration config or definition files were found. Task not executed."))
			})
		})
		Context("one good integration configuration is present in upstream payload", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsCollect": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ConfigElement{
							config.ConfigElement{
								FileName: "correct.yml",
								FilePath: "fixtures/test_yml/",
							},
						},
					},
				}
			})
			It("should return an expected Success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected Success result summary", func() {
				Expect(result.Summary).To(Equal("Successfully validated 1 yaml file(s)"))
			})
		})

		Context("bad integration configuration is present in upstream payload", func() {
			BeforeEach(func() {
				p.fileReader = os.Open
				upstream = map[string]tasks.Result{
					"Infra/Config/IntegrationsCollect": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ConfigElement{
							config.ConfigElement{
								FileName: "newrelic-infra.yml",
								FilePath: "fixtures/invalid_infra_config/",
							},
						},
					},
				}
			})
			It("should return an expected Failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected Failure result summary", func() {
				expectedSummary := "Error validating on-host integration configuration files:\nfixtures/invalid_infra_config/newrelic-infra.yml: Unable to parse yaml: yaml: line 14: did not find expected key.\nThis can mean that you either have incorrect spacing/indentation around this line or that you have a syntax error, such as a missing/invalid character\n"
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})
	})
})
