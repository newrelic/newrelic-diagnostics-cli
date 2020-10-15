package env

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Python/Env/*")
	registrationFunc(PythonEnvVersion{cmdExec: tasks.CmdExecutor}, true)
	registrationFunc(PythonEnvDependencies{}, true)
}
