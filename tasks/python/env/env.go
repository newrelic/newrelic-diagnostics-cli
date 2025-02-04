package env

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/python/repository"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Python/Env/*")
	pythonEnv := repository.PythonEnv{CmdExec: tasks.CmdExecutor}
	pipEnv := repository.PipEnv{}
	registrationFunc(PythonEnvVersion{
		iPythonEnvVersion: pythonEnv},
		true)
	registrationFunc(PythonEnvDependencies{
		iPipEnvVersion: pipEnv,
	}, true)
}
