package env

import (
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// NodeEnvNpmVersion - This struct defines the Ruby version
type NodeEnvNpmVersion struct {
	cmdExecutor      tasks.CmdExecFunc
	npmVersionGetter getNpmVersionFunc
}

type getNpmVersionFunc func(tasks.CmdExecFunc) (string, error)

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvNpmVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/NpmVersion")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvNpmVersion) Explain() string {
	return "Determine NPM version"
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvNpmVersion) Dependencies() []string {
	return []string{
		"Node/Env/Version",
	}
}

// Execute - The core work within each task
func (p NodeEnvNpmVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result //pass the result back to core and report to UI

	if upstream["Node/Env/Version"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Node.js was not detected. This task didn't run."
		return result
	}

	npmVersion, err := p.npmVersionGetter(p.cmdExecutor)
	if err != nil {
		result.Status = tasks.Error
		result.Summary = "Unable to execute command: $ npm -v. Error: " + err.Error()
		return result
	}

	if len(npmVersion) > 0 {
		result.Summary = npmVersion
		result.Status = tasks.Info
		return result
	}

	result.Status = tasks.Error
	result.Summary = "Unable to get npm version."
	return result
}

func getNpmVersion(commandExecutor tasks.CmdExecFunc) (string, error) {
	version, err := commandExecutor("npm", "-v")
	log.Debug("Running npm -v returned ", version)
	if err != nil {
		log.Debug("Error running npm -v ", version)
		return "", err
	}
	return strings.TrimSpace(string(version)), nil
}
