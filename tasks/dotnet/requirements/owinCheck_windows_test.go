package requirements

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/OwinCheck", func() {
	var p DotnetRequirementsOwinCheck
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "OwinCheck", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain test", func() {
			Expect(p.Explain()).To(Equal("Check application's OWIN version compatibility with New Relic .NET agent"))
		})
	})
	Describe("Dependencies()", func() {
		It("Should return Dependencies list", func() {
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
		Context("With Owin dll not present", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Didn't detect Owin dlls"))
			})
		})
		Context("With Owin dll present but error getting dll version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/bar"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "", errors.New("i'm a little teapot")
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("OWIN dlls detected but unable to confirm OWIN version. See debug logs for more information on error. Version returned i'm a little teapot"))
			})
		})
		Context("With Owin dll present but error validating supported version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/bar"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "Short and Stout", nil
				}
			})
			It("Should return Error status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Error validating OWIN version, see debug logs for more details"))
			})
		})
		Context("With supported Owin dll present", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/bar"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "3.1", nil
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected as OWIN hosted with version 3.1"))
			})
		})
		Context("With unsupported Owin dll present", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/bar"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "2.9", nil
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Detected OWIN dlls but version not detected as v3 or higher. Detected version is 2.9"))
			})
		})

	})
})
