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
			Expect(p.Explain()).To(Equal("Check application's .NET version compatibility with New Relic .NET agent"))
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
					"DotNet/Agent/Installed":           tasks.Result{Status: tasks.Success},
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
					"DotNet/Agent/Installed":           tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{Status: tasks.Info},
					"DotNet/Agent/Version":             tasks.Result{Status: tasks.Failure},
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Did not detect .Net Agent version, this check did not run"))
			})
		})
		Context("With valid versions", func() {
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
				Expect(result.Summary).To(Equal(".Net target and Agent version compatible. .Net Target detected as 4.5, Agent version detected as 7.0.2.0"))
			})
		})
		Context("With multiple dotnet target versions detected", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "4.5, 3.5",
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
				Expect(result.Summary).To(Equal("Detected multiple versions for .Net Target or unable to determine targets. .Net Targets detected as 4.5,3.5, Agent version detected as 7.0.2.0"))
			})
		})
		Context("With error parsing dotnet target version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "superman",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "7.0.2.0",
					},
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Error parsing Target .Net version Unable to convert superman to an integer"))
			})
		})
		Context("With unsupported dotnet target version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "1.9",
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
				Expect(result.Summary).To(Equal("Target .Net version detected as below 2.0. This version of .Net is not supported by any agent versions"))
			})
		})
		Context("With supported dotnet target version 2.0", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Summary: "2.0",
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Summary: "6.0.2.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal(".Net target and Agent version compatible. .Net Target detected as 2.0, Agent version detected as 6.0.2.0"))
			})
		})
		Context("With error parsing agent version", func() {
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
						Summary: "I am batmat!",
					},
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Error parsing Agent version"))
			})
		})
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
		})

	})
})
