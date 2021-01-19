package requirements

import (
	"fmt"
	"strconv"
	"strings"

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

const resultURL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-core#net-version"

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

	if !ok {
		result.Status = tasks.Error
		result.Summary = "Could not resolve payload of dependent task, DotNetCore/Env/Versions."
		return
	}

	unsupportedVersions, errorMessage := checkCoreVersionsAreSupported(coreInstalledVersions)

	if len(errorMessage) > 0 {
		return tasks.Result{
			Status: tasks.Error,
			Summary: errorMessage,
		}
	} 

	if len(unsupportedVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: fmt.Sprintf(".NET Core 2.0 or higher detected: %s", strings.Join(coreInstalledVersions, ", ")),
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: fmt.Sprintf("One or more .NET Core versions did not meet our agent version requirements: %s", strings.Join(unsupportedVersions, ", ")),
		URL:     resultURL,
	}
}

func checkCoreVersionsAreSupported(dotnetCoreInstalledVers []string) ([]string, string) {
	var unsupportedVers []string
	var supportedVers []string
	errorMessage := "We were unable to validate if this application is using a supported .NET core version because we ran into some unexpected error(s)\n"
	for _, coreVersion := range dotnetCoreInstalledVers {
		parsedVersion, err := tasks.ParseVersion(coreVersion)
		if err != nil {
			errorMessage += err.Error() + "\n"
			continue
		}
		majorVer, minorVer, _, _ := tasks.GetVersionSplit(parsedVersion.String())
		majorMinorVer := strconv.Itoa(majorVer) + "." + strconv.Itoa(minorVer)
		_, isPresent := compatibilityVars.DotnetCoreSupportedVersions[majorMinorVer]
		if !isPresent {
			unsupportedVers = append(unsupportedVers, coreVersion)
		} else {
			supportedVers = append(supportedVers, coreVersion)
		}
	}

	if len(unsupportedVers) == 0 && len(supportedVers) == 0 {
		//looks like we were unable to parse any version from the slice of strings we received from payload
		return unsupportedVers, errorMessage
	}
	return unsupportedVers, ""//we are going to ignore errorMessage because we were able to parse at least one version from the slice payload
}
