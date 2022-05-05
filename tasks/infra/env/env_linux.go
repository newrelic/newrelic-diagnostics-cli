package env

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterLinuxWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Env/*")

	registrationFunc(InfraEnvValidateZookeeperPath{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
