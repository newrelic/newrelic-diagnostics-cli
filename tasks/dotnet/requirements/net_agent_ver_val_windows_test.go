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
						Payload: []string{"3.0"},
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "8.3.360.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found a Target Framework version(s) that is not supported by the New Relic .NET agent: 3.0"))
			})
		})

		Context("With multiple target frameworks and only one is supported", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Payload: []string{"4.5", "4.0"},
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "7.0.2.0",
					},
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found that your New Relic .NET agent version 7.0.2.0 is not compatible with the following Target .NET version(s): 4.0"))
			})
		})
		Context("With success for one target framework version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Payload: []string{"4.5", "4.6"},
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "7.0.2.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Your .NET agent version 7.0.2.0 is fully compatible with the following found Target .NET version(s): 4.5, 4.6"))
			})
		})

		Context("With multiple dotnet target versions detected and neither are compatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"DotNet/Env/TargetVersion": tasks.Result{
						Status:  tasks.Info,
						Payload: []string{"3.5", "4.0"},
					},
					"DotNet/Agent/Version": tasks.Result{
						Status:  tasks.Info,
						Payload: "7.0.2.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found that your New Relic .NET agent version 7.0.2.0 is not compatible with the following Target .NET version(s): 3.5, 4.0"))
			})
		})

	})
})
