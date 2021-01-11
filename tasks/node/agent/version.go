package agent

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	logtask "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
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
	var result tasks.Result
	var nodeModuleVersions []NodeModuleVersion

	if upstream["Node/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Node agent not detected"
		return result
	}
	if upstream["Node/Env/Dependencies"].Status == tasks.Success {
		nodeModuleVersions, ok = upstream["Node/Env/Dependencies"].Payload.([]NodeModuleVersion) 
		if !ok {
			return tasks.result {
				Status: tasks.None,
				Summary:  "Node module version not detected",
			}
		}
	}

	logs, ok := upstream["Base/Log/Copy"].Payload.([]logtask.LogElement)
	if !ok {
		result.Status = tasks.None
		result.Summary = "Node agent version not detected"
		return result
	}
	agentVersion := getNodeVerFromLog(logs)
	log.Debug("Agent version", agentVersion)
	for _, nodeLog := range agentVersion {
		if nodeLog.MatchFound == true {
			result.Status = tasks.Info
			result.Summary = nodeLog.AgentVersion
			result.Payload = nodeLog.AgentVersion
			return result
		}
	}
	return result
}

func getNodeVerFromPayload(payload []NodeModuleVersion) (string) {
	var results [][]string
	regexKey, err := regexp.Compile(`Node[.]js[.] Agent`)
	for _, nodeModuleVersion := range payload {
		matches := regexKey.FindStringSubmatch(nodeModuleVersion.Module)
			if len(matches) > 0 {
				results = append(results, nodeModuleVersion.Module + NodeModuleVersion.Version)
			}
		}
		if len(results) > 0 {
			return results, nil
		}
		log.Debug("Node Agent not found in Node Modules")
		return [][]string{}, nil
	}
}