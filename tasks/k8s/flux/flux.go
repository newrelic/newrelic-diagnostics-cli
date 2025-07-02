package flux

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const (
	kubectlBin = "kubectl"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering K8s/Flux/*")
	registrationFunc(FluxCharts{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(FluxReleases{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(FluxRepositories{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
