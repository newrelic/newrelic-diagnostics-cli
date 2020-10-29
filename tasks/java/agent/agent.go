package agent

import (
	"os"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Java/Agent/*")

	registrationFunc(JavaAgentVersion{
		wdGetter:     os.Getwd,
		cmdExec:      tasks.CmdExecutor,
		findTheFiles: tasks.FindFiles,
	}, true)
}
