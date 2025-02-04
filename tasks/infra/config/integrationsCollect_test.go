package config

// Tests for Infra/Config/IntegrationsCollect

import (
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Infra/Config/IntegrationsCollect", func() {
	var p InfraConfigIntegrationsCollect

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Config",
				Name:        "IntegrationsCollect",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Collect New Relic Infrastructure on-host integration configuration and definition files"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Infra/Config/Agent"}
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
					"Infra/Config/Agent": {
						Status: tasks.None,
					},
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No Infra Agent detected. Task not executed."))
			})
		})

		Context("no integration configuration files are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Config/Agent": {
						Status: tasks.Success,
					},
				}
				p = InfraConfigIntegrationsCollect{fileFinder: func([]string, []string) []string {
					return []string{}
				}}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("No on-host integration yml files found"))
			})
		})

		if runtime.GOOS != "windows" {
			Context("two integration configuration files are found on non-windows system", func() {

				BeforeEach(func() {
					options = tasks.Options{
						Options: map[string]string{
							"YesToAll": "true",
						},
					}
					upstream = map[string]tasks.Result{
						"Infra/Config/Agent": {
							Status: tasks.Success,
						},
					}
					p = InfraConfigIntegrationsCollect{fileFinder: func([]string, []string) []string {
						return []string{"config/path/config.yml", "definition/path/definition.yml"}
					}}
				})

				It("should return an expected Success result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return an expected success result summary", func() {
					Expect(result.Summary).To(Equal("2 on-host integration yml file(s) found"))
				})

				It("should return an expected success result payload", func() {
					expectedPayload := []config.ConfigElement{
						{FileName: "config.yml", FilePath: "config/path/"},
						{FileName: "definition.yml", FilePath: "definition/path/"},
					}
					Expect(result.Payload).To(Equal(expectedPayload))
				})

				It("should return an expected success result FilesToCopy", func() {

					expectedFilesToCopy := []tasks.FileCopyEnvelope{
						{Path: "config/path/config.yml"},
						{Path: "definition/path/definition.yml"},
					}
					Expect(result.FilesToCopy).To(Equal(expectedFilesToCopy))
				})
			})
		}

		if runtime.GOOS == "windows" {
			Context("two integration configuration files are found on windows", func() {

				BeforeEach(func() {
					options = tasks.Options{
						Options: map[string]string{
							"YesToAll": "true",
						},
					}
					upstream = map[string]tasks.Result{
						"Infra/Config/Agent": {
							Status: tasks.Success,
						},
					}
					p = InfraConfigIntegrationsCollect{fileFinder: func([]string, []string) []string {
						return []string{`C:\Program Files\New Relic\newrelic-infra\integrations.d\config.yml`, `C:\Program Files\New Relic\newrelic-infra\custom-integrations\definition.yml`}
					}}
				})

				It("should return an expected Success result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return an expected success result summary", func() {
					Expect(result.Summary).To(Equal("2 on-host integration yml file(s) found"))
				})

				It("should return an expected success result payload", func() {
					expectedPayload := []config.ConfigElement{
						{FileName: "config.yml", FilePath: `C:\Program Files\New Relic\newrelic-infra\integrations.d\`},
						{FileName: "definition.yml", FilePath: `C:\Program Files\New Relic\newrelic-infra\custom-integrations\`},
					}
					Expect(result.Payload).To(Equal(expectedPayload))
				})

				It("should return an expected success result FilesToCopy", func() {

					expectedFilesToCopy := []tasks.FileCopyEnvelope{
						{Path: `C:\Program Files\New Relic\newrelic-infra\integrations.d\config.yml`},
						{Path: `C:\Program Files\New Relic\newrelic-infra\custom-integrations\definition.yml`},
					}
					Expect(result.FilesToCopy).To(Equal(expectedFilesToCopy))
				})
			})
		}
	})
})
