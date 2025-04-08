package flux

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// FluxRepositories - This struct defined the sample plugin which can be used as a starting point
type FluxRepositories struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p FluxRepositories) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Flux/Repositories")
}

// Explain - Returns the help text for each individual task
func (p FluxRepositories) Explain() string {
	return "Collects Flux Helm Repositories information."
}

// Dependencies - Returns the dependencies for each task.
func (p FluxRepositories) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p FluxRepositories) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		res []byte
		err error
	)

	namespace := options.Options["k8sNamespace"]
	res, err = p.runCommand(namespace)
	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving flux repositories: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(res), stream)

	return tasks.Result{
		Summary:     "Successfully collected Flux Helm Repositories",
		Status:      tasks.Info,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "FluxHelmRepositories.txt", Stream: stream}},
	}
}

func (p FluxRepositories) runCommand(namespace string) ([]byte, error) {
	if namespace == "" {
		return p.cmdExec(
			kubectlBin,
			"get",
			"helmrepositories.source.toolkit.fluxcd.io",
			"-o",
			"yaml",
		)
	}
	return p.cmdExec(
		kubectlBin,
		"get",
		"helmrepositories.source.toolkit.fluxcd.io",
		"-n",
		namespace,
		"-o",
		"yaml",
	)
}
