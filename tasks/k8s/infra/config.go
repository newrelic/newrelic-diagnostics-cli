package infra

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// K8sConfigs - This struct defined the sample plugin which can be used as a starting point
type K8sConfigs struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sConfigs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Infra/Config")
}

// Explain - Returns the help text for each individual task
func (p K8sConfigs) Explain() string {
	return "Collects K8s configMaps for newrelic-infrastructure in YAML format."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sConfigs) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sConfigs) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["namespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving newrelic-infrastructure configMaps: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected K8s newrelic-infrastructure configMaps ",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "k8sInfraConfigMaps.txt", Stream: stream}},
	}
}

func (p K8sConfigs) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"get",
			"configMaps",
			"-l",
			"app.kubernetes.io/name=newrelic-infrastructure",
			"-o",
			"yaml",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"get",
		"configMaps",
		"-n",
		namespace,
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
		"-o",
		"yaml",
	)
}
