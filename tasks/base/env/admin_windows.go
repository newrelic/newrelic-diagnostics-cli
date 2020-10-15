package env

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"os/user"

	"github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// types used for dependency injection
type fileOpenFunc func(string) (*os.File, error)
type getEnvVarsFunc func() (tasks.EnvironmentVariables, error)
type getCurrentUser func() (*user.User, error)
type isUserAdmin func(string, string) (bool, error)

// BaseEnvCheckWindowsAdmin - base struct
type BaseEnvCheckWindowsAdmin struct {
	fileOpener     fileOpenFunc
	getEnvVars     getEnvVarsFunc
	getCurrentUser getCurrentUser
	isUserAdmin    isUserAdmin
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvCheckWindowsAdmin) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/CheckWindowsAdmin")
}

// Explain - Returns the help text for each individual task
func (p BaseEnvCheckWindowsAdmin) Explain() string {
	return "Detect if running with Administrator priviliges"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseEnvCheckWindowsAdmin) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p BaseEnvCheckWindowsAdmin) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	var errorsSlice = []error{}

	// 1- check FS - this will fail in docker (this is tested with a unit test)
	isAdminFromFS, fsErr := p.adminFSCheck()
	if isAdminFromFS {
		result.Summary = "NR Diag detected having Elevated permissions"
		result.Status = tasks.Success
		return result
	} else if fsErr != nil {
		errorsSlice = append(errorsSlice, fsErr)
	}

	// 2- check USERNAME env variable - will be Administrator if elevated (this is tested with a unit test)
	isAdminFromUsernameEnvVar, usernameEnvVarErr := p.adminUsernameEnvVarCheck()
	if isAdminFromUsernameEnvVar {
		result.Summary = "NR Diag detected having Elevated permissions"
		result.Status = tasks.Success
		return result
	} else if usernameEnvVarErr != nil {
		errorsSlice = append(errorsSlice, usernameEnvVarErr)
	}

	// 3- check logged in user (this is tested with an integration test)
	isAdminFromLoggedInUser, loggedInUserErr := p.adminLoggedInUserCheck()
	if isAdminFromLoggedInUser {
		result.Summary = "NR Diag detected having Elevated permissions"
		result.Status = tasks.Success
		return result
	} else if loggedInUserErr != nil {
		errorsSlice = append(errorsSlice, loggedInUserErr)
	}

	// not admin
	result.Status = tasks.Warning
	result.Summary = "NR Diag did not detect having Elevated permissions. Some Tasks may fail. If possible re-run from an Admin cmd prompt or PowerShell."
	if len(errorsSlice) > 0 {
		result.Summary += "\n" + strconv.Itoa(len(errorsSlice)) + " errors encountered: "
		for _, e := range errorsSlice {
			result.Summary += "\n - " + e.Error()
		}
	}
	result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
	return result
}

func (p BaseEnvCheckWindowsAdmin) adminLoggedInUserCheck() (bool, error) {
	currentUser, uErr := p.getCurrentUser()
	if uErr != nil {
		return false, uErr
	}
	username, domain := p.getUserAndDomain(currentUser.Username)
	isAdmin, adminErr := p.isUserAdmin(username, domain)
	return isAdmin, adminErr
}

func (p BaseEnvCheckWindowsAdmin) adminUsernameEnvVarCheck() (bool, error) {
	envVars, err := p.getEnvVars()
	if err != nil {
		return false, err
	}
	username := envVars.FindCaseInsensitive("username")
	r := regexp.MustCompile("(?i)" + username)
	isAdmin := r.MatchString("Administrator")
	return isAdmin, nil
}

func (p BaseEnvCheckWindowsAdmin) adminFSCheck() (bool, error) {
	_, err := p.fileOpener("\\\\.\\PHYSICALDRIVE0")
	if err != nil {
		logger.Debug("FS Check failed: ", err.Error())
		return false, err
	}
	return true, nil
}

func (p BaseEnvCheckWindowsAdmin) getUserAndDomain(input string) (string, string) {
	s := strings.Split(input, `\`)
	return s[1], s[0]
}
