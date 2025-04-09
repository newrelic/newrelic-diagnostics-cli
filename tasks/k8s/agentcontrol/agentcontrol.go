package agentcontrol

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering K8s/AgentControl/*")
	registrationFunc(K8sAgentControlLogs{
		cmdExec:       tasks.CmdExecutor,
		appName:       "helm-controller",
		labelSelector: "app=helm-controller",
	}, true)
	registrationFunc(K8sAgentControlLogs{
		cmdExec:       tasks.CmdExecutor,
		appName:       "source-controller",
		labelSelector: "app=source-controller",
	}, true)
	registrationFunc(K8sAgentControlLogs{
		cmdExec:       tasks.CmdExecutor,
		appName:       "agent-control",
		labelSelector: "app.kubernetes.io/name=agent-control",
	}, true)
	registrationFunc(K8sAgentControlStatusServer{
		cmdExec:       tasks.CmdExecutor,
		appName:       "agent-control",
		labelSelector: "app.kubernetes.io/name=agent-control",
	}, true)
}
