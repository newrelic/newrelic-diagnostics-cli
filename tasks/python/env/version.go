package env

import (
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/repository"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// PythonEnvVersion - The struct defines the Python version.

type PythonEnvVersion struct {
	iPythonEnvVersion repository.IPythonEnvVersion
}

// Identifier - This returns the Category, Subcategory and Name of this task.
func (p PythonEnvVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Env/Version")
}

// Explain - Returns the help text for the Python/Env/Version task.
func (p PythonEnvVersion) Explain() string {
	return "Determine Python version"
}

// Dependencies - Returns the dependencies for this task.
func (p PythonEnvVersion) Dependencies() []string {
	return []string{
		"Python/Config/Agent",
	}
}

// Execute - The core work within this task.
func (p PythonEnvVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	if upstream["Python/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Python Agent not installed. This task didn't run."
		return result
	}
	// pythonDeps := new(PythonDeps)
	return p.RunPythonCommands()

}

func (p PythonEnvVersion) RunPythonCommands() tasks.Result {
	var errorsToReturn []string
	var successesToReturn []string
	result_1 := p.iPythonEnvVersion.CheckPythonVersion("python")
	if result_1.Status == tasks.Error {
		errorsToReturn = append(errorsToReturn, result_1.Summary)
	} else {
		successesToReturn = append(successesToReturn, result_1.Summary)
	}
	result_2 := p.iPythonEnvVersion.CheckPythonVersion("python3")
	if result_2.Status == tasks.Error {
		errorsToReturn = append(errorsToReturn, result_2.Summary)
	} else {
		successesToReturn = append(successesToReturn, result_2.Summary)
	}
	errorStr := strings.Join(errorsToReturn, "\n")
	successStr := strings.Join(successesToReturn, "\n")
	if len(errorsToReturn) == 2 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: errorStr,
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	} else if len(errorsToReturn) > 0 {

		return tasks.Result{
			Status:  tasks.Warning,
			Summary: errorStr + "\n" + successStr,
			Payload: successStr,
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}
	return tasks.Result{
		Status:  tasks.Success,
		Summary: successStr,
		Payload: successesToReturn,
	}
}
