package agent

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestPHPAgentVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Php/Agent/Verson test suite")
}

var _ = Describe("Php/Agent/Verson", func() {
	var p PHPAgentVersion //instance of our task struct to be used in tests

	//Tests below!

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})
		Context("When upstream['PHP/Config/Agent'].Payload.([]config.ValidateElement) returns  Ok == false 0", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status:  tasks.Success,
						Payload: []string{},
					},
				}
			})
			It("Should return task result status of error ", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return this task result summary  ", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})
		Context("When upstream['PHP/Config/Agent'].Payload.([]config.ValidateElement) returns  len(validations) ==  0", func() {
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status:  tasks.Success,
						Payload: []config.ValidateElement{},
					},
				}
			})
			It("Should return task result status of none ", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return this task result summary ", func() {
				Expect(result.Summary).To(Equal("There were no logs found to check for the agent version."))
			})
		})
		Context("When len(agentVersion) == 0 because ParsedResult.Key != 'newrelic.logfile'", func() {
			mockValidateElmentNoAgent := config.ValidateElement{
				ParsedResult: tasks.ValidateBlob{
					Key:      "wrongnewrelic.logfile",
					Children: []tasks.ValidateBlob{},
				},
			}
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							mockValidateElmentNoAgent,
						},
					},
				}
			})
			It("Should return task result status of warning ", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return this task result summary ", func() {
				Expect(result.Summary).To(Equal("Unable to determine PHP Agent version from log file"))
			})
		})
		Context("When len(agentVersion) == 1 because ParsedResult.Key contains only one 'newrelic.logfile' at Key field", func() {
			mockValidateElmentOneAgentFile := config.ValidateElement{
				ParsedResult: tasks.ValidateBlob{
					Key:      "newrelic.logfile",
					Children: []tasks.ValidateBlob{},
				},
			}
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							mockValidateElmentOneAgentFile,
						},
					},
				}
				//injecting mock function here
				p.returnLastMatchInFile = func(search string, filepath string) ([]string, error) {
					return []string{
						"info: New Relic 7.5.0.199 (\"vaughan\"",
						"7.5.0.199",
					}, nil
				}
			})
			It("Should return task result status of Info ", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("Should return task result summary of the php version string ", func() {
				versionString := "7.5.0.199"
				Expect(result.Summary).To(Equal(versionString))
			})
			It("Should return task result payload matching expectedVerPayload ", func() {
				expectedVerPayload := tasks.Ver{
					Major: 7,
					Minor: 5,
					Patch: 0,
					Build: 199,
				}
				Expect(result.Payload).To(Equal(expectedVerPayload))
			})

		})
		Context("When tasks.ParseVersions returns an error", func() {
			mockValidateElementVersionErrors := config.ValidateElement{
				ParsedResult: tasks.ValidateBlob{
					Key:      "newrelic.logfile",
					RawValue: "VersionFormatErrors",
				},
			}
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							mockValidateElementVersionErrors,
						},
					},
				}
				p.returnLastMatchInFile = func(search string, filepath string) ([]string, error) {
					return []string{
						"info: New Relic 7.5.0.199 (\"vaughan\"",
						"%^&^&%",
					}, nil
				}
			})
			It("Should return task result status of Warning", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return task result summary of err.Erro()", func() {
				Expect(result.Summary).To(Equal("unable to convert %^&^&% to an integer"))
			})
		})
		Context("When len(agentVersion) > 1 ", func() {
			mockValidateElmentOneAgentFile := config.ValidateElement{
				ParsedResult: tasks.ValidateBlob{
					Key:      "newrelic.logfile",
					Children: []tasks.ValidateBlob{},
				},
			}
			mockValidateElementNoChildren := config.ValidateElement{
				ParsedResult: tasks.ValidateBlob{
					Key:      "newrelic.logfile",
					RawValue: "ForTestingAgentVersion>1",
				},
			}
			BeforeEach(func() {
				upstream = map[string]tasks.Result{
					"PHP/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							mockValidateElmentOneAgentFile,
							mockValidateElementNoChildren,
						},
					},
				}
				p.returnLastMatchInFile = func(search string, filepath string) ([]string, error) {
					matches := []string{
						"info: New Relic 7.5.0.199 (\"vaughan\"",
						"7.5.0.199",
					}
					if filepath == "ForTestingAgentVersion>1" {
						matches[1] = "TestingMoreThanOneAgentVersion"
					}
					return matches, nil
				}
			})
			It("Should return task result status of Warning", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return task result summary of Expected 1, but found 2  versions of the PHP Agent", func() {
				Expect(result.Summary).To(Equal("Expected 1, but found 2 versions of the PHP Agent"))
			})
		})
	})
})
