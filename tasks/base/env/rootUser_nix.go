//go:build linux || darwin
// +build linux darwin

package env

import (
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseEnvRootUser - base struct
type BaseEnvRootUser struct {
	isUserRoot func() (bool, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvRootUser) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/RootUser")
}

// Explain - Returns the help text for each individual task
func (p BaseEnvRootUser) Explain() string {
	return "Detect if running with root permissions"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseEnvRootUser) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p BaseEnvRootUser) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	isAdmin, err := p.isUserRoot()
	if err != nil {
		result.Summary = "Encountered an error while determining if " + tasks.ThisProgramFullName + " has root permissions."
		result.Status = tasks.Error
		return result
	}
	if !isAdmin {
		result.Status = tasks.Warning
		result.Summary = "Root permissions not detected. If you see any permissions errors, consider re-running " + tasks.ThisProgramFullName + " using 'sudo -E'. The '-E' option will help preserve the environment variables needed for running this program."
		if runtime.GOOS == "darwin" {
			result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/solve-common-issues/diagnostics-cli-nrdiag/run-diagnostics-cli-nrdiag/#macos-run"
		} else {
			result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/solve-common-issues/diagnostics-cli-nrdiag/run-diagnostics-cli-nrdiag/#linux-run"
		}
		return result
	}

	result.Summary = "Root permissions detected"
	result.Status = tasks.Success
	return result
}
