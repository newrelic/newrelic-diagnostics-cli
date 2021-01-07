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
	// for _, logElement := range logs {
	// 	logAgentVersion, errSearchLogs := searchLogs(logElement)
	// 	var logElementNode logNodeAgentVersion
	// 	if errSearchLogs != nil || logAgentVersion == "" {
	// 		logElementNode.Logfile = logElement.FilePath + logElement.FileName
	// 		logElementNode.AgentVersion = ""
	// 		logElementNode.MatchFound = false
	// 	} else {
	// 		logElementNode.Logfile = logElement.FilePath + logElement.FileName
	// 		logElementNode.AgentVersion = logAgentVersion
	// 		logElementNode.MatchFound = true
	// 	}
	// 	agentVersion = append(agentVersion, logElementNode)
	// 	log.Debug("Agent version is", logAgentVersion)
	// }
	// return
	payload


}

func searchLogs(logElement logtask.LogElement) (string, error) {
	searchString := `Node[.]js[.] Agent version:\s*([0-9.]+);`
	filepath := logElement.FilePath + logElement.FileName
	agentVersion, err := tasks.ReturnLastStringSubmatchInFile(searchString, filepath)



	// cmdOutput, cmdError := p.cmdExec("npm", "ls", "--parseable=true", "--long=true", "--depth=0")
	// modulesList := string(cmdOutput)
	// if cmdError != nil {
	// 	return modulesList, cmdError
	// }
	// return modulesList, nil

	if err != nil {
		return "", err
	}

	if len(agentVersion) > 0 {
		// we've got a match, return the first capture group
		return agentVersion[1], nil
	}

	return "", nil
}
