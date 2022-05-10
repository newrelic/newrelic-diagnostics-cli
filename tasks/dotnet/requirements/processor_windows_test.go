package requirements

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/ProcessorType", func() {
	var p DotnetRequirementsProcessorType
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "ProcessorType", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain string", func() {
			Expect(p.Explain()).To(Equal("Check processor architecture compatibility with New Relic .NET agent"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return Dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"DotNet/Agent/Installed"}))
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
		Context("With error getting processor type", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getProcessorArch = func() (string, error) {
					return "", errors.New("i am a banana")
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Error while getting Processor type, see debug logs for more details."))
			})
		})
		Context("With supported processor type x86", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getProcessorArch = func() (string, error) {
					return "x86", nil
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Processor detected as x86"))
			})
		})
		Context("With supported processor type AMD64", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getProcessorArch = func() (string, error) {
					return "AMD64", nil
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Processor detected as x64"))
			})
		})

		Context("With unsupported processor type", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getProcessorArch = func() (string, error) {
					return "banana", nil
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Processor detected as banana. .Net Framework Agent only supports x86 and x64 processors"))
			})
		})

	})
})
