package env

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// NodeEnvYarnVersion - This struct defines the Ruby version
type NodeEnvYarnVersion struct {
  cmdExecutor      tasks.CmdExecFunc
	yarnVersionGetter getYarnVersionFunc
}

type getYarnVersionFunc func(tasks.CmdExecFunc) (string, error)

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvYarnVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/YarnVersion")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvYarnVersion) Explain() string {
	return "Determine Yarn version"
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvYarnVersion) Dependencies() []string {
	return []string{
		"Node/Env/Version",
	}
}

// Execute - The core work within each task
func (p NodeEnvYarnVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result //pass the result back to core and report to UI

	if upstream["Node/Env/Version"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Node.js was not detected. This task didn't run."
		return result
	}

	yarnVersion, err := p.yarnVersionGetter(p.cmdExecutor)
	if err != nil {
		result.Status = tasks.Error
		result.Summary = "Unable to execute command: $ yarn -v. Error: " + err.Error()
		return result
	}

	if len(yarnVersion) > 0 {
		result.Summary = yarnVersion
		result.Status = tasks.Info
		return result
	}

	result.Status = tasks.Error
	result.Summary = "Unable to get yarn version."
	return result
}

func getYarnVersion(commandExecutor tasks.CmdExecFunc) (string, error) {
	version, err := commandExecutor("yarn", "-v") //output example:1.22.5
	if err != nil {
		log.Debug("Error running yarn -v ", version) //-bash: yarn: command not found
		return "", err
	}
	return strings.TrimSpace(string(version)), nil
}
