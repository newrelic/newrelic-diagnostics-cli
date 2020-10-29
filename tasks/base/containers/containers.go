package containers

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Containers/*")

	// every task in a package needs to be registered to run
	// if you pass "false" as the second parameter it will be runnable, but not included by default
	registrationFunc(BaseContainersDetectDocker{
		executeCommand: tasks.CmdExecutor,
	}, true)

}
