package flux

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// FluxReleases - This struct defined the sample plugin which can be used as a starting point
type FluxReleases struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p FluxReleases) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Flux/Releases")
}

// Explain - Returns the help text for each individual task
func (p FluxReleases) Explain() string {
	return "Collects Flux Helm Releases information."
}

// Dependencies - Returns the dependencies for each task.
func (p FluxReleases) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p FluxReleases) Execute(options tasks.Options, _ map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["k8sNamespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving Flux Helm Releases: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected Flux Helm Releases",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "FluxHelmReleases.txt", Stream: stream}},
	}
}

func (p FluxReleases) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"get",
			"helmreleases.helm.toolkit.fluxcd.io",
			"-o",
			"yaml",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"get",
		"helmreleases.helm.toolkit.fluxcd.io",
		"-n",
		namespace,
		"-o",
		"yaml",
	)
}
