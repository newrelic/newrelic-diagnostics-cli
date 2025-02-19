package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	registrationFunc(K8sVersion{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
