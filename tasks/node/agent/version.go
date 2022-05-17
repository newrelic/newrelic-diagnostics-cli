package agent

import (
	"fmt"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	NodeEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/env"
)

// NodeAgentVersion - This struct defines the function to collect the current running Node agent version from its logfile.
type NodeAgentVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t NodeAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Agent/Version")
}

// Explain - Returns the help text for this task
func (t NodeAgentVersion) Explain() string {
	return "Determine New Relic Nodejs agent version"
}

// Dependencies - Returns the dependencies for this task.
func (t NodeAgentVersion) Dependencies() []string {
	return []string{
		"Node/Env/Dependencies",
	}
}

// Execute - The core work within this task
func (t NodeAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Node/Env/Dependencies"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Node Modules not detected. This task did not run.",
		}
	}

	nodeModuleVersions, ok := upstream["Node/Env/Dependencies"].Payload.([]NodeEnv.NodeModuleVersion)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	agentVersion := getNodeVerFromPayload(nodeModuleVersions)
	if agentVersion == "" {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "We were unable to find the 'newrelic' module required for the Node Agent installation. Make sure to run 'npm install newrelic' and verify that 'newrelic' is listed in your package.json.",
		}
	}

	log.Debug("Agent version", agentVersion)
	return tasks.Result{
		Status:  tasks.Info,
		Summary: fmt.Sprintf("Node Agent Version %s found", agentVersion),
		Payload: agentVersion,
	}
}

func getNodeVerFromPayload(payload []NodeEnv.NodeModuleVersion) string {
	for _, nodeModuleVersion := range payload {
		if nodeModuleVersion.Module == "newrelic" {
			return nodeModuleVersion.Version
		}
	}

	log.Debug("Node Agent not found in Node Modules")
	return ""
}
