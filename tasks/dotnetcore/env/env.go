package env

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNetCore/Env/*")

	registrationFunc(DotNetCoreEnvProcess{}, true)
	registrationFunc(DotNetCoreEnvVersions{}, true)
}
