package agent

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	logtask "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
)

// PythonAgentVersion - This struct defines the Python agent version.
type PythonAgentVersion struct {
}

type LogPythonAgentVersion struct {
	Logfile      string
	AgentVersion string
	MatchFound   bool
}

// Identifier - This returns the Category, Subcategory and Name of this task.
func (t PythonAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Agent/Version")
}

// Explain - Returns the help text for the PythonAgentVersion task.
func (t PythonAgentVersion) Explain() string {
	return "Determine New Relic Python agent version"
}

// Dependencies - Returns the dependencies for this task.
func (t PythonAgentVersion) Dependencies() []string {
	return []string{
		"Python/Config/Agent",
		"Base/Log/Copy",
	}
}

// Execute - The core work within this task.
func (t PythonAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if upstream["Python/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Python agent not detected"
		return result
	}

	logs, ok := upstream["Base/Log/Copy"].Payload.([]logtask.LogElement)
	if !ok {
		result.Status = tasks.None
		result.Summary = "Python agent version not detected"
		return result
	}
	agentVersion := getPythonVersionFromLog(logs)
	log.Debug("Agent version", agentVersion)
	for _, pythonLog := range agentVersion {
		if pythonLog.MatchFound {
			result.Status = tasks.Info
			result.Summary = pythonLog.AgentVersion
			result.Payload = pythonLog.AgentVersion
			return result
		}
	}
	return result
}

func getPythonVersionFromLog(logs []logtask.LogElement) (agentVersion []LogPythonAgentVersion) {
	for _, logElement := range logs {
		logAgentVersion, errSearchLogs := searchLogs(logElement)
		var logElementPython LogPythonAgentVersion
		if errSearchLogs != nil || logAgentVersion == "" {
			logElementPython.Logfile = logElement.FilePath + logElement.FileName
			logElementPython.AgentVersion = ""
			logElementPython.MatchFound = false
		} else {
			logElementPython.Logfile = logElement.FilePath + logElement.FileName
			logElementPython.AgentVersion = logAgentVersion
			logElementPython.MatchFound = true
		}
		agentVersion = append(agentVersion, logElementPython)
		log.Debug("Agent version is", logAgentVersion)
	}
	return
}

func searchLogs(logElement logtask.LogElement) (string, error) {
	search := `New Relic Python Agent \(([0-9.]+)\)`
	filepath := logElement.FilePath + logElement.FileName
	agentVersion, err := tasks.ReturnLastStringSubmatchInFile(search, filepath)
	if err != nil {
		return "", err
	}

	if len(agentVersion) > 0 {
		// we've got a match, return the first capture group
		return agentVersion[1], nil
	}

	return "", nil
}
