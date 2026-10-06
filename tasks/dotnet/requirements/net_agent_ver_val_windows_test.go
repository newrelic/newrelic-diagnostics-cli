//go:build windows
// +build windows

package requirements

import (
	. "github.com/onsi/ginkgo/v2"
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
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Failure,
					},
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal(tasks.UpstreamFailedSummary + "DotNet/Agent/Installed"))
			})
		})
		Context("With unsuccessful upstream DotnetTarget", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":   {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {Status: tasks.Failure},
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
					"DotNet/Agent/Installed":   {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {Status: tasks.Info},
					"DotNet/Agent/Version":     {Status: tasks.Failure},
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
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"3.0"},
					},
					"DotNet/Agent/Version": {
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
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"4.8", "4.0"},
					},
					"DotNet/Agent/Version": {
						Status:  tasks.Info,
						Payload: "10.45.0.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found that your New Relic .NET agent version 10.45.0.0 is not compatible with the following Target .NET version(s): 4.0"))
			})
		})
		Context("With multiple supported target framework versions", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"4.6.2", "4.8"},
					},
					"DotNet/Agent/Version": {
						Status:  tasks.Info,
						Payload: "10.45.0.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Your .NET agent version 10.45.0.0 is fully compatible with the following found Target .NET version(s): 4.6.2, 4.8"))
			})
		})
		Context("With .NET Framework 4.8.1 and a compatible agent", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"4.8.1"},
					},
					"DotNet/Agent/Version": {
						Status:  tasks.Info,
						Payload: "10.45.0.0",
					},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
		})
		Context("With .NET Framework 4.6.1, which is no longer supported", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"4.6.1"},
					},
					"DotNet/Agent/Version": {
						Status:  tasks.Info,
						Payload: "10.45.0.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found a Target Framework version(s) that is not supported by the New Relic .NET agent: 4.6.1"))
			})
		})
		Context("With .NET Framework 4.7.2 and an agent older than 10.0", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"4.7.2"},
					},
					"DotNet/Agent/Version": {
						Status:  tasks.Info,
						Payload: "9.9.0.0",
					},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("We found that your New Relic .NET agent version 9.9.0.0 is not compatible with the following Target .NET version(s): 4.7.2"))
			})
		})

		Context("With multiple dotnet target versions detected and neither are compatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
					"DotNet/Env/TargetVersion": {
						Status:  tasks.Info,
						Payload: []string{"3.5", "4.0"},
					},
					"DotNet/Agent/Version": {
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
	Describe("normalizeFrameworkVersions()", func() {
		It("Should drop a zero patch and keep a non-zero patch", func() {
			normalized, err := normalizeFrameworkVersions([]string{"4.8", "4.8.0.0", "4.7.2", "4.6.2.0", "4.0"})
			Expect(err).To(BeNil())
			Expect(normalized).To(Equal([]string{"4.8", "4.8", "4.7.2", "4.6.2", "4.0"}))
		})
		It("Should return an error for an unparseable version", func() {
			_, err := normalizeFrameworkVersions([]string{"not-a-version"})
			Expect(err).ToNot(BeNil())
		})
	})
})
