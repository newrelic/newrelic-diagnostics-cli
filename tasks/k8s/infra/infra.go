package infra

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering K8s/Infra/*")
	registrationFunc(K8sConfigs{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sDeployment{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sLogs{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sDaemonset{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sPods{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
