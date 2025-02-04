package requirements

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/MessagingServicesCheck", func() {
	var p DotnetRequirementsDatastores

	Describe("Identify()", func() {
		It("Should return Identity object", func() {
			Expect(p.Identifier()).To(Equal(tasks.Identifier{Name: "Datastores", Category: "DotNet", Subcategory: "Requirements"}))
		})
	})

	Describe("Explain()", func() {
		It("Should return Explain test", func() {
			Expect(p.Explain()).To(Equal("Check database version compatibility with New Relic .NET agent"))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return Dependencies list", func() {
			Expect(p.Dependencies()).To(Equal([]string{"DotNet/Agent/Installed"}))
		})
	})

	Describe("checkMongoVer()", func() {
		var (
			version   string
			supported bool
		)
		JustBeforeEach(func() {
			supported = checkMongoVer(version)
		})

		Context("With supported version 1.10", func() {
			BeforeEach(func() {
				version = "1.10"
			})
			It("Should return supported", func() {
				Expect(supported).To(BeTrue())
			})
		})
		Context("With supported version 1.6", func() {
			BeforeEach(func() {
				version = "1.6"
			})
			It("Should return supported", func() {
				Expect(supported).To(BeTrue())
			})
		})
		Context("With unsupported version 1.11", func() {
			BeforeEach(func() {
				version = "1.11"
			})
			It("Should return unsupported", func() {
				Expect(supported).To(BeFalse())
			})
		})
		Context("With supported version 2.3", func() {
			BeforeEach(func() {
				version = "2.3"
			})
			It("Should return supported", func() {
				Expect(supported).To(BeTrue())
			})
		})
		Context("With unsupported version 2.8", func() {
			BeforeEach(func() {
				version = "2.8"
			})
			It("Should return unsupported", func() {
				Expect(supported).To(BeFalse())
			})
		})

	})

	Describe("checkCouchBaseVer()", func() {
		var (
			version   string
			supported bool
		)
		JustBeforeEach(func() {
			supported = checkCouchBaseVer(version)
		})

		Context("With unsupported version 1", func() {
			BeforeEach(func() {
				version = "1"
			})
			It("Should return unsupported", func() {
				Expect(supported).To(BeFalse())
			})
		})

		Context("With supported version 2", func() {
			BeforeEach(func() {
				version = "2"
			})
			It("Should return supported", func() {
				Expect(supported).To(BeTrue())
			})
		})

		Context("With supported version 3", func() {
			BeforeEach(func() {
				version = "3"
			})
			It("Should return supported", func() {
				Expect(supported).To(BeTrue())
			})
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

		Context("When no dlls are found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}
			})
			It("Should return None status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("Did not find any dlls associated with Datastores supported by the .Net Agent"))
			})
		})

		Context("When all detected datastores are compatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/IBM.Data.DB2.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "1.2", nil
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("All datastores detected as compatible, see plugin.json for more details."))
			})
		})

		Context("When multiple detected datastores are compatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/MongoDB.Driver.dll", "/bar/Oracle.ManagedDataAccess.dll"}
				}
				p.getFileVersion = func(input string) (string, error) {
					if input == "/foo/MongoDB.Driver.dll" {
						return "1.2", nil
					} else if input == "/bar/Oracle.ManagedDataAccess.dll" {
						return "2.2", nil
					}
					// Fall through, mock received unexpected input. (Shouldn't happen)
					return "", errors.New("red Alert, Red Alert")
				}
			})
			It("Should return Success status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})
			It("Should return expected summary", func() {
				Expect(result.Summary).To(Equal("All datastores detected as compatible, see plugin.json for more details."))
			})
			It("Should return expected payload", func() {
				expectedPayload := []string{
					"Found MongoDB.Driver.dll with version 1.2",
					"Found Oracle.ManagedDataAccess.dll with version 2.2",
				}
				Expect(result.Payload).To(Equal(expectedPayload))
			})

		})

		Context("When all detected datastores are incompatible", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/MongoDB.Driver.dll", "/bar/CouchbaseNetClient.dll"}
				}
				p.getFileVersion = func(input string) (string, error) {
					if input == "/foo/MongoDB.Driver.dll" {
						return "2.11", nil
					} else if input == "/bar/CouchbaseNetClient.dll" {
						return "1.0", nil
					}
					// Fall through, mock received unexpected input. (Shouldn't happen)
					return "", errors.New("danger Danger Will Robinson")
				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				expectedSummary := "Incompatible datastores detected, see plugin.json for more details. Detected the following datastores as incompatible: \n" +
					"Incompatible version of MongoDB.Driver.dll detected. Found version 2.11\n" +
					"Incompatible version of CouchbaseNetClient.dll detected. Found version 1.0\n"
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})

		Context("When mix of unsupported and supported datastores found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}

				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/MongoDB.Driver.dll", "/bar/Oracle.ManagedDataAccess.dll"}
				}
				p.getFileVersion = func(input string) (string, error) {
					if input == "/foo/MongoDB.Driver.dll" {
						return "2.11", nil
					} else if input == "/bar/Oracle.ManagedDataAccess.dll" {
						return "2.2", nil
					}
					// Fall through, mock received unexpected input. (Shouldn't happen)
					return "", errors.New("what are you doing, Dave?")

				}
			})
			It("Should return Failure status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})
			It("Should return expected summary", func() {
				expectedSummary := "Incompatible datastores detected, see plugin.json for more details. Detected the following datastores as incompatible: \n" +
					"Incompatible version of MongoDB.Driver.dll detected. Found version 2.11\n"
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})

		Context("When version could not be found for dll", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {Status: tasks.Success},
				}

				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/MongoDB.Driver.dll"}
				}
				p.getFileVersion = func(input string) (string, error) {
					return "", errors.New("unable to determine version")
				}
			})
			It("Should return Warning status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It("Should return expected summary", func() {
				expectedSummary := "Couldn't get version of some datastore dlls: \n" +
					"Unable to find version for MongoDB.Driver.dll\n"
				Expect(result.Summary).To(Equal(expectedSummary))
			})
		})

	})
})
