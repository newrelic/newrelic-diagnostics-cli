package env

import (
	"errors"
	"os"
	"os/user"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	fileOpenErrorString    = "Mock file open error"
	getEnvVarsErrorString  = "Mock GetEnvVars error"
	isUserAdminErrorString = "Mock IsUserAdmin failure"
)

func mockFileOpenerFailure(name string) (*os.File, error) {
	return nil, errors.New(fileOpenErrorString)
}

func mockFileOpenerSuccess(name string) (*os.File, error) {
	return nil, nil
}

func mockGetEnvVarsSuccess() (tasks.EnvironmentVariables, error) {
	envVars := make(map[string]string)
	envVars["USERNAME"] = "Administrator"
	return tasks.EnvironmentVariables{
		All: envVars,
	}, nil
}

func mockGetEnvVarsUsernameNotAdministrator() (tasks.EnvironmentVariables, error) {
	envVars := make(map[string]string)
	envVars["USERNAME"] = "User1"
	return tasks.EnvironmentVariables{
		All: envVars,
	}, nil
}

func mockGetEnvVarsWithError() (tasks.EnvironmentVariables, error) {
	return tasks.EnvironmentVariables{}, errors.New(getEnvVarsErrorString)
}

func mockGetCurrentUser() (*user.User, error) {
	return &user.User{
		Username: `Domain\MockUser`,
	}, nil
}

func mockIsUserAdminSuccess(username string, domain string) (bool, error) {
	return true, nil
}

func mockIsUserAdminFailure(username string, domain string) (bool, error) {
	return false, errors.New(isUserAdminErrorString)
}

func mockIsUserAdminNotAdmin(username string, domain string) (bool, error) {
	return false, nil
}

var _ = Describe("Base/Env/CheckWindowsAdmin", func() {
	var p BaseEnvCheckWindowsAdmin //instance of our task struct to be used in tests

	Describe("Execute()", func() {
		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("when the file is opened successfully", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerSuccess
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

		})

		Context("when there is an error opening the file and USERNAME is set to Administrator", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerFailure
				p.getEnvVars = mockGetEnvVarsSuccess
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

		})

		Context("when there is an error opening the file and USERNAME is not Administrator but logged in user is an Admin", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerFailure
				p.getEnvVars = mockGetEnvVarsUsernameNotAdministrator
				p.getCurrentUser = mockGetCurrentUser
				p.isUserAdmin = mockIsUserAdminSuccess
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

		})

		Context("when there is an error opening the file and an error getting env vars and logged in user is an Admin", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerFailure
				p.getEnvVars = mockGetEnvVarsWithError
				p.getCurrentUser = mockGetCurrentUser
				p.isUserAdmin = mockIsUserAdminSuccess
			})

			It("should return an expected success result status", func() {
				Expect(result.Status).To(Equal(tasks.Success))
			})

		})

		Context("when there is an error opening the file and an error getting env vars and logged in user is not an Admin", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerFailure
				p.getEnvVars = mockGetEnvVarsWithError
				p.getCurrentUser = mockGetCurrentUser
				p.isUserAdmin = mockIsUserAdminNotAdmin
			})

			It("should return an expected warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It(`should return an expected error "`+fileOpenErrorString+`"`, func() {
				Expect(result.Summary).To(ContainSubstring(fileOpenErrorString))
			})
			It(`should return an expected error "`+getEnvVarsErrorString+`"`, func() {
				Expect(result.Summary).To(ContainSubstring(getEnvVarsErrorString))
			})

		})

		Context("when there is an error opening the file and an error getting env vars and an error checking current user permissions", func() {
			BeforeEach(func() {
				p.fileOpener = mockFileOpenerFailure
				p.getEnvVars = mockGetEnvVarsWithError
				p.getCurrentUser = mockGetCurrentUser
				p.isUserAdmin = mockIsUserAdminFailure
			})

			It("should return an expected warning result status", func() {
				Expect(result.Status).To(Equal(tasks.Warning))
			})
			It(`should return an expected error "`+fileOpenErrorString+`"`, func() {
				Expect(result.Summary).To(ContainSubstring(fileOpenErrorString))
			})
			It(`should return an expected error "`+getEnvVarsErrorString+`"`, func() {
				Expect(result.Summary).To(ContainSubstring(getEnvVarsErrorString))
			})
			It(`should return an expected error "`+isUserAdminErrorString+`"`, func() {
				Expect(result.Summary).To(ContainSubstring(isUserAdminErrorString))
			})

		})
	})
})
