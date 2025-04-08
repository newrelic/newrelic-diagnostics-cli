package helm

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// HelmReleases - This struct defined the sample plugin which can be used as a starting point
type HelmReleases struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p HelmReleases) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Helm/Releases")
}

// Explain - Returns the help text for each individual task
func (p HelmReleases) Explain() string {
	return "Collects the list of helm releases."
}

// Dependencies - Returns the dependencies for each task.
func (p HelmReleases) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p HelmReleases) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["k8sNamespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving the list of helm releases: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected the list of helm releases",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "HelmReleases.txt", Stream: stream}},
	}
}

func (p HelmReleases) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			helmBin,
			"list",
			"-a",
		)
	}
	return p.cmdExec(
		helmBin,
		"list",
		"-n",
		namespace,
		"-a",
	)
}
