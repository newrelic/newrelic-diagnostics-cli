package agent

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// ExampleTemplateMinimalTask - This struct defines this plugin
type DotNetAgentVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t DotNetAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (t DotNetAgentVersion) Explain() string {
	return "Determine version of New Relic .NET agent"
}

// Dependencies - Returns the dependencies for this task.

func (t DotNetAgentVersion) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within this task
func (t DotNetAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}

	agentInstall, ok := upstream["DotNet/Agent/Installed"].Payload.(DotNetAgentInstall)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	return getVersion(agentInstall)
}

// Does the work of getting the version of the Agent.Core.dll, version of this dll should always == published version of the .Net Agent
func getVersion(agentInstall DotNetAgentInstall) (result tasks.Result) {

	agentVersion, err := tasks.GetFileVersion(agentInstall.AgentPath)

	if err != nil {
		log.Info("Error finding .Net Agent version. The error is ", err)
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error finding .Net Agent version",
		}
	}
	return tasks.Result{
		Status:  tasks.Info,
		Summary: agentVersion,
		Payload: agentVersion,
	}
}
