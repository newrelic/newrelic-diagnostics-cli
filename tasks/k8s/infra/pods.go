package infra

import (
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sPods - This struct defined the sample plugin which can be used as a starting point
type K8sPods struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sPods) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Infra/Pods")
}

// Explain - Returns the help text for each individual task
func (p K8sPods) Explain() string {
	return "Collects newrelic-infrastructure pod information."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sPods) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sPods) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["namespace"]
	if namespace != "" {
		res, err = p.runCommand(namespace)
	} else {
		res, err = p.runCommand("")
	}

	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving pods: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	if namespace == "" &&
		(strings.TrimSpace(string(res)) == "No resources found in default namespace." ||
			strings.TrimSpace(string(res)) == "apiVersion: v1\nitems: []\nkind: List\nmetadata:\n  resourceVersion: \"\"") {
		res, err = p.runCommand("newrelic")
		if err != nil {
			return tasks.Result{
				Summary: "Error retrieving pods: " + err.Error(),
				Status:  tasks.Error,
			}
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s newrelic-infrastructure pods",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "k8sInfraPods.txt", Stream: stream}},
	}
}

func (p K8sPods) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"get",
			"pods",
			"-l",
			"app.kubernetes.io/name=newrelic-infrastructure",
			"-o",
			"yaml",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"get",
		"pods",
		"-n",
		namespace,
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
		"-o",
		"yaml",
	)
}
