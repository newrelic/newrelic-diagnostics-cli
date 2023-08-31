package log

import (
	"errors"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInfraLogCollect(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Infra/Log/Collect test suite")
}

var _ = Describe("Infra/Log/Collect", func() {

	var p InfraLogCollect

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Log",
				Name:        "Collect",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Collect New Relic Infrastructure agent log files"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return expected dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Infra/Config/Agent", "Base/Config/Validate"}))
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
		Context("when upstream dependency Infra/Config/Collect fails", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": {
						Status: tasks.Failure,
					},
				}
			})
			It("Should return a none result", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return the correct summary", func() {
				Expect(result.Summary).To(Equal("Not executing task. Infra agent not found."))
			})
		})

		Context("When upstream dependency Base/Config/Validate fails", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Failure,
					},
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
			})
			It("Should return a none result", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return the correct summary", func() {
				Expect(result.Summary).To(Equal("Not executing task. Infra config file not found."))
			})
		})

		Context("When Base/Config/Validate payload type assertion fails", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status:  tasks.Success,
						Payload: "not []config.ValidateElement like it should be",
					},
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
			})

			It("Should return a none result", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return the correct summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("When Execute() succeeds", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "log_file",
									RawValue: "/var/log/newrelic-infra/newrelic-infra.log",
								},
							},
						},
					},
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
				p.validatePaths = func([]string) []tasks.CollectFileStatus {
					return []tasks.CollectFileStatus{
						{
							Path:     "/var/log/newrelic-infra/newrelic-infra.log",
							IsValid:  true,
							ErrorMsg: nil,
						},
					}
				}
			})

			It("Should return a Success", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Summary should contain 'logs found'", func() {
				Expect(result.Summary).To(ContainSubstring("logs found"))
			})
			It("Should return an slice of fileCopyEnvelope with expected field", func() {
				Expect(result.FilesToCopy[0].Path).To(Equal("/var/log/newrelic-infra/newrelic-infra.log"))
			})

			It("Should return a payload containing a slice of strings", func() {
				_, ok := result.Payload.([]string)
				Expect(ok).To(Equal(true))
			})

			It("Should return a payload containing a slice of strings with expected value", func() {
				stringSlice, _ := result.Payload.([]string)
				Expect(stringSlice[0]).To(Equal("/var/log/newrelic-infra/newrelic-infra.log"))
			})

		})

		Context("When Execute() sets a warning because the log path provided was not accessible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "log_file",
									RawValue: "/var/log/newrelic-infra/newrelic-infra.log",
								},
							},
						},
					},
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
				p.validatePaths = func([]string) []tasks.CollectFileStatus {
					return []tasks.CollectFileStatus{
						{
							Path:     "/var/log/newrelic-infra/newrelic-infra.log",
							IsValid:  false,
							ErrorMsg: errors.New("stat /var/log/newrelic-infra/newrelic-infra.log: no such file or directory"),
						},
					}
				}
			})

			It("Should return a warning status and message", func() {
				expectedSummary := `The log file path found in the New Relic config file ("/var/log/newrelic-infra/newrelic-infra.log") did not provide a file that was accessible to us:
"stat /var/log/newrelic-infra/newrelic-infra.log: no such file or directory"
If you are working with a support ticket, manually provide your New Relic log file for further troubleshooting`
				Expect(result.Status).To(Equal(tasks.Warning))
				Expect(result.Summary).To(Equal(expectedSummary))
			})

		})

		Context("When no log file is specified in config", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
							},
						},
					},
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
			})
			It("Should return a None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("New Relic Infrastructure configuration file did not specify log file path"))
			})

		})
	})
	Describe("getLogFilePaths() with old log file configuration option", func() {
		var (
			result            []string
			parsedConfigFiles []config.ValidateElement
		)
		JustBeforeEach(func() {
			result = p.getLogFilePaths(parsedConfigFiles)
		})
		Context("When given a slice of validate elements containing log file entries", func() {
			parsedConfigFiles = []config.ValidateElement{
				{
					Config: config.ConfigElement{
						FileName: "newrelic-infra.yml",
						FilePath: "/etc/",
					},
					ParsedResult: tasks.ValidateBlob{
						Key:      "log_file",
						RawValue: "/var/log/messages",
					},
				},
			}
			It("Should return a slice of log_file paths", func() {
				expectedResult := []string{"/var/log/messages"}
				Expect(result).To(Equal(expectedResult))
			})
		})

	})
	Describe("getLogFilePaths() with new log file configuration option", func() {
		var (
			result            []string
			parsedConfigFiles []config.ValidateElement
		)
		JustBeforeEach(func() {
			result = p.getLogFilePaths(parsedConfigFiles)
		})
		Context("When given a slice of validate elements containing log file entries", func() {
			parsedConfigFiles = []config.ValidateElement{
				{
					Config: config.ConfigElement{
						FileName: "newrelic-infra.yml",
						FilePath: "/etc/",
					},
					ParsedResult: tasks.ValidateBlob{
						Key:      "log/file",
						RawValue: "/var/log/newrelic-infra/newrelic-infra.log",
					},
				},
			}
			p.findFiles = func([]string, []string) []string {
				return []string{
					"/var/log/newrelic-infra/newrelic-infra.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-15_11-12-04.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-15_12-12-03.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-11_12-12-03.log.gz",
				}
			}
			It("Should return a slice of log_file paths", func() {
				expectedResult := []string{
					"/var/log/newrelic-infra/newrelic-infra.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-15_11-12-04.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-15_12-12-03.log",
					"/var/log/newrelic-infra/newrelic-infra_2022-07-11_12-12-03.log.gz",
				}
				Expect(result).To(Equal(expectedResult))
			})
		})

	})
})
