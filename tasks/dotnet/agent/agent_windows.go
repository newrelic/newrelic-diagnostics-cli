package agent

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Agent/*")

	registrationFunc(DotNetAgentInstalled{agentInstallPaths: agentInstallPaths}, true)
	registrationFunc(DotNetAgentVersion{}, true)
}
