package configs

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering k8s/configs/*")
	registrationFunc(K8sConfigs{
		cmdExec: tasks.CmdExecutor,
	}, true)
}

// K8sConfigs - This struct defined the sample plugin which can be used as a starting point
type K8sConfigs struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sConfigs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Configs/Configs")
}

// Explain - Returns the help text for each individual task
func (p K8sConfigs) Explain() string {
	return "This task collects YAML of configMaps."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sConfigs) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sConfigs) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	res, err := p.cmdExec(
		kubectlBin,
		"get",
		"configMaps",
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
		"-o",
		"yaml",
	)

	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving config: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	ch := make(chan string)
	go func() {
		ch <- string(res)
		close(ch)
	}()

	var filesToCopy []tasks.FileCopyEnvelope
	filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: "forNowAllConfigsYaml", Stream: ch, Identifier: p.Identifier().String()})

	return tasks.Result{
		Summary:     "Full configMap YAML",
		Status:      tasks.Info,
		Payload:     string(res),
		FilesToCopy: filesToCopy,
	}
}
