package env

import (
	"fmt"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// To update compatibility requirements, modify what's the contents of supportedNodeVersions
// which are passed to this task's registrationFunc in this package's defintiion file (./env.go)

// NodeEnvVersionCompatibility - This struct defines Node.js version compatibility
type NodeEnvVersionCompatibility struct {
	supportedNodeVersions []string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvVersionCompatibility) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/VersionCompatibility")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvVersionCompatibility) Explain() string {
	return "Check Nodejs version compatibility with New Relic Nodejs agent"
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvVersionCompatibility) Dependencies() []string {
	return []string{"Node/Env/Version"}
}

// Execute - The core work within each task
func (p NodeEnvVersionCompatibility) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Node/Env/Version"].Status != tasks.Info {
		return tasks.Result{
			Summary: "Task did not meet requirements necessary to run: Node is not installed",
			Status:  tasks.None,
		}
	}

	nodeVersion, ok := upstream["Node/Env/Version"].Payload.(tasks.Ver)
	if !ok {
		return tasks.Result{
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
			Status:  tasks.None,
		}
	}

	isItCompatible, err := nodeVersion.CheckCompatibility(p.supportedNodeVersions)
	if err != nil {
		return tasks.Result{
			Summary: "There was an issue when checking for Node.js Version compatibility: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	if !isItCompatible {
		return tasks.Result{
			Summary: fmt.Sprintf("Your current Node.js version, %v, is not compatible with New Relic's Node.js agent", nodeVersion),
			Status:  tasks.Failure,
			URL:     "https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent",
		}
	}

	return tasks.Result{
		Summary: fmt.Sprintf("Your current Node.js version, %v, is compatible with New Relic's Node.js agent", nodeVersion),
		Status:  tasks.Success,
	}

}
