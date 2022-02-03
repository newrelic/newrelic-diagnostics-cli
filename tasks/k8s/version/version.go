package version

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering k8s/version/*")
	registrationFunc(K8sVersion{
		cmdExec: tasks.CmdExecutor,
	}, true)
}

// K8sVersion - This struct defined the sample plugin which can be used as a starting point
type K8sVersion struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Version/Version")
}

// Explain - Returns the help text for each individual task
func (p K8sVersion) Explain() string {
	return "This task retrieves the version of the client and of the cluster."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sVersion) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	res, err := p.cmdExec(kubectlBin, "version")

	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving version: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	return tasks.Result{
		Summary: "kubectl and cluster version: " + string(res),
		Status:  tasks.Info,
		Payload: string(res),
	}
}
