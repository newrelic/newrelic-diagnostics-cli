package repository

import (
	"fmt"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type PythonEnv struct {
	CmdExec tasks.CmdExecFunc
}

func (p PythonEnv) CheckPythonVersion(pythonCmd string) (result tasks.Result) {

	versionRaw, cmdBuildErr := p.CmdExec(pythonCmd, "--version")

	if cmdBuildErr != nil {
		result.Status = tasks.Error
		summary := fmt.Sprintf("Unable to execute command: $ %s --version. Error: ", pythonCmd)
		result.Summary = summary + cmdBuildErr.Error()
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

	summary := fmt.Sprintf("Python %s found in environment using $ %s --version.", versionString, pythonCmd)
	result.Summary = summary
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
