package agent

import (
	"path/filepath"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetCoreAgentVersion - This struct defines the task
type DotNetCoreAgentVersion struct {
	getFileVersion tasks.GetFileVersionFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetCoreAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (p DotNetCoreAgentVersion) Explain() string {
	return "Determine version of New Relic .NET Core agent"
}

// Dependencies - Returns the dependencies for this task.
func (p DotNetCoreAgentVersion) Dependencies() []string {
	return []string{
		"DotNetCore/Agent/Installed",
	}
}

// Execute - Reads the file version of NewRelic.Agent.Core.dll, which always matches the published version of the agent
func (p DotNetCoreAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNetCore/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNetCore/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNetCore/Agent/Installed",
		}
	}

	agentDir, ok := upstream["DotNetCore/Agent/Installed"].Payload.(string)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	agentVersion, err := p.getFileVersion(filepath.Join(agentDir, coreAgentDllFilename))
	if err != nil {
		log.Info("Error finding .NET Core agent version. The error is ", err)
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error finding .NET Core agent version",
		}
	}

	return tasks.Result{
		Status:  tasks.Info,
		Summary: agentVersion,
		Payload: agentVersion,
	}
}
