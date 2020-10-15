package env

import (
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// PythonEnvVersion - The struct defines the Python version.
type PythonEnvVersion struct {
	cmdExec tasks.CmdExecFunc
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
	return p.checkPythonVersion()
}

func (p PythonEnvVersion) checkPythonVersion() (result tasks.Result) {

	versionRaw, cmdBuildErr := p.cmdExec("python", "--version")

	if cmdBuildErr != nil {
		result.Status = tasks.Error
		result.Summary = "Unable to execute command: $ python --version. Error: " + cmdBuildErr.Error()
		result.URL = "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic"
		return
	}

	versionString, isValid := parsePythonVersion(versionRaw)
	if !isValid {
		result.Status = tasks.Error
		result.Summary = "Unable to detect installed Python version. "
		result.URL = "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic"
		return
	}

	result.Summary = versionString
	result.Status = tasks.Info
	result.Payload = versionString

	return
}

func parsePythonVersion(versionRaw []byte) (string, bool) {
	versionString := strings.TrimSpace(string(versionRaw))
	if strings.HasPrefix(versionString, "Python ") {
		versionString = strings.TrimPrefix(versionString, "Python ")
		log.Debug("Python version found. Version is: " + versionString)
		return versionString, true
	}

	return "", false
}
