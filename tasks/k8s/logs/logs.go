package logs

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const kubectlBin = "kubectl"

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering k8s/logs/*")
	registrationFunc(K8sLogs{
		cmdExec: tasks.CmdExecutor,
	}, true)
}

// K8sLogs - This struct defined the sample plugin which can be used as a starting point
type K8sLogs struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p K8sLogs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("K8s/Logs/Logs")
}

// Explain - Returns the help text for each individual task
func (p K8sLogs) Explain() string {
	return "This task collects the logs of the k8s solution."
}

// Dependencies - Returns the dependencies for each task.
func (p K8sLogs) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p K8sLogs) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	res, err := p.cmdExec(
		kubectlBin,
		"logs",
		"-l",
		"app.kubernetes.io/name=newrelic-infrastructure",
		"--all-containers",
		"--prefix",
	)

	if err != nil {
		return tasks.Result{
			Summary: "Error retrieving logs: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	ch := make(chan string)
	go func() {
		ch <- string(res)
		close(ch)
	}()

	var filesToCopy []tasks.FileCopyEnvelope
	filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: "forNowAllPodsLogs", Stream: ch, Identifier: p.Identifier().String()})

	return tasks.Result{
		Summary:     "Full logs",
		Status:      tasks.Info,
		Payload:     string(res),
		FilesToCopy: filesToCopy,
	}
}
