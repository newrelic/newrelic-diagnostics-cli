package agentcontrol

import (
	"fmt"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sACLogs - This struct defined the sample plugin which can be used as a starting point
type K8sAgentControlLogs struct {
	cmdExec       tasks.CmdExecFunc
	appName       string
	labelSelector string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sAgentControlLogs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString(fmt.Sprintf("K8s/AgentControl/%s-logs", p.appName))
}

// Explain - Returns the help text for each individual task
func (p K8sAgentControlLogs) Explain() string {
	return "Collects agent-control " + p.appName + " pod logs"
}

// Dependencies - Returns the dependencies for each task.
func (p K8sAgentControlLogs) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sAgentControlLogs) Execute(options tasks.Options, _ map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["k8sNamespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving logs: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s agent-control " + p.appName + " pod logs",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: fmt.Sprintf("%s.log", p.appName), Stream: stream}},
	}
}

func (p K8sAgentControlLogs) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"logs",
			"-l",
			p.labelSelector,
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
		p.labelSelector,
		"--all-containers",
		"--prefix",
	)
}
