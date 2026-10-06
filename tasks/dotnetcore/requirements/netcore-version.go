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
		"DotNetCore/Agent/Version",
	}
}

const resultURL = "https://docs.newrelic.com/docs/apm/agents/net-agent/getting-started/net-agent-compatibility-requirements/#net-version-core"

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

	//the agent version is optional: without it we can still tell whether the .NET versions are supported at all
	agentVersion := ""
	if upstream["DotNetCore/Agent/Version"].Status == tasks.Info {
		agentVersion, _ = upstream["DotNetCore/Agent/Version"].Payload.(string)
	}

	unsupportedVersions, incompatibleVersions, errorMessage := checkCoreVersionsAreSupported(coreInstalledVersions, agentVersion)

	if len(errorMessage) > 0 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: errorMessage,
		}
	}

	if len(unsupportedVersions) == 0 && len(incompatibleVersions) == 0 {
		summary := fmt.Sprintf("Supported .NET version(s) detected: %s", strings.Join(coreInstalledVersions, ", "))
		if agentVersion != "" {
			summary = fmt.Sprintf("Your .NET agent version %s is compatible with the detected .NET version(s): %s", agentVersion, strings.Join(coreInstalledVersions, ", "))
		}
		return tasks.Result{
			Status:  tasks.Success,
			Summary: summary,
		}
	}

	var warnings []string
	if len(unsupportedVersions) > 0 {
		warnings = append(warnings, fmt.Sprintf("One or more .NET versions are not supported by the New Relic .NET agent: %s", strings.Join(unsupportedVersions, ", ")))
	}
	if len(incompatibleVersions) > 0 {
		warnings = append(warnings, fmt.Sprintf("Your .NET agent version %s is too old to support the following .NET version(s): %s", agentVersion, strings.Join(incompatibleVersions, ", ")))
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: strings.Join(warnings, "\n"),
		URL:     resultURL,
	}
}

// checkCoreVersionsAreSupported returns the .NET versions the agent does not support at all, and the versions that
// require a newer agent than agentVersion. The agent version check is skipped when agentVersion is empty.
func checkCoreVersionsAreSupported(dotnetCoreInstalledVers []string, agentVersion string) ([]string, []string, string) {
	var unsupportedVers []string
	var incompatibleVers []string
	parsedAny := false
	errorMessage := "We were unable to validate if this application is using a supported .NET core version because we ran into some unexpected error(s)\n"
	for _, coreVersion := range dotnetCoreInstalledVers {
		parsedVersion, err := tasks.ParseVersion(coreVersion)
		if err != nil {
			errorMessage += err.Error() + "\n"
			continue
		}
		parsedAny = true
		majorMinorVer := strconv.Itoa(parsedVersion.Major) + "." + strconv.Itoa(parsedVersion.Minor)
		requiredAgentVersions, isPresent := compatibilityVars.DotnetCoreSupportedVersions[majorMinorVer]
		if !isPresent {
			unsupportedVers = append(unsupportedVers, coreVersion)
			continue
		}
		if agentVersion == "" {
			continue
		}
		isCompatible, err := tasks.VersionIsCompatible(agentVersion, requiredAgentVersions)
		if err != nil {
			return nil, nil, errorMessage + err.Error() + "\n"
		}
		if !isCompatible {
			incompatibleVers = append(incompatibleVers, coreVersion)
		}
	}

	if !parsedAny {
		//looks like we were unable to parse any version from the slice of strings we received from payload
		return unsupportedVers, incompatibleVers, errorMessage
	}
	return unsupportedVers, incompatibleVers, "" //we are going to ignore errorMessage because we were able to parse at least one version from the slice payload
}
