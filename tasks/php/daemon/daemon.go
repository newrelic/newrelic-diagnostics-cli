package daemon

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering PHP/Daemon*")

	registrationFunc(PHPDaemonRunning{
		processFinder:     tasks.FindProcessByName,
		fileExistsChecker: tasks.FileExists,
	}, true)
}
