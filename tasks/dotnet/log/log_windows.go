package log

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWinWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Log/*")

	registrationFunc(DotNetLogLevelValidate{}, true)
	registrationFunc(DotNetLogLevelCollect{}, true)

}
