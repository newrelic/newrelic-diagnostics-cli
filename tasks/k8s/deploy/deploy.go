package deploy

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering k8s/deploy/*")
	registrationFunc(K8sDeployment{
		cmdExec: tasks.CmdExecutor,
	}, true)
}

// K8sDeployment - This struct defined the sample plugin which can be used as a starting point
type K8sDeployment struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sDeployment) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Deploy/Deploy")
}

// Explain - Returns the help text for each individual task
func (p K8sDeployment) Explain() string {
	return "This task collects describe of deployments."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sDeployment) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sDeployment) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	res, err := p.cmdExec(
		kubectlBin,
		"describe",
		"deployment",
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
	)

	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving config: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	ch := make(chan string)
	go func() {
		ch <- string(res)
		close(ch)
	}()

	var filesToCopy []tasks.FileCopyEnvelope
	filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: "forNowAllDeployDescribe", Stream: ch, Identifier: p.Identifier().String()})

	return tasks.Result{
		Summary:     "All describe for deploys",
		Status:      tasks.Info,
		Payload:     string(res),
		FilesToCopy: filesToCopy,
	}
}
