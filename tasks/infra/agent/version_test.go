package agent

// Tests for Infra/Config/IntegrationsCollect

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Infra/Agent/Version", func() {
	var p InfraAgentVersion

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Agent",
				Name:        "Version",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExaplanation := "Determine version of New Relic Infrastructure agent"
			Expect(p.Explain()).To(Equal(expectedExaplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Infra/Config/Agent", "Base/Env/CollectEnvVars"}
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

		Context("upstream dependency for retrieving environment variables failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.None,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Upstream dependency failed"))
			})
		})

		Context("upstream dependency for retrieving environment variables returns unexpected payload type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: "I should be a map[string]string{}, but I'm a string",
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(tasks.AssertionErrorSummary))
			})
		})

		Context("upstream dependency for detecting presence of infrastructure agent failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: Infrastructure Agent not detected on system"))
			})
		})

		Context("Infrastructure agent is present and returns version in the expected format", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "darwin"
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("New Relic Infrastructure Agent version: 1.5.40"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("1.5.40"))
			})
		})

		Context("Infrastructure agent is present on windows but ProgramFiles env variable is not set", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "windows"
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("New Relic Infrastructure Agent version: 1.5.40"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to determine New Relic Infrastructure binary path: ProgramFiles environment variable not set"))
			})
		})

		Context("Infrastructure agent is present but -version returns unparseable result", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Env/CollectEnvVars": tasks.Result{
						Status:  tasks.Info,
						Payload: map[string]string{},
					},
					"Infra/Config/Agent": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.runtimeOS = "linux"
				p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
					return []byte("newrelic-infra: Permission denied"), nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to parse New Relic Infrastructure Agent version from: newrelic-infra: Permission denied"))
			})

			Context("Infrastructure agent is present but version check returns an error", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Env/CollectEnvVars": tasks.Result{
							Status:  tasks.Info,
							Payload: map[string]string{},
						},
						"Infra/Config/Agent": tasks.Result{
							Status: tasks.Success,
						},
					}
					p.runtimeOS = "linux"
					p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
						return []byte(""), errors.New("Fromlet was defrobozticated")
					}

				})

				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("New Relic Infrastructure Agent version could not be determined because Diagnostics CLI encountered this issue when running the command 'newrelic-infra -version': Fromlet was defrobozticated"))
				})
			})
			Context("Infrastructure agent is present but version check returns unparseable version string", func() {

				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Env/CollectEnvVars": tasks.Result{
							Status:  tasks.Info,
							Payload: map[string]string{},
						},
						"Infra/Config/Agent": tasks.Result{
							Status: tasks.Success,
						},
					}
					p.runtimeOS = "linux"
					p.cmdExecutor = func(a string, b ...string) ([]byte, error) {
						return []byte("New Relic Infrastructure Agent version: .19.1."), nil
					}

				})

				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Error))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Unable to parse New Relic Infrastructure Agent version from: New Relic Infrastructure Agent version: .19.1."))
				})
			})
		})

		Describe("getBinaryPath()", func() {

			var (
				envVars map[string]string
				result  string
				err     error
			)

			JustBeforeEach(func() {
				result, err = p.getBinaryPath(envVars)
			})
			Context("On Windows systems", func() {
				BeforeEach(func() {
					envVars = map[string]string{
						"ProgramFiles": `C:\Program Files`,
					}
					p.runtimeOS = "windows"
				})

				It("should reference the ProgramFiles ENV var when constructing the path", func() {
					Expect(result).To(Equal(`C:\Program Files\New Relic\newrelic-infra\newrelic-infra.exe`))
					Expect(err).To(BeNil())
				})
			})

			Context("On non-windows systems", func() {
				BeforeEach(func() {
					p.runtimeOS = "darwin"
				})

				It("should expect the newrelic-infra binary to exist in the PATH", func() {
					Expect(result).To(Equal(`newrelic-infra`))
				})
			})
		})
	})
})
