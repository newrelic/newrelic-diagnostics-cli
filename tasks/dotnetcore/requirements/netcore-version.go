package requirements

import (
	"strconv"
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
		"DotNet/"
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

	coreInstalledVersions, ok := upstream["DotNetCore/Env/Versions"].Payload.([]string)

	// dotnetCoreVers, ok := upstream["DotNetCore/Agent/Installed"].Payload.([]string)

	if !ok {
		result.Status = tasks.Error
		result.Summary = "Could not resolve payload of dependent task, DotNetCore/Env/Versions."
		return
	}

	if isDotnetCoreVerSupported(coreInstalledVersions) {
		return tasks.Result{
			Status: tasks.Success,
			Summary: fmt.Sprintf(".NET Core 2.0 or higher detected: %s", strings.Join(coreInstalledVersions, ", ")),
		}
	}

	return tasks.Result{
		Status: tasks.Warning,
		Summary: fmt.Sprintf("One or more .NET Core versions did not meet our agent version requirements: %s", strings.Join(installedVersions, ", ")),
		URL: "https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-core#net-version",
	}
}

func isDotnetCoreVerSupported(dotnetCoreInstalledVers []string) (result tasks.Result) {
	var successfulDotnetCoreVers []string
	for _, dotnetCoreVer := range dotnetCoreInstalledVers {
		majorVer, _, _, _ := tasks.GetVersionSplit(dotnetCoreVer)
		majorVerToStr := strconv.Itoa(majorVer)

		minimumAgentVer, isPresent := compatibilityVars.DotnetCoreSupportedVersions[majorVerToStr]

		if isPresent {
			successfulDotnetCoreVers = append(successfulDotnetCoreVers, majorVerToStr)
		}

		// if majorVer >= compatibilityVars.DotnetCoreSupportedVersions {
		// 	result.Status = tasks.Success
		// 	result.Summary = ".NET Core 2.0 or higher detected."
		// 	return
		// }
	}

	result.Status = tasks.Failure
	result.Summary = ".NET Core 2.0 or higher not detected."
	result.URL = resultURL
	return
}
