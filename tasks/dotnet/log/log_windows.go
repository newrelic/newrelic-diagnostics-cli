package log

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWinWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Log/*")

	registrationFunc(DotNetLogLevelValidate{}, true)
	registrationFunc(DotNetLogLevelCollect{}, true)

}
