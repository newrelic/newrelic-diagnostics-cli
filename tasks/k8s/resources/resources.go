package resources

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering K8s/Resources/*")
	registrationFunc(K8sConfigs{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sDeployment{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sDaemonset{
		cmdExec: tasks.CmdExecutor,
	}, true)
	registrationFunc(K8sPods{
		cmdExec: tasks.CmdExecutor,
	}, true)
}

func getResources(
	options tasks.Options,
	runCommand func(namespace string) ([]byte, error),
) ([]byte, error) {
	namespace := options.Options["k8sNamespace"]
	namespaces := []string{namespace}

	agentsNamespace := options.Options["ACAgentsNamespace"]
	if agentsNamespace != "" && agentsNamespace != namespace {
		namespaces = append(namespaces, agentsNamespace)
	}

	var result []byte
	for _, ns := range namespaces {
		res, err := runCommand(ns)
		if err != nil {
			return nil, err
		}
		result = append(result, res...)
	}

	return result, nil
}
