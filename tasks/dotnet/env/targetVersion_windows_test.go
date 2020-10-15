package env

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"


)

func TestDotNetEnv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dotnet/Env/*")
}

var _ = Describe("Dotnet/Env/TargetVersion", func() {
	var p DotNetEnvTargetVersion
	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "TargetVersion", Category: "DotNet", Subcategory: "Env"}))
		})
	})
	Describe("Explain()", func() {
		It("Should return Explain string", func() {
			Expect(p.Explain()).To(Equal("Determine version of .NET application"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return Dependencies", func() {
			Expect(p.Dependencies()).To(Equal([]string{"DotNet/Agent/Installed"}))
		})
	})

	Describe("execute()", func() {
		var (
			options  tasks.Options
			upstream map[string]tasks.Result
			result   tasks.Result
		)
		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})
		Context("when upstream dependency task failed", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Failure,
					},
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Did not detect .Net Agent as being installed, this check did not run"))
			})
		})
		Context("when unable to get working directory", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", errors.New("This is an error")
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error getting current working directory."))
			})
		})
		Context("when no config files found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to find app config file. Are you running NR Diag from your application's parent directory?"))
			})
		})
		Context("when unable to find .NET version", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/app.config"}
				}
				p.returnStringInFile = func(string, string) ([]string, error) {
					return []string{}, nil
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to find .NET version."))
			})
		})
		Context("when error finding .NET version in config files", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/app.config"}
				}
				p.returnStringInFile = func(string, string) ([]string, error) {
					return nil, errors.New("This is a special news announcement")
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error finding targetFramework"))
			})
		})
		Context("when .NET version found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/app.config"}
				}
				p.returnStringInFile = func(string, string) ([]string, error) {
					return []string{`httpRuntime targetFramework="4.5`, "4.5"}, nil
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("4.5"))
			})
			It("should return an expected result payload", func() {
				Expect(result.Payload).To(Equal([]string{"4.5"}))
			})
		})
		Context("when multiple .NET versions found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"fixtures/App.config", "fixtures/Web.config"}
				}
				p.returnStringInFile = tasks.ReturnStringSubmatchInFile
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("4.6.1, 4.7"))
			})
			It("should return an expected result payload", func() {
				Expect(result.Payload).To(Equal([]string{"4.6.1", "4.7"}))
			})
		})
		Context("when .NET version from app.config", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"fixtures/App.config"}
				}
				p.returnStringInFile = tasks.ReturnStringSubmatchInFile
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("4.6.1"))
			})
			It("should return an expected result payload", func() {
				Expect(result.Payload).To(Equal([]string{"4.6.1"}))
			})
		})
		Context("when .NET version from web.config", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"fixtures/Web.config"}
				}
				p.returnStringInFile = tasks.ReturnStringSubmatchInFile
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Info))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("4.7"))
			})
			It("should return an expected result payload", func() {
				Expect(result.Payload).To(Equal([]string{"4.7"}))
			})
		})
		Context("when an unparsable .NET version is found", func() {
			// We split on ", should receive an error status if this fails
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/app.config"}
				}
				p.returnStringInFile = func(search string, path string) ([]string, error) {
					if search == "httpRuntime targetFramework=\"([0-9.]+)" {
						return []string{`httpRuntime targetFramework=`}, errors.New("string not found")
					}
					return nil, errors.New("string not found")
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error finding targetFramework"))
			})
			It("should return a nil result payload", func() {
				Expect(result.Payload).To(BeNil())
			})
		})
		Context("when there is an error searching the config file for .NETFramework,Version=.*", func() {
			// We split on ", should receive an error status if this fails
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": tasks.Result{
						Status: tasks.Success,
					},
				}
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/app.config"}
				}
				p.returnStringInFile = func(search string, path string) ([]string, error) {
					if search == "httpRuntime targetFramework=\".*" {
						return []string{}, nil
					} else if search == ".NETFramework,Version=.*" {
						return nil, errors.New("dummy error")
					}
					return nil, errors.New("string not found")
				}
			})
			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error finding targetFramework"))
			})
			It("should return a nil result payload", func() {
				Expect(result.Payload).To(BeNil())
			})
		})

	})

})
