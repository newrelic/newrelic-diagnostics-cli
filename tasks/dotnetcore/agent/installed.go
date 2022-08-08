package agent

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var coreAgentDllFilename = "NewRelic.Agent.Core.dll"

type DotNetCoreAgentInstalled struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetCoreAgentInstalled) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Agent/Installed")
}

// Explain - Returns the help text for each individual task
func (p DotNetCoreAgentInstalled) Explain() string {
	return "Detect New Relic .NET Core agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p DotNetCoreAgentInstalled) Dependencies() []string {
	return []string{}
}

// Execute - This tasks checks if the .Net Agent is installed by looking for required dlls
func (p DotNetCoreAgentInstalled) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	dllPath, isDllFilesFound := checkForAgentDll()

	if !isDllFilesFound {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Unable to locate the .NET Core agent's installation files",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Found the .NET Core Agent installed in: " + dllPath,
		Payload: dllPath,
	}
}

// checks for NewRelic.Agent.Core.dll in directory.
func checkForAgentDll() (string, bool) {
	for _, path := range DotNetCoreAgentPaths {
		if tasks.FileExists(path + coreAgentDllFilename) {
			return path, true
		}
	}
	return "", false
}
