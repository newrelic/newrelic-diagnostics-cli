package log

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Log/*")

	registrationFunc(BaseLogCollect{}, false)
	registrationFunc(BaseLogCopy{}, true)
	registrationFunc(BaseLogReportingTo{}, true)
}
