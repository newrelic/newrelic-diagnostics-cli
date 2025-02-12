package infra

import (
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sLogs - This struct defined the sample plugin which can be used as a starting point
type K8sLogs struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sLogs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Infra/Logs")
}

// Explain - Returns the help text for each individual task
func (p K8sLogs) Explain() string {
	return "Collects newrelic-infrastructure pod logs."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sLogs) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sLogs) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
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
			Summary: "Error retrieving logs: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	if namespace == "" &&
		(strings.TrimSpace(string(res)) == "No resources found in default namespace." ||
			strings.TrimSpace(string(res)) == "apiVersion: v1\nitems: []\nkind: List\nmetadata:\n  resourceVersion: \"\"") {
		res, err = p.runCommand("newrelic")
		if err != nil {
			return tasks.Result{
				Summary: "Error retrieving logs: " + err.Error(),
				Status:  tasks.Error,
			}
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s newrelic-infrastructure logs",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "newrelic-infrastructure.log", Stream: stream}},
	}
}

func (p K8sLogs) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"logs",
			"-l",
			"app.kubernetes.io/name=newrelic-infrastructure",
			"--all-containers",
			"--prefix",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"logs",
		"-n",
		namespace,
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
		"--all-containers",
		"--prefix",
	)
}
