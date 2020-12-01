// +build windows

package requirements

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	tasks "github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var _ = Describe("Dotnet/Requirements/NetTargetAgentVerValidate", func() {
	var p DotnetRequirementsNetTargetAgentVerValidate
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "NetTargetAgentVersionValidate", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain string", func() {
			Expect(p.Explain()).To(Equal("Check application's .NET Framework version compatibility with New Relic .NET agent"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return Dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{
				"DotNet/Agent/Installed",
				"DotNet/Env/TargetVersion",
				"DotNet/Agent/Version",
			}))
		})
	})
	Describe("Execute", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("With unsuccessful upstream agent detection", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Did not detect .Net Agent as being installed, this check did not run"))
			})
		})
		Context("With unsuccessful upstream DotnetTarget", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":   tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{Status: tasks.Failure},
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Did not detect App Target .Net version, this check did not run"))
			})
		})
		Context("With unsuccessful upstream DotNetAgentVersion", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":   tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{Status: tasks.Info},
					"DotNet/Agent/Version":     tasks.Result{Status: tasks.Failure},
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Did not detect .Net Agent version, this check did not run"))
			})
		})
		Context("With unsupported Dotnet Framework Version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "3.0",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("The detected Target .NET version is not supported by any .NET agent version: 3.0"))
			})
		})
		Context("With error parsing dotnet framework", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "4.0",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "hot dogs and baloney",
					},
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Error parsing Target .Net Agent version Unable to convert hot dogs and baloney to an integer"))
			})
		})
		Context("With multiple target frameworks and only one is supported", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "4.5,4.0",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected multiple target .NET versions.\nThe target .NET versions detected as: 4.5,4.0 and Agent version detected as: 7.0.2.0\nIncompatible Version detected: 4.0\nCompatible Version detected: 4.5\n"))
			})
		})
		Context("With success for one target framework version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "4.5",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Compatible Version detected: 4.5\n"))
			})
		})
		Context("With success for multiple target framework versions", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "4.5, 4.6",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Compatible Version detected: 4.5\nCompatible Version detected: 4.6\n"))
			})
		})
		Context("With multiple dotnet target versions detected and neither are compatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "3.5, 4.0",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected multiple target .NET versions.\nThe target .NET versions detected as: 3.5,4.0 and Agent version detected as: 7.0.2.0\nIncompatible Version detected: 3.5\nIncompatible Version detected: 4.0\n"))
			})
		})
		/*
			Context("With compatible versions multiple dotnet targets", func() {
				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
						"DotNet/Env/TargetVersion": tasks.Result{
							Status:  tasks.Info,
							Summary: "4.5, 4.5",
						},
						"DotNet/Agent/Version": tasks.Result{
							Status:  tasks.Info,
							Summary: "6.18",
						},
					}
				})
				It("Should return Success status", func() {
					Expect(result.Status).To(Equal(tasks.Success))
				})
				It("Should return expected summary", func() {
					Expect(result.Summary).To(Equal(".Net target and Agent version compatible. .Net Target detected as 4.5, Agent version detected as 6.18"))
				})
			})
			Context("With invalid combination of versions", func() {
				BeforeEach(func() {
					options = tasks.Options{}
					upstream = map[string]tasks.Result{
						"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
						"DotNet/Env/TargetVersion": tasks.Result{
							Status:  tasks.Info,
							Summary: "4",
						},
						"DotNet/Agent/Version": tasks.Result{
							Status:  tasks.Info,
							Summary: "7.0.2.0",
						},
					}
				})
				It("Should return Failure status", func() {
					Expect(result.Status).To(Equal(tasks.Failure))
				})
				It("Should return expected summary", func() {
					Expect(result.Summary).To(Equal("App detected as targeting a version of .Net below 4.5 with an Agent version of 7 or above. .Net Target detected as 4, Agent version detected as 7.0.2.0"))
				})
			})*/

	})
})
