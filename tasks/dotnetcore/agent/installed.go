package agent

import (
	"runtime"

	"github.com/newrelic/NrDiag/tasks"
)

var (
	windowsCoreInstallPaths = []string{
		// DLL paths for .NET agent >= 8.19
		`C:\Program Files\New Relic\.NET Agent\netcore\`,
		// DLL paths for .NET agent < 8.19
		`C:\Program Files\New Relic\.NET Agent\`,
	}
	linuxCoreInstallPath = "/usr/local/newrelic-netcore20-agent/"
	coreAgentDllFilename = "NewRelic.Agent.Core.dll"
)

type DotNetCoreAgentInstalled struct {
	name string
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
			Summary: "The .NET CoreCLR agent is not installed in: " + dllPath,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Found the .NET CoreCLR Agent installed in: " + dllPath,
		Payload: dllPath,
	}
}

// checks for NewRelic.Agent.Core.dll in directory.
func checkForAgentDll() (string, bool) {
	if runtime.GOOS == "windows" {
		for _, path := range windowsCoreInstallPaths {
			if tasks.FileExists(path + coreAgentDllFilename) {
				return path, true
			}
		}
		return "", false
	}
	return linuxCoreInstallPath, tasks.FileExists(linuxCoreInstallPath + coreAgentDllFilename)
}
