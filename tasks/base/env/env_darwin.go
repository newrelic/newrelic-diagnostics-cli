package env

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterDarwinWith - will register any plugins in this package
func RegisterDarwinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Env/*")

	registrationFunc(BaseEnvRootUser{
		isUserRoot: tasks.IsUserRoot,
	}, true)
}
