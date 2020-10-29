package env

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Ruby/Env/*")
	registrationFunc(RubyEnvProcess{}, true)
	registrationFunc(RubyEnvVersion{
		cmdExecutor: tasks.CmdExecutor,
	}, true)

}
