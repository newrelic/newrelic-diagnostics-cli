package infra

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sDaemonset - This struct defined the sample plugin which can be used as a starting point
type K8sDaemonset struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sDaemonset) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Infra/Daemonset")
}

// Explain - Returns the help text for each individual task
func (p K8sDaemonset) Explain() string {
	return "Collects newrelic-infrastructure daemonset information."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sDaemonset) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sDaemonset) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["namespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving daemonset details: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s newrelic-infrastructure daemonset",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "k8sInfraDaemonset.txt", Stream: stream}},
	}
}

func (p K8sDaemonset) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"describe",
			"daemonset",
			"-l",
			"app.kubernetes.io/name=newrelic-infrastructure",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"describe",
		"daemonset",
		"-n",
		namespace,
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
	)
}
