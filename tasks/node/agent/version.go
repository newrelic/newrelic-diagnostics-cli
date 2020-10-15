package agent

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	logtask "github.com/newrelic/NrDiag/tasks/base/log"
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

	if upstream["Node/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Node agent not detected"
		return result
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

func getNodeVerFromLog(logs []logtask.LogElement) (agentVersion []logNodeAgentVersion) {
	for _, logElement := range logs {
		logAgentVersion, errSearchLogs := searchLogs(logElement)
		var logElementNode logNodeAgentVersion
		if errSearchLogs != nil || logAgentVersion == "" {
			logElementNode.Logfile = logElement.FilePath + logElement.FileName
			logElementNode.AgentVersion = ""
			logElementNode.MatchFound = false
		} else {
			logElementNode.Logfile = logElement.FilePath + logElement.FileName
			logElementNode.AgentVersion = logAgentVersion
			logElementNode.MatchFound = true
		}
		agentVersion = append(agentVersion, logElementNode)
		log.Debug("Agent version is", logAgentVersion)
	}
	return
}

func searchLogs(logElement logtask.LogElement) (string, error) {
	searchString := `Node[.]js[.] Agent version:\s*([0-9.]+);`
	filepath := logElement.FilePath + logElement.FileName
	agentVersion, err := tasks.ReturnLastStringSubmatchInFile(searchString, filepath)
	if err != nil {
		return "", err
	}

	if len(agentVersion) > 0 {
		// we've got a match, return the first capture group
		return agentVersion[1], nil
	}

	return "", nil
}
