package agent

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Agent/*")

	registrationFunc(DotNetAgentInstalled{ agentInstallPaths: agentInstallPaths }, true)
	registrationFunc(DotNetAgentVersion{}, true)
}
