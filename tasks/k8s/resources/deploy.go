package resources

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sDeployment - This struct defined the sample plugin which can be used as a starting point
type K8sDeployment struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sDeployment) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Resources/Deploy")
}

// Explain - Returns the help text for each individual task
func (p K8sDeployment) Explain() string {
	return "Collects K8s deployments information for the given namespace in YAML format."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sDeployment) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sDeployment) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	stream := make(chan string)

	result, err := getResources(options, p.runCommand)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving deployments details: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	go tasks.StreamBlob(string(result), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s newrelic-infrastructure deployment",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "k8sDeployment.txt", Stream: stream}},
	}
}

func (p K8sDeployment) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"describe",
			"deployment",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"describe",
		"deployment",
		"-n",
		namespace,
	)
}
