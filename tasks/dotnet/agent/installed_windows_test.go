package agent

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func mockDllSuccess(agentPath, profilerPath string) bool {
	return true
}
func mockDllFail(agentPath, profilerPath string) bool {
	return false
}

func successfulUpstream() map[string]tasks.Result {
	return map[string]tasks.Result{
		"Base/Config/Validate": tasks.Result{Payload: []config.ValidateElement{
			config.ValidateElement{
				Config: config.ConfigElement{FileName: "newrelic.config"},
			},
		},
		},
	}
}

func TestCollector(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DotNetAgentInstalled Suite")
}

var _ = Describe("DotNetAgentInstalled", func() {
	var p DotNetAgentInstalled

	Context("Identifier()", func() {
		It("Should return identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "Installed", Category: "DotNet", Subcategory: "Agent"}))
		})
	})

	Context("Explain()", func() {
		It("Should return explain", func() {
			Expect(p.Explain()).To(Equal("Detect New Relic .NET agent"))
		})
	})

	Context("Dependencies()", func() {
		It("Should return list of dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"Base/Config/Validate"}))
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

		Context("Upstream dependency for config validation fails", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Not executing task: .NET agent config file not found."))
			})
		})

		Context("When no .NET config files are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.yml", 
									FilePath: "",
								}, 
							},
						},
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(".NET agent not detected"))
			})
		})

		Context("When no .NET config files are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.yml", 
									FilePath: "",
								}, 
							},
						},
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(".NET agent not detected"))
			})
		})

		Context("When .NET config files are found, but no agent DLLs are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{
						AgentPath: "fixtures/does_not_exist.dll",
						ProfilerPath: "fixtures/does_not_exist.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Could NOT find one or more dlls required by the .Net Agent. Either the .NET Agent is not installed or missing essential dlls. Try running the installer to resolve the issue."))
			})
		})

		Context("When .NET config files are found, but profiler dll is missing", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{
						AgentPath: "fixtures/file1.dll",
						ProfilerPath: "fixtures/does_not_exist.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Could NOT find one or more dlls required by the .Net Agent. Either the .NET Agent is not installed or missing essential dlls. Try running the installer to resolve the issue."))
			})
		})

		Context("When .NET config files are found, but agent dll is missing", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{
						AgentPath: "fixtures/does_not_exist.dll",
						ProfilerPath: "fixtures/file1.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Could NOT find one or more dlls required by the .Net Agent. Either the .NET Agent is not installed or missing essential dlls. Try running the installer to resolve the issue."))
			})
		})

		Context("When .NET config files and agent and profile dlls are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{
						AgentPath: "fixtures/file1.dll",
						ProfilerPath: "fixtures/file2.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Found dlls required by the .NET Agent"))
			})
		})

		Context("When .NET config files and agent and profile dlls are found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{
						AgentPath: "fixtures/file1.dll",
						ProfilerPath: "fixtures/file2.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Found dlls required by the .NET Agent"))
			})
		})

		Context("When .NET config files and agent and profile dlls are found for both current and legacy agents", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []config.ValidateElement{
							config.ValidateElement{
								Config: config.ConfigElement{
									FileName: "newrelic.config", 
									FilePath: "",
								}, 
							},
						},
					},
				}
				p.agentInstallPaths = []DotNetAgentInstall{
					DotNetAgentInstall{ //current agent
						AgentPath: "fixtures/file1.dll",
						ProfilerPath: "fixtures/file1.dll",
					},
					DotNetAgentInstall{ //legacy agent
						AgentPath: "fixtures/file2.dll",
						ProfilerPath: "fixtures/file2.dll",
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Found dlls required by the .NET Agent"))
			})

			It("should return payload specifying current agent", func() {
				resultPayload, _ := result.Payload.(DotNetAgentInstall)
				Expect(resultPayload.AgentPath).To(Equal("fixtures/file1.dll"))
			})
		})

	})

})
