package requirements

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
)

// DotNetCoreRequirementsNetCoreVersion - This task checks the .NET Core version against the .Net Core Agent requirements
type DotNetCoreRequirementsNetCoreVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of the task
func (t DotNetCoreRequirementsNetCoreVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Requirements/DotNetCoreVersion")
}

// Explain - Returns the help text of the task
func (t DotNetCoreRequirementsNetCoreVersion) Explain() string {
	return "Check .NET Core version compatibility with New Relic .NET Core agent"
}

// Dependencies - Returns the dependencies of the task
func (t DotNetCoreRequirementsNetCoreVersion) Dependencies() []string {
	return []string{
		"DotNetCore/Agent/Installed",
		"DotNetCore/Env/Versions",
	}
}

const resultURL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#net-version"

// Execute - The core work within each task
func (t DotNetCoreRequirementsNetCoreVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Did not detect .Net Core Agent as being installed, skipping this task."
		return
	}
	if upstream["DotNetCore/Env/Versions"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Unable to determine versions of .NET Core installed, skipping this task."
		return
	}

	installedVersions, ok := upstream["DotNetCore/Env/Versions"].Payload.([]string)

	if !ok {
		result.Status = tasks.Error
		result.Summary = "Could not resolve payload of dependent task, DotNetCore/Env/Versions."
		return
	}

	result = checkVersion(installedVersions)
	return
}

func checkVersion(installedVersions []string) (result tasks.Result) {
	for _, version := range installedVersions {
		majorVer, _, _, _ := tasks.GetVersionSplit(version)
		if majorVer >= compatibilityVars.DotnetCoreSupportedVersions {
			result.Status = tasks.Success
			result.Summary = ".NET Core 2.0 or higher detected."
			return
		}
	}

	result.Status = tasks.Failure
	result.Summary = ".NET Core 2.0 or higher not detected."
	result.URL = resultURL
	return
}
