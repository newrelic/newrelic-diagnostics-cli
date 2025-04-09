package agentcontrol

import (
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sAgentControlStatusServer - This struct defined the sample plugin which can be used as a starting point
type K8sAgentControlStatusServer struct {
	cmdExec       tasks.CmdExecFunc
	appName       string
	labelSelector string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sAgentControlStatusServer) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString(fmt.Sprintf("K8s/AgentControl/%s-status-server", p.appName))
}

// Explain - Returns the help text for each individual task
func (p K8sAgentControlStatusServer) Explain() string {
	return "Collects agent-control " + p.appName + " status sever"
}

// Dependencies - Returns the dependencies for each task.
func (p K8sAgentControlStatusServer) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sAgentControlStatusServer) Execute(options tasks.Options, _ map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["k8sNamespace"]
	podName, err := p.retrievePodName(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving agent-control podName: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	res, err = p.runCommand(namespace, podName)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving status server output: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s agent-control status server output",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: fmt.Sprintf("%s status-server-output", p.appName), Stream: stream}},
	}
}

func (p K8sAgentControlStatusServer) runCommand(namespace, podName string) ([]byte, error) {
	args := []string{
		"debug",
		podName,
		"--image",
		"busybox",
		"-q",
		"-i",
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	// debug command
	args = append(args, "--", "wget", "localhost:51200/status", "-q", "-O", "-")

	return p.cmdExec(
		kubectlBin,
		args...,
	)
}

func (p K8sAgentControlStatusServer) retrievePodName(namespace string) (string, error) {
	args := []string{
		"get",
		"-l",
		p.labelSelector,
		"pods",
		"-o",
		"jsonpath='{.items[0].metadata.name}'",
	}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	podNameRaw, err := p.cmdExec(
		kubectlBin,
		args...,
	)
	podName := strings.Trim(string(podNameRaw), "'")
	if err != nil {
		return "", fmt.Errorf("retrieving podName :%w", err)
	}
	if podName == "" {
		return "", fmt.Errorf("no pod with label %s found in namespace %s", p.labelSelector, namespace)
	}
	return podName, nil
}
