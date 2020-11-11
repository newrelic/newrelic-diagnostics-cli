package agent

// Tests for Infra/Config/IntegrationsCollect

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Infra/Agent/Debug", func() {
	
	var p InfraAgentDebug

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Infra",
				Subcategory: "Agent",
				Name:        "Debug",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExaplanation := "Dynamically enable New Relic Infrastructure agent debug logging by running newrelic-infra-ctl"
			Expect(p.Explain()).To(Equal(expectedExaplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Base/Log/Copy", "Infra/Agent/Version", "Base/Env/CollectEnvVars"}
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

		Context("upstream dependency for log collection failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.Info,
					},
					"Base/Log/Copy": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("No New Relic Infrastructure log files detected. If your log files are in a custom location, re-run the " + tasks.ThisProgramFullName + " after setting the NRIA_LOG_FILE environment variable."))
			})
		})

		Context("upstream dependency for infra agent detection failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("New Relic Infrastructure agent not detected on system. Please ensure the Infrastructure agent is installed and running."))
			})
		})

		Context("Upstream Infra agent version returned unexpected type", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
					"Infra/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: []string{"a", "b", "c"},
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Task did not meet requirements necessary to run: type assertion failure"))
			})
		})

		Context("when there is an error executing newrelic-infra-ctl", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 1,
							Minor: 4,
							Patch: 0,
							Build: 0,
						},
					},
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
				}
				p = InfraAgentDebug{
					cmdExecutor: func(a string, b ...string) ([]byte, error) {
						return []byte("Additional error details"), errors.New("newrelic-infra-ctl not found in $PATH")
					},
					runtimeOS: "linux",
				}
			})

			It("should return an expected none result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Error executing newrelic-infra-ctl: newrelic-infra-ctl not found in $PATH\n\tAdditional error details"))
			})
		})
		Context("upstream dependency for infra agent detection failed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.None,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("New Relic Infrastructure agent not detected on system. Please ensure the Infrastructure agent is installed and running."))
			})
		})
		Context("Detected infra agent version is too old to feature infra-ctl app", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 1,
							Minor: 2,
							Patch: 0,
							Build: 0,
						},
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Infrastructure debug CTL binary not available in detected version of Infrastructure Agent(1.2.0.0). Minimum required Infrastructure Agent version is: 1.4.0.0"))
			})

		})

		Context("When Running on Windows", func() {
			Describe("and the infra version is not supported", func() { 
				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Log/Copy": tasks.Result{
							Status: tasks.Success,
						},
						"Infra/Agent/Version": tasks.Result{
							Status: tasks.Info,
							Payload: tasks.Ver{
								Major: 1,
								Minor: 5,
								Patch: 0,
								Build: 0,
							},
						},
						"Base/Env/CollectEnvVars": tasks.Result{
							Status: tasks.None,
						},
					}
					p.runtimeOS = "windows"
				})
				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Infrastructure debug CTL binary not available in detected version of Infrastructure Agent(1.5.0.0). Minimum required Infrastructure Agent version is: 1.7.0.0"))
				})
			})
			Describe("and the infra version is supported", func() { 
				BeforeEach(func() {

					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"Base/Log/Copy": tasks.Result{
							Status: tasks.Success,
						},
						"Infra/Agent/Version": tasks.Result{
							Status: tasks.Info,
							Payload: tasks.Ver{
								Major: 1,
								Minor: 7,
								Patch: 1,
								Build: 0,
							},
						},
						"Base/Env/CollectEnvVars": tasks.Result{
							Status: tasks.Info,
							Payload: map[string]string{
								"ProgramFiles": `C:\Program Files`,
							},
						},
					}
					p.runtimeOS = "windows"
					p.blockWithProgressbar = func(int) {}
					p.cmdExecutor = func(string, ...string) ([]byte, error) { return []byte{}, nil }
				})
				It("should return an expected result status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})

				It("should return an expected result summary", func() {
					Expect(result.Summary).To(Equal("Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl"))
				})
			})
		})

		Context("when debug logs are enabled and the wait is completed", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Infra/Agent/Version": tasks.Result{
						Status: tasks.Info,
						Payload: tasks.Ver{
							Major: 1,
							Minor: 4,
							Patch: 0,
							Build: 0,
						},
					},
					"Base/Log/Copy": tasks.Result{
						Status: tasks.Success,
					},
				}
				p = InfraAgentDebug{
					cmdExecutor: func(a string, b ...string) ([]byte, error) {
						return []byte("Debug logging enabled"), nil
					},
					blockWithProgressbar: func(a int) {},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected none result summary", func() {
				Expect(result.Summary).To(Equal("Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl"))
			})
		})

	})
})
