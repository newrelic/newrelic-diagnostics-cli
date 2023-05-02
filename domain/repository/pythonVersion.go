package repository

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type IPythonEnvVersion interface {
	CheckPythonVersion(pythonCmd string) tasks.Result
}

type IPipEnvVersion interface {
	CheckPipVersion(pipCmd string) tasks.Result
}
