package log

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Infra/Log/LevelCheck", func() {

	var p InfraLogLevelCheck

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Log",
				Name:        "LevelCheck",
			}
			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})
	})

	Describe("Explain()", func() {
		It("Should return correct explain string", func() {
			Expect(p.Explain()).To(Equal("Check if New Relic Infrastructure agent logging level is set to debug"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return expected dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Infra/Config/Agent"}))
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
		Context("when upstream dependency Infra/Config/Agent fails", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})
			It("Should return a none result", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return the correct summary", func() {
				Expect(result.Summary).To(Equal("Infrastructure Agent config not present"))
			})
		})

		Context("When Execute() succeeds with old verbose configuration", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "verbose",
									RawValue: "1",
								},
							},
						},
					},
				}
			})

			It("Should return a Success", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Summary should contain 'logging level is set to verbose'", func() {
				Expect(result.Summary).To(ContainSubstring("logging level is set to verbose"))
			})

		})

		Context("When Execute() succeeds with new log level configuration", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "log/level",
									RawValue: "debug",
								},
							},
						},
					},
				}
			})

			It("Should return a Success", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Summary should contain 'logging level is set to verbose'", func() {
				Expect(result.Summary).To(ContainSubstring("logging level is set to verbose"))
			})

		})

		Context("When Execute() warns with new log level configuration", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic-infra.yml",
									FilePath: "/etc/",
								},
								ParsedResult: tasks.ValidateBlob{
									Key:      "log/level",
									RawValue: "info",
								},
							},
						},
					},
				}
			})

			It("Should return a Warning", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Summary should contain 'logging level not set to verbose'", func() {
				Expect(result.Summary).To(ContainSubstring("logging level not set to verbose"))
			})
			It("URL should be: https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/troubleshooting/generate-logs-troubleshooting-infrastructure", func() {
				Expect(result.URL).To(Equal("https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/troubleshooting/generate-logs-troubleshooting-infrastructure"))
			})

		})
	})
})
