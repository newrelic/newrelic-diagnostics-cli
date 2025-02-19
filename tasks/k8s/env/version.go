package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sVersion - This struct defined the sample plugin which can be used as a starting point
type K8sVersion struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Env/Version")
}

// Explain - Returns the help text for each individual task
func (p K8sVersion) Explain() string {
	return "Retrieves the version of the kubectl client and of the cluster."
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

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "kubectl version output successfully collected",
		Status:      tasks.Info,
		Payload:     string(res),
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "kubectlVersion.txt", Stream: stream}},
	}
}
