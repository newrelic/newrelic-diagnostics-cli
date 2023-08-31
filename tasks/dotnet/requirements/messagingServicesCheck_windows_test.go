package requirements

import (
	"errors"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Dotnet/Requirements/MessagingServicesCheck", func() {
	var p DotnetRequirementsMessagingServicesCheck

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
					"DotNet/Agent/Installed": {
						Status: tasks.Failure,
					},
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal(tasks.UpstreamFailedSummary + "DotNet/Agent/Installed"))
			})
		})
		Context("when list of directories is empty", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{}
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Unable to determine the " + tasks.ThisProgramFullName + " working and executable directory paths."))
			})
		})
		Context("when no Dlls found", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{}
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.None))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Did not find any dlls associated with Messaging Services supported by the .Net Agent"))
			})
		})
		Context("when unsupported version of NServiceVer detected", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/NServiceBus.Core.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "4.0", nil
				}
				p.versionIsCompatible = func(version string, _ []string) (bool, error) {
					return false, nil
				}
			})

			It("should return an expected failure result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Incompatible Messaging Services detected, see output.json for more details. Detected the following Messaging Services as incompatible: \nIncompatible version of /foo/NServiceBus.Core.dll detected. Found version 4.0\n"))
			})
		})
		Context("when unsupported version of rabbit MQ detected", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/RabbitMQ.Client.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "2.0", nil
				}
				p.versionIsCompatible = func(version string, _ []string) (bool, error) {
					return false, nil
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Incompatible Messaging Services detected, see output.json for more details. Detected the following Messaging Services as incompatible: \nIncompatible version of /foo/RabbitMQ.Client.dll detected. Found version 2.0\n"))
			})
		})
		Context("when supported nService but unsupported Rabbit MQ ", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/RabbitMQ.Client.dll", "/bar/NServiceBus.Core.dll"}
				}
				p.getFileVersion = func(path string) (string, error) {
					if strings.Contains(path, "Rabbit") {
						return "2.0", nil
					}
					return "5.0", nil
				}
				p.versionIsCompatible = func(version string, _ []string) (bool, error) {
					if version == "5.0" {
						return true, nil
					}
					return false, nil
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Failure))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Incompatible Messaging Services detected, see output.json for more details. Detected the following Messaging Services as incompatible: \nIncompatible version of /foo/RabbitMQ.Client.dll detected. Found version 2.0\n"))
			})
		})
		Context("when supported nService and supported Rabbit MQ ", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/RabbitMQ.Client.dll", "/bar/NServiceBus.Core.dll"}
				}
				p.getFileVersion = func(path string) (string, error) {
					if strings.Contains(path, "Rabbit") {
						return "4.0", nil
					}
					return "5.0", nil
				}
				p.versionIsCompatible = func(version string, _ []string) (bool, error) {
					return true, nil
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("All Messaging Services detected as compatible, see output.json for more details."))
			})
		})
		Context("when supported nService, supported Rabbit MQ and supported system messaging detected", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/foo/RabbitMQ.Client.dll", "/bar/NServiceBus.Core.dll", "/bar/System.Messaging.dll"}
				}
				p.getFileVersion = func(path string) (string, error) {
					if strings.Contains(path, "Rabbit") {
						return "4.0", nil
					}
					return "5.0", nil
				}
				p.versionIsCompatible = func(version string, _ []string) (bool, error) {
					return true, nil
				}
			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("All Messaging Services detected as compatible, see output.json for more details."))
			})
		})
		Context("when error processing SystemMessaging DLL", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/bar/System.Messaging.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "", errors.New("tacos are the best")
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("tacos are the best")
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error parsing SystemMessaging DLL version. Error was: tacos are the best"))
			})
		})
		Context("when error processing RabbitMQ DLL", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"RabbitMQ.Client.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "4.0", nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("burritos are better")
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error parsing RabbitMQ DLL version. Error was: burritos are better"))
			})
		})
		Context("when error processing NServiceBus DLL", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/bar/NServiceBus.Core.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "5.0", nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("tostadas are supreme")
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Error))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Error parsing NServiceBus DLL version. Error was: tostadas are supreme"))
			})
		})
		Context("when version couldn't be determined for some messagingService", func() {

			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status: tasks.Success,
					},
				}
				p.getWorkingDirectories = func() []string {
					return []string{"/foo"}
				}
				p.findFiles = func([]string, []string) []string {
					return []string{"/bar/NServiceBus.Core.dll"}
				}
				p.getFileVersion = func(string) (string, error) {
					return "", nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, nil
				}

			})

			It("should return an expected result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})

			It("should return an expected result summary", func() {
				Expect(result.Summary).To(Equal("Couldn't get version of some Messaging Services dlls: \nNo version of /bar/NServiceBus.Core.dll detected. Found empty version\n"))
			})
		})
	})

	Describe("checkNServiceBus()", func() {
		var (
			path                     string
			result                   messagingService
			resultErr                error
			expectedMessagingService messagingService
		)

		JustBeforeEach(func() {
			path = ""
			result, resultErr = p.checkNServiceBus(path)
		})

		Context("when version is not compatible", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "6.0.0.0", nil
				}
				p.versionIsCompatible = tasks.VersionIsCompatible
				expectedMessagingService = messagingService{
					name:        "",
					version:     "6.0.0.0",
					installed:   true,
					versionGood: false,
				}
			})

			It("should return an expected invalid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("when version is compatible", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "5.1.0.0", nil
				}
				p.versionIsCompatible = tasks.VersionIsCompatible
				expectedMessagingService = messagingService{
					name:        "",
					version:     "5.1.0.0",
					installed:   true,
					versionGood: true,
				}
			})

			It("should return an expected valid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("when version compatible returns an error", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "", nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("i like Bananas")
				}
				expectedMessagingService = messagingService{
					name:        "",
					version:     "",
					installed:   true,
					versionGood: false,
				}
			})

			It("should return an expected invalid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a expected error", func() {
				Expect(resultErr.Error()).To(Equal("i like Bananas"))
			})
		})
	})

	Describe("checkRabbitMq()", func() {
		var (
			path                     string
			result                   messagingService
			resultErr                error
			expectedMessagingService messagingService
		)

		JustBeforeEach(func() {
			path = ""
			result, resultErr = p.checkRabbitMq(path)
		})

		Context("when version is not compatible", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "3.4.0.0", nil
				}
				p.versionIsCompatible = tasks.VersionIsCompatible
				expectedMessagingService = messagingService{
					name:        "",
					version:     "3.4.0.0",
					installed:   true,
					versionGood: false,
				}
			})

			It("should return an expected invalid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("when version is compatible", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "4.1.0.0", nil
				}
				p.versionIsCompatible = tasks.VersionIsCompatible
				expectedMessagingService = messagingService{
					name:        "",
					version:     "4.1.0.0",
					installed:   true,
					versionGood: true,
				}
			})

			It("should return an expected valid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("when version compatible returns an error", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "", nil
				}
				p.versionIsCompatible = func(string, []string) (bool, error) {
					return false, errors.New("i like Bananas")
				}
				expectedMessagingService = messagingService{
					name:        "",
					version:     "",
					installed:   true,
					versionGood: false,
				}
			})

			It("should return an expected invalid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a expected error", func() {
				Expect(resultErr.Error()).To(Equal("i like Bananas"))
			})
		})
	})

	Describe("checkSystemMessaging()", func() {
		var (
			path                     string
			result                   messagingService
			resultErr                error
			expectedMessagingService messagingService
		)

		JustBeforeEach(func() {
			path = ""
			result, resultErr = p.checkSystemMessaging(path)
		})

		Context("when version is compatible", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "5.1.0.0", nil
				}
				expectedMessagingService = messagingService{
					name:        "",
					version:     "5.1.0.0",
					installed:   true,
					versionGood: true,
				}
			})

			It("should return an expected valid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a nil error", func() {
				Expect(resultErr).To(BeNil())
			})
		})
		Context("when version compatible returns an error", func() {

			BeforeEach(func() {
				p.getFileVersion = func(string) (string, error) {
					return "", errors.New("i like Bananas")
				}
				expectedMessagingService = messagingService{
					name:        "",
					version:     "",
					installed:   true,
					versionGood: false,
				}
			})

			It("should return an expected invalid messagingService", func() {
				Expect(result).To(Equal(expectedMessagingService))
			})

			It("should return a expected error", func() {
				Expect(resultErr.Error()).To(Equal("i like Bananas"))
			})
		})
	})

})
