package agent

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// Paths to check for agent DLLs. Order of this slice indicates priority of returned install when multiple
// agent installs are detected.
var agentInstallPaths = []DotNetAgentInstall{
	// DLL paths for .NET agent >= 8.19
	DotNetAgentInstall{
		AgentPath:    `C:\Program Files\New Relic\.NET Agent\netframework\NewRelic.Agent.Core.dll`,
		ProfilerPath: `C:\Program Files\New Relic\.NET Agent\netframework\NewRelic.Profiler.dll`,
	},
	// DLL paths for .NET agent < 8.19
	DotNetAgentInstall{
		AgentPath:    `C:\Program Files\New Relic\.NET Agent\NewRelic.Agent.Core.dll`,
		ProfilerPath: `C:\Program Files\New Relic\.NET Agent\NewRelic.Profiler.dll`,
	},
}

// DotNetAgentInstalled - This struct defined the DotNetAgentInstalled plugin which is used to check if Agent is installed
type DotNetAgentInstalled struct {
	agentInstallPaths []DotNetAgentInstall
}

// DotNetAgentInstall - Contains information about .NET agent install detected on system.
type DotNetAgentInstall struct {
	AgentPath    string
	ProfilerPath string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetAgentInstalled) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Agent/Installed")
}

// Explain - Returns the help text for each individual task
func (p DotNetAgentInstalled) Explain() string {
	return "Detect New Relic .NET agent"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p DotNetAgentInstalled) Dependencies() []string {
	return []string{
		"Base/Config/Validate",
	}
}

func (p DotNetAgentInstalled) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if !upstream["Base/Config/Validate"].HasPayload() {
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.NoAgentDetectedSummary,
		}
	}

	validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	for _, validation := range validations {

		if (validation.Config.FileName) != "newrelic.config" {
			continue
		}

		agentInstall, ok := findAgentInstall(p.agentInstallPaths)
		if !ok {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: "Could NOT find one or more dlls required by the .Net Agent. Either the .NET Agent is not installed or missing essential dlls. Try running the installer to resolve the issue.",
			}
		}

		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Found dlls required by the .NET Agent",
			Payload: agentInstall,
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: tasks.NoAgentDetectedSummary,
	}
}

// findAgentInstall will return the first DotNetAgentInstall where both paths point to files that exist.
// Because it returns on the first found match, order of the paths slice parameter is important.
func findAgentInstall(paths []DotNetAgentInstall) (DotNetAgentInstall, bool) {
	for _, p := range paths {
		if tasks.FileExists(p.AgentPath) && tasks.FileExists(p.ProfilerPath) {
			return p, true
		}
	}
	return DotNetAgentInstall{}, false
}
