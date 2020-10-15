package config

import (
	"sort"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

type byFileName []IntegrationMatchError

func (p byFileName) Len() int {
	return len(p)
}

func (p byFileName) Less(i, j int) bool {
	return p[i].IntegrationFile.Config.FileName < p[j].IntegrationFile.Config.FileName
}

func (p byFileName) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

var _ = Describe("Infra/Config/IntegrationMatch", func() {

	var p InfraConfigIntegrationsMatch

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Config",
				Name:        "IntegrationsMatch",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Validate New Relic Infrastructure on-host integration configuration and definition file pairs"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return a slice with dependencies", func() {
			expectedDependencies := []string{"Infra/Config/IntegrationsValidate"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

	Describe("Execute()", func() {
		Context("If upstream is not successful", func() {

			It("Should return expected result", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Status: tasks.Failure,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "Task did not meet requirements necessary to run: no validated integrations",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns an unexpected type", func() {

			It("Should return expected result", func() {
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []string{"test", "another one"},
						Status:  tasks.Success,
					},
				}

				expectedResult := tasks.Result{
					Status:  tasks.None,
					Summary: "Task did not meet requirements necessary to run: type assertion failure",
				}

				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})

		Context("If upstream returns no matching pairs", func() {
			It("Should return a payload with no matching pairs with a status of failure and the expected URL", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "banana-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.banana",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "kafka-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.kafka",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedErrors := []IntegrationMatchError{
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "banana-config.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "integration_name",
										Path:     "",
										RawValue: "com.banana.banana",
										Children: nil,
									},
								},
							},
						},
						Reason: "Configuration file '/etc/newrelic-infra/integrations.d/banana-config.yml' does not have matching Definition file",
					},
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "kafka-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "name",
										Path:     "",
										RawValue: "com.banana.kafka",
										Children: nil,
									},
								},
							},
						},
						Reason: "Definition file '/var/db/newrelic-infra/custom-integrations/kafka-definition.yml' does not have matching Configuration file",
					},
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "apache-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
						Reason: "Definition file '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' does not have matching Configuration file",
					},
				}

				expectedResult := tasks.Result{
					Status: tasks.Failure,
					Summary: "No matching integration files found" +
						"\nConfiguration file '/etc/newrelic-infra/integrations.d/banana-config.yml' does not have matching Definition file" +
						"\nDefinition file '/var/db/newrelic-infra/custom-integrations/kafka-definition.yml' does not have matching Configuration file" +
						"\nDefinition file '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' does not have matching Configuration file",
					URL: "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				}
				result := p.Execute(executeOptions, executeUpstream)

				resultPayload := result.Payload.(MatchedIntegrationFiles)
				resultErrors := resultPayload.Errors
				sort.Sort(byFileName(resultErrors))

				sort.Sort(byFileName(expectedErrors))

				Expect(resultErrors).To(Equal(expectedErrors))
				Expect(result.Status).To(Equal(expectedResult.Status))
				Expect(result.URL).To(Equal(expectedResult.URL))
			})
		})

		Context("If upstream returns some matching pairs and some unmatched files", func() {
			It("Should return a payload with some matching pairs and unmatched with a status of failure", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "banana-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.banana",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, { //extra matching pair to handle the case where a config file is found first
								Config: config.ConfigElement{
									FileName: "kafka-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.kafka",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "kafka-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.kafka",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "banana-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "name",
										Path:     "",
										RawValue: "com.banana.banana",
										Children: nil,
									},
								},
							},
						},
						Reason: "Definition file '/var/db/newrelic-infra/custom-integrations/banana-definition.yml' does not have matching Configuration file",
					},
				}

				expectedMatchedPairs["apache"] = &IntegrationFilePair{
					Configuration: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-config.yml",
							FilePath: "/etc/newrelic-infra/integrations.d/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "integration_name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
					Definition: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-definition.yml",
							FilePath: "/var/db/newrelic-infra/custom-integrations/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
				}

				expectedMatchedPairs["kafka"] = &IntegrationFilePair{
					Configuration: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "kafka-config.yml",
							FilePath: "/etc/newrelic-infra/integrations.d/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "integration_name",
									Path:     "",
									RawValue: "com.banana.kafka",
									Children: nil,
								},
							},
						},
					},
					Definition: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "kafka-definition.yml",
							FilePath: "/var/db/newrelic-infra/custom-integrations/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "name",
									Path:     "",
									RawValue: "com.banana.kafka",
									Children: nil,
								},
							},
						},
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status: tasks.Failure,
					Summary: "Found matching integration files with some errors: " +
						"\nDefinition file '/var/db/newrelic-infra/custom-integrations/banana-definition.yml' does not have matching Configuration file",
					Payload: expectedPayload,
					URL:     "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})

		Context("If upstream returns all matching pairs on a windows system", func() {
			It("Should return a payload with only matching pairs and a status of success", func() {
				p.runtimeOS = "windows"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "C:\\Program Files\\New Relic\\newrelic-infra\\custom-integrations\\",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "C:\\Program Files\\New Relic\\newrelic-infra\\integrations.d\\",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{}

				expectedMatchedPairs["apache"] = &IntegrationFilePair{
					Configuration: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-config.yml",
							FilePath: "C:\\Program Files\\New Relic\\newrelic-infra\\integrations.d\\",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "integration_name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
					Definition: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-definition.yml",
							FilePath: "C:\\Program Files\\New Relic\\newrelic-infra\\custom-integrations\\",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status:  tasks.Success,
					Summary: "Found matching integration files",
					Payload: expectedPayload,
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns all matching pairs", func() {
			It("Should return a payload with only matching pairs and a status of success", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{}

				expectedMatchedPairs["apache"] = &IntegrationFilePair{
					Configuration: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-config.yml",
							FilePath: "/etc/newrelic-infra/integrations.d/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "integration_name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
					Definition: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-definition.yml",
							FilePath: "/var/db/newrelic-infra/custom-integrations/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status:  tasks.Success,
					Summary: "Found matching integration files",
					Payload: expectedPayload,
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If upstream returns all matching pair in standard integration directory", func() {
			It("Should return a payload with only matching pairs and a status of success", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/newrelic-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{}

				expectedMatchedPairs["apache"] = &IntegrationFilePair{
					Configuration: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-config.yml",
							FilePath: "/etc/newrelic-infra/integrations.d/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "integration_name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
					Definition: config.ValidateElement{
						Config: config.ConfigElement{
							FileName: "apache-definition.yml",
							FilePath: "/var/db/newrelic-infra/newrelic-integrations/",
						},
						ParsedResult: tasks.ValidateBlob{
							Key:      "",
							Path:     "",
							RawValue: nil,
							Children: []tasks.ValidateBlob{
								tasks.ValidateBlob{
									Key:      "name",
									Path:     "",
									RawValue: "com.banana.apache",
									Children: nil,
								},
							},
						},
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status:  tasks.Success,
					Summary: "Found matching integration files",
					Payload: expectedPayload,
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})

		Context("If upstream returns integration files in the wrong path", func() {
			It("Should return a payload of IntegrationMatchErrors and a status of failed", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "apache-definition.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
						Reason: "Filepath '/etc/newrelic-infra/integrations.d/' not a valid location for this Integration file 'apache-definition.yml'",
					},
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "apache-config.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "integration_name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
						Reason: "Filepath '/var/db/newrelic-infra/custom-integrations/' not a valid location for this Integration file 'apache-config.yml'",
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status: tasks.Failure,
					Summary: "No matching integration files found" +
						"\nFilepath '/etc/newrelic-infra/integrations.d/' not a valid location for this Integration file 'apache-definition.yml'" +
						"\nFilepath '/var/db/newrelic-infra/custom-integrations/' not a valid location for this Integration file 'apache-config.yml'",
					Payload: expectedPayload,
					URL:     "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})
		Context("If Integration file config pairs are found with mismatching integration names in the yaml", func() {
			It("Should return a payload of IntegrationMatchErrors and a status of failed", func() {
				p.runtimeOS = "linux"
				executeOptions := tasks.Options{}
				executeUpstream := map[string]tasks.Result{
					"Infra/Config/IntegrationsValidate": tasks.Result{
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: nil,
										},
									},
								},
							}, {
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: nil,
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "integration_name",
											Path:     "",
											RawValue: "com.banana.apache-integration",
											Children: nil,
										},
									},
								},
							},
						},
						Status: tasks.Success,
					},
				}

				expectedMatchedPairs := make(map[string]*IntegrationFilePair)
				expectedErrors := []IntegrationMatchError{
					IntegrationMatchError{
						IntegrationFile: config.ValidateElement{
							Config: config.ConfigElement{
								FileName: "apache-config.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "integration_name",
										Path:     "",
										RawValue: "com.banana.apache-integration",
										Children: nil,
									},
								},
							},
						},
						Reason: "Integration Configuration file '/etc/newrelic-infra/integrations.d/apache-config.yml' 'integration_name': 'com.banana.apache-integration', expected to match Definition file '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' 'name': 'com.banana.apache'",
					},
				}

				expectedPayload := MatchedIntegrationFiles{
					IntegrationFilePairs: expectedMatchedPairs,
					Errors:               expectedErrors,
				}

				expectedResult := tasks.Result{
					Status: tasks.Failure,
					Summary: "No matching integration files found" +
						"\nIntegration Configuration file '/etc/newrelic-infra/integrations.d/apache-config.yml' 'integration_name': 'com.banana.apache-integration', expected to match Definition file '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' 'name': 'com.banana.apache'",
					Payload: expectedPayload,
					URL:     "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				}
				Expect(p.Execute(executeOptions, executeUpstream)).To(Equal(expectedResult))
			})
		})

	})

	Context("If upstream returns a matching pair with a missing `integration_name` key in the config file", func() {
		It("Should return a payload with no matching pairs, a status of Failure, and one error", func() {
			p.runtimeOS = "linux"
			executeOptions := tasks.Options{}
			executeUpstream := map[string]tasks.Result{
				"Infra/Config/IntegrationsValidate": tasks.Result{
					Payload: []config.ValidateElement{
						{
							Config: config.ConfigElement{
								FileName: "apache-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						}, {
							Config: config.ConfigElement{
								FileName: "apache-config.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "inteRgration_name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
					},
					Status: tasks.Success,
				},
			}

			expectedResult := tasks.Result{
				Status:      tasks.Failure,
				Summary:     "No matching integration files found\nIntegration Configuration File '/etc/newrelic-infra/integrations.d/apache-config.yml' is missing key 'integration_name'",
				URL:         "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				FilesToCopy: []tasks.FileCopyEnvelope(nil),
				Payload: MatchedIntegrationFiles{
					IntegrationFilePairs: map[string]*IntegrationFilePair{},
					Errors: []IntegrationMatchError{
						IntegrationMatchError{
							IntegrationFile: config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "apache-config.yml",
									FilePath: "/etc/newrelic-infra/integrations.d/",
								},
								Status: 0,
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: interface{}(nil),
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "inteRgration_name",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: []tasks.ValidateBlob(nil),
										},
									},
								},
							},
							Reason: "Integration Configuration File '/etc/newrelic-infra/integrations.d/apache-config.yml' is missing key 'integration_name'",
						},
					},
				},
			}

			resultski := p.Execute(executeOptions, executeUpstream)
			Expect(resultski).To(Equal(expectedResult))
		})
	})

	Context("If upstream returns a matching pair with a missing `name` key in the definition file", func() {
		It("Should return a payload with no matching pairs, a status of Failure, and one error", func() {
			p.runtimeOS = "linux"
			executeOptions := tasks.Options{}
			executeUpstream := map[string]tasks.Result{
				"Infra/Config/IntegrationsValidate": tasks.Result{
					Payload: []config.ValidateElement{
						{
							Config: config.ConfigElement{
								FileName: "apache-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "naRme",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						}, {
							Config: config.ConfigElement{
								FileName: "apache-config.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "integration_name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
					},
					Status: tasks.Success,
				},
			}

			expectedResult := tasks.Result{
				Status:      tasks.Failure,
				Summary:     "No matching integration files found\nIntegration Definition File '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' is missing key 'name'",
				URL:         "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
				FilesToCopy: []tasks.FileCopyEnvelope(nil),
				Payload: MatchedIntegrationFiles{
					IntegrationFilePairs: map[string]*IntegrationFilePair{},
					Errors: []IntegrationMatchError{
						IntegrationMatchError{
							IntegrationFile: config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "apache-definition.yml",
									FilePath: "/var/db/newrelic-infra/custom-integrations/",
								},
								Status: 0,
								ParsedResult: tasks.ValidateBlob{
									Key:      "",
									Path:     "",
									RawValue: interface{}(nil),
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob{
											Key:      "naRme",
											Path:     "",
											RawValue: "com.banana.apache",
											Children: []tasks.ValidateBlob(nil),
										},
									},
								},
							},
							Reason: "Integration Definition File '/var/db/newrelic-infra/custom-integrations/apache-definition.yml' is missing key 'name'",
						},
					},
				},
			}

			resultski := p.Execute(executeOptions, executeUpstream)
			Expect(resultski).To(Equal(expectedResult))
		})
	})

	Context("If upstream returns a matching pair with a missing `name` key in the definition file and a missing `integration_name` in the configuration file", func() {
		It("Should return a payload with no matching pairs, a status of Failure, and two errors", func() {
			p = InfraConfigIntegrationsMatch{runtimeOS: "linux"}
			executeOptions := tasks.Options{}
			executeUpstream := map[string]tasks.Result{
				"Infra/Config/IntegrationsValidate": tasks.Result{
					Payload: []config.ValidateElement{
						{
							Config: config.ConfigElement{
								FileName: "apache-definition.yml",
								FilePath: "/var/db/newrelic-infra/custom-integrations/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "naRme",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						}, {
							Config: config.ConfigElement{
								FileName: "apache-config.yml",
								FilePath: "/etc/newrelic-infra/integrations.d/",
							},
							ParsedResult: tasks.ValidateBlob{
								Key:      "",
								Path:     "",
								RawValue: nil,
								Children: []tasks.ValidateBlob{
									tasks.ValidateBlob{
										Key:      "inteRgration_name",
										Path:     "",
										RawValue: "com.banana.apache",
										Children: nil,
									},
								},
							},
						},
					},
					Status: tasks.Success,
				},
			}

			resultski := p.Execute(executeOptions, executeUpstream)
			resultPayload := resultski.Payload.(MatchedIntegrationFiles)
			/* Errors come back in variable order, so let's assert against
			   the length instead of writing another custom sort */
			Expect(len(resultPayload.Errors)).To(Equal(2))
			Expect(len(resultPayload.IntegrationFilePairs)).To(Equal(0))
		})
	})

})
