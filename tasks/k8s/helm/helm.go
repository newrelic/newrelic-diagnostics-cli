package helm

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const (
	helmBin = "helm"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering K8s/Helm/*")
	registrationFunc(HelmReleases{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
