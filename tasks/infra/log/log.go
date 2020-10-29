package log

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Log/*")
	registrationFunc(InfraLogLevelCheck{}, true)
	registrationFunc(InfraLogCollect{
		validatePaths: tasks.ValidatePaths,
	}, true)
}
