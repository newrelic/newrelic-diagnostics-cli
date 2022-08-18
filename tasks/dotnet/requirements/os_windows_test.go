//go:build !darwin || !linux
// +build !darwin !linux

package requirements

import (
	tasks "github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/ProcessorType", func() {
	var p DotnetRequirementsOS
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "OS", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain string", func() {
			Expect(p.Explain()).To(Equal("Check operating system compatibility with New Relic .NET agent"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return Dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{
				"DotNet/Agent/Installed",
				"Base/Env/HostInfo",
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
		Context("With unsuccessful upstream", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status:  tasks.None,
						Summary: tasks.NoAgentDetectedSummary,
					},
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal(tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed"))
			})
		})
		Context("With invalid payload from hostinfo task", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: "string"},
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Could not resolve payload of dependent task, HostInfo."))
			})
		})
		Context("With incomplete upstream payload", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: env.HostInfo{PlatformVersion: ""}},
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Could not get OS version to check compatibility"))
			})
		})
		Context("With osVersion empty string", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: env.HostInfo{PlatformVersion: ""}},
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Could not get OS version to check compatibility"))
			})
		})
		Context("With supported OS", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: env.HostInfo{PlatformVersion: "10.0.14393 Build 14393"}},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("OS detected as meeting requirements. See HostInfo task Payload for more info on OS"))
			})
		})
		Context("With server 2003", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo": tasks.Result{
						Payload: env.HostInfo{
							PlatformVersion: "5.2.111 Build 111",
							PlatformFamily:  "Server",
						},
					},
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("OS detected as Windows Server 2003. Last .Net Agent version compatible is 6.11.613"))
			})
		})
		Context("With -1 getting OS version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: env.HostInfo{PlatformVersion: "the.black.knight"}},
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Encountered issues getting full OS version to check compatibility"))
			})
		})
		Context("With unsupported OS version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
					"Base/Env/HostInfo":      tasks.Result{Payload: env.HostInfo{PlatformVersion: "3.1"}},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("OS not detected as compatible with the .Net Framework Agent"))
			})
		})

	})
})
