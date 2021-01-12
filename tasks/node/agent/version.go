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

type logNodeAgentVersion struct {
	Logfile      string
	AgentVersion string
	MatchFound   bool
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
		"Node/Config/Agent",
		"Base/Log/Copy",
	}
}

// Execute - The core work within this task
func (t NodeAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var nodeModuleVersions []NodeEnv.NodeModuleVersion
	var ok bool

	if upstream["Node/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Node agent not detected",
		}
	}
	if upstream["Node/Env/Dependencies"].Status == tasks.Info {
		nodeModuleVersions, ok = upstream["Node/Env/Dependencies"].Payload.([]NodeEnv.NodeModuleVersion)
		if !ok {
			return tasks.Result{
				Status:  tasks.None,
				Summary: "Node module version not detected",
			}
		}
	}

	agentVersion := getNodeVerFromPayload(nodeModuleVersions)
	if agentVersion == "" {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Node Agent Module not found for newrelic",
		}
	}
	log.Debug("Agent version", agentVersion)
	return tasks.Result{
		Status:  tasks.Info,
		Summary: fmt.Sprintf("Node Agent Version  %s found", agentVersion),
	}
}

func getNodeVerFromPayload(payload []NodeEnv.NodeModuleVersion) string {
	for _, nodeModuleVersion := range payload {
		if nodeModuleVersion.Module == "newrelic" {
			return fmt.Sprintf("%s %s", nodeModuleVersion.Module, nodeModuleVersion.Version)
		}
	}
	log.Debug("Node Agent not found in Node Modules")
	return ""
}
