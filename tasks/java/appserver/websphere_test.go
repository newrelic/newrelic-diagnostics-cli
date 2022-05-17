package appserver

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JavaAppServerWebSphere", func() {

	var (
		p JavaAppServerWebSphere
	)

	Describe("Identifier", func() {
		It("Should return the identifier", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "WebSphere", Category: "Java", Subcategory: "AppServer"}))
		})
	})
	Describe("Explain", func() {
		It("Should return explain", func() {
			Expect(p.Explain()).To(Equal("Check Websphere AS version compatibility with New Relic Java agent"))
		})
	})
	Describe("Dependencies", func() {
		It("Should return dependencies list", func() {
			Expect(p.Dependencies()).To(Equal([]string{}))
		})
	})

	Describe("Execute", func() {
		var (
			result tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(tasks.Options{}, map[string]tasks.Result{})
		})
		Context("When directoryGetter returns error", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "", errors.New("error getting menu")
				}
				p.osGetExecutable = func() (string, error) {
					return "", errors.New("error getting list of burritos")
				}
			})
			It("Should return error Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should return error Summary", func() {
				Expect(result.Summary).To(Equal("Task unable to complete. Errors encountered while determining path to the executable and working dir."))
			})
		})
		Context("When fileFinder returns no directories", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.osGetExecutable = func() (string, error) {
					return "", nil
				}

				p.fileFinder = func([]string, []string) []string {
					return []string{}
				}
			})
			It("Should return none Status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return error Summary", func() {
				Expect(result.Summary).To(Equal("Websphere was not detected in this environment."))
			})
		})
		Context("When fileFinder returns directories", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "", nil
				}
				p.osGetExecutable = func() (string, error) {
					return "", nil
				}
				p.fileFinder = func([]string, []string) []string {
					return []string{"/foo", "/bar"}
				}
				p.returnSubstring = func(string, string) ([]string, error) {
					return []string{}, nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return true, nil
				}
			})
			It("Should return Success Status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected Summary", func() {
				Expect(result.Summary).To(Equal("Supported version of Websphere detected in this environment."))
			})
		})

	})

	Describe("getDirs", func() {
		var (
			dirs      []string
			getDirErr error
		)

		JustBeforeEach(func() {
			dirs, getDirErr = p.getDirs()
		})

		Context("When getwd returns error but osexecutable returns directory", func() {
			BeforeEach(func() {

				p.osGetwd = func() (string, error) {
					return "", errors.New("one burrito bowl")
				}
				p.osGetExecutable = func() (string, error) {
					return "/bar", nil
				}
			})
			It("Should return directory list", func() {
				Expect(dirs).To(Equal([]string{"/bar"}))

			})
			It("Should return nil error", func() {
				Expect(getDirErr).To(BeNil())
			})
		})

		Context("When getwd returns direcotry but osexecutable returns error", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "/foo", nil
				}
				p.osGetExecutable = func() (string, error) {
					return "", errors.New("one nachos plate")
				}
			})
			It("Should return directory list", func() {
				Expect(dirs).To(Equal([]string{"/foo"}))
			})
			It("Should return nil error", func() {
				Expect(getDirErr).To(BeNil())
			})

		})

		Context("When getwd returns error and osexecutable returns error", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "", errors.New("one burrito error")
				}
				p.osGetExecutable = func() (string, error) {
					return "", errors.New("two burrito error")
				}
			})
			It("Should return empty directory list", func() {
				Expect(dirs).To(BeEmpty())
			})
			It("Should return expected error", func() {
				Expect(getDirErr.Error()).To(Equal("obtained neither the current working directory nor the executable directory location"))
			})
		})

		Context("When getwd returns directory and osexecutable returns directory", func() {
			BeforeEach(func() {
				p.osGetwd = func() (string, error) {
					return "/foo", nil
				}
				p.osGetExecutable = func() (string, error) {
					return "/bar", nil
				}
			})
			It("Should return expected directory list", func() {
				Expect(dirs).To(Equal([]string{"/foo", "/bar"}))
			})
			It("Should return nil error", func() {
				Expect(getDirErr).To(BeNil())
			})
		})

	})

	Describe("searchFilesForVersion", func() {
		var (
			files   []string
			payload WebspherePayload
			// findStringInFile tasks.FindStringInFileFunc
			// returnSubstring  tasks.ReturnSubstringInFileFunc
		)

		JustBeforeEach(func() {
			payload = p.searchFilesForVersion(files)
		})

		Context("When Websphere version found", func() {
			BeforeEach(func() {
				files = []string{"version.txt"}

				p.findStringInFile = func(string, string) bool {
					return true
				}
				p.returnSubstring = func(string, string) ([]string, error) {
					return []string{"7.1"}, nil
				}
			})
			It("Should return payload version", func() {
				Expect(payload.Version).To(Equal("7.1"))
			})
		})

		Context("When no valid files found", func() {
			BeforeEach(func() {
				p.findStringInFile = func(string, string) bool {
					return false
				}
			})
			It("Should return unknown payload Version", func() {
				Expect(payload.Version).To(Equal("Unknown"))
			})

			Context("When returnsubString returns expected error", func() {
				BeforeEach(func() {
					files = []string{"version.txt"}

					p.findStringInFile = func(string, string) bool {
						return true
					}
					p.returnSubstring = func(string, string) ([]string, error) {
						return []string{}, errors.New("string not found in file")
					}
				})
				It("Should return unknown payload Version", func() {
					Expect(payload.Version).To(Equal("Unknown"))
				})
			})

			Context("When returnsubString returns unexpected error", func() {
				BeforeEach(func() {
					files = []string{"version.txt"}

					p.findStringInFile = func(string, string) bool {
						return true
					}
					p.returnSubstring = func(string, string) ([]string, error) {
						return []string{}, errors.New("totally weird error here man")
					}
				})
				It("Should return unknown payload Version", func() {
					Expect(payload.Version).To(Equal("Unknown"))
				})
			})

		})

	})

	Describe("determineResult", func() {
		var (
			payload WebspherePayload
			result  tasks.Result
		)
		JustBeforeEach(func() {
			result = p.determineResult(payload)
		})

		Context("When payload Version is unknown", func() {
			BeforeEach(func() {
				payload.Version = unknown

			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected Summary", func() {
				Expect(result.Summary).To(Equal("We suspect this is a WebSphere environment but we're unable to determine the version. Supported status is unknown."))
			})
		})
		Context("When there is an error parsing the version", func() {
			BeforeEach(func() {
				payload.Version = "burritos are good for you"
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("couldn't parse version")
				}
			})
			It("Should return error Status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})
			It("Should error Summary", func() {
				Expect(result.Summary).To(Equal("Websphere and its version were found in this environment but we're unable to determine its supported status."))
			})
		})
		Context("When supported payload version", func() {
			BeforeEach(func() {
				payload.Version = "7.0"
				p.versionIsCompatible = tasks.VersionIsCompatible
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return supported Summary", func() {
				Expect(result.Summary).To(Equal("Supported version of Websphere detected in this environment."))
			})
		})
		Context("When unsupported payload version", func() {
			BeforeEach(func() {
				payload.Version = "10.0"
				p.versionIsCompatible = tasks.VersionIsCompatible
			})
			It("Should return failure Status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return failure Summary", func() {
				Expect(result.Summary).To(Equal("Unsupported version of Websphere detected in this environment."))
			})
		})

	})
})
