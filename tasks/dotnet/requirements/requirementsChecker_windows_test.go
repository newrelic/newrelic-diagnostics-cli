package requirements

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/ProcessorType", func() {
	var p DotnetRequirementsRequirementCheck
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "RequirementCheck", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain string", func() {
			Expect(p.Explain()).To(Equal("Validate New Relic .NET agent related diagnostic checks"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return Dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{
				"DotNet/Agent/Installed",
				"DotNet/Requirements/NetTargetAgentVersionValidate",
				"DotNet/Requirements/OS",
				"Dotnet/Requirements/OwinCheck",
				"DotNet/Requirements/ProcessorType",
				"DotNet/Requirements/Datastores",
				"DotNet/Requirements/MessagingServicesCheck",
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

		Context("With all successful upstream", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":                            {Status: tasks.Success},
					"DotNet/Requirements/NetTargetAgentVersionValidate": {Status: tasks.Success},
					"DotNet/Requirements/OS":                            {Status: tasks.Success},
					"Dotnet/Requirements/OwinCheck":                     {Status: tasks.Success},
					"DotNet/Requirements/ProcessorType":                 {Status: tasks.Success},
					"DotNet/Requirements/Datastores":                    {Status: tasks.Success},
					"DotNet/Requirements/MessagingServicesCheck":        {Status: tasks.Success},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("All DotNet Agent requirement checks detected as having succeeded."))
			})
		})

		Context("With mix of successful and none upstream", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":                            {Status: tasks.Success},
					"DotNet/Requirements/NetTargetAgentVersionValidate": {Status: tasks.Success},
					"DotNet/Requirements/OS":                            {Status: tasks.Success},
					"Dotnet/Requirements/OwinCheck":                     {Status: tasks.None},
					"DotNet/Requirements/ProcessorType":                 {Status: tasks.Success},
					"DotNet/Requirements/Datastores":                    {Status: tasks.Success},
					"DotNet/Requirements/MessagingServicesCheck":        {Status: tasks.None},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("All DotNet Agent requirement checks detected as having succeeded."))
			})
		})

		Context("With some failed upstream", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":                            {Status: tasks.Success},
					"DotNet/Requirements/NetTargetAgentVersionValidate": {Status: tasks.Failure},
					"DotNet/Requirements/OS":                            {Status: tasks.Failure},
					"Dotnet/Requirements/OwinCheck":                     {Status: tasks.Success},
					"DotNet/Requirements/ProcessorType":                 {Status: tasks.Success},
					"DotNet/Requirements/Datastores":                    {Status: tasks.Success},
					"DotNet/Requirements/MessagingServicesCheck":        {Status: tasks.Success},
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected failed DotNet Agent requirement checks. Failed checks: \nDotNet/Requirements/NetTargetAgentVersionValidate\nDotNet/Requirements/OS\n"))
			})
		})

		Context("With some missing upstream required tasks", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed":                     {Status: tasks.Success},
					"Dotnet/Requirements/OwinCheck":              {Status: tasks.Success},
					"DotNet/Requirements/ProcessorType":          {Status: tasks.Success},
					"DotNet/Requirements/Datastores":             {Status: tasks.Success},
					"DotNet/Requirements/MessagingServicesCheck": {Status: tasks.Success},
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected failed DotNet Agent requirement checks. Failed checks: \nDotNet/Requirements/NetTargetAgentVersionValidate\nDotNet/Requirements/OS\n"))
			})
		})

	})
})
