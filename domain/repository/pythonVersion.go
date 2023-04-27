package repository

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type IPythonEnvVersion interface {
	CheckPythonVersion(pythonCmd string) tasks.Result
}
