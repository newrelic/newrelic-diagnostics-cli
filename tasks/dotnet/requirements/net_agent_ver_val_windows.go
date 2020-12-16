package requirements

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
)

// DotnetRequirementsNetTargetAgentVerValidate - This struct defines the task
type DotnetRequirementsNetTargetAgentVerValidate struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t DotnetRequirementsNetTargetAgentVerValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/NetTargetAgentVersionValidate")
}

// Explain - Returns the help text for this task
func (t DotnetRequirementsNetTargetAgentVerValidate) Explain() string {
	return "Check application's .NET Framework version compatibility with New Relic .NET agent"
}

// Dependencies - Returns the dependencies for this task.
func (t DotnetRequirementsNetTargetAgentVerValidate) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
		"DotNet/Env/TargetVersion",
		"DotNet/Agent/Version",
	}
}

// Execute - The core work within this task
func (t DotnetRequirementsNetTargetAgentVerValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	depsOK, thisTaskDidNotRunSummary := checkDependencies(upstream)

	if !depsOK {
		return tasks.Result{
			Status:  tasks.None,
			Summary: thisTaskDidNotRunSummary,
		}
	}

	agentVersion, ok := upstream["DotNet/Agent/Version"].Payload.(string) //Examples of how this string looks like: 8.30.0.0 or 8.3.360.0
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Type Assertion failure from upstream task DotNet/Agent/Version in DotNet/Requirements/NetTargetAgentVersionValidate",
		}
	}

	frameworkVersions, ok := upstream["DotNet/Env/TargetVersion"].Payload.([]string) //gets a slice containing multiple dotnet versions: .Net Targets detected as 4.6,4.6,4.6,4.7.2,4.6
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Type Assertion failure from upstream task DotNet/Env/TargetVersion in DotNet/Requirements/NetTargetAgentVersionValidate",
		}
	}

	//truncate framework versions to only use two digits (if they happen to be more than two) because both our map and documentation ignores any patch version to assess if it is supported
	truncatedFrameworkVersions, err := removePatchBuildDigits(frameworkVersions)

	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.ThisProgramFullName + " was unable to validate that your .NET agent version is compatible with your .NET target version because it ran into an unexpected error: " + err.Error(),
		}
	}

	var unsupportedFrameworkVersions []string
	var incompatibleFrameworkVersions []string

	for _, frameworkVer := range truncatedFrameworkVersions {
		isFrameworkVerSupported, requiredAgentVersions := checkFrameworkVerIsSupported(frameworkVer)
		if !isFrameworkVerSupported {
			unsupportedFrameworkVersions = append(unsupportedFrameworkVersions, frameworkVer)
			continue
		}
		isCompatibleWithAgent, err := checkCompatibilityWithAgentVer(requiredAgentVersions, agentVersion)

		if err != nil {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.ThisProgramFullName + " was unable to validate that your .NET agent version is compatible with your .NET target version because it ran into an unexpected error: " + err.Error(),
			}
		}
		if !isCompatibleWithAgent {
			incompatibleFrameworkVersions = append(incompatibleFrameworkVersions, frameworkVer)
		}
	}

	var failureSummary string
	if len(unsupportedFrameworkVersions) > 0 {
		failureSummary += fmt.Sprintf("We found a Target Framework version(s) that is not supported by the New Relic .NET agent: %s", strings.Join(unsupportedFrameworkVersions, ", "))
	}
	if len(incompatibleFrameworkVersions) > 0 {
		failureSummary += fmt.Sprintf("We found that your New Relic .NET agent version %s is not compatible with the following Target .NET version(s): %s", agentVersion, strings.Join(incompatibleFrameworkVersions, ", "))
	}

	legacyDocURL := "https://docs.newrelic.com/docs/agents/net-agent/troubleshooting/technical-support-net-framework-40-or-lower"
	requirementsDocURL := "https://docs.newrelic.com/docs/agents/net-agent/getting-started/net-agent-compatibility-requirements-net-framework"

	if len(failureSummary) > 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: failureSummary,
			URL:     requirementsDocURL + "\n" + legacyDocURL,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("Your .NET agent version %s is fully compatible with the following found Target .NET version(s): %s", agentVersion, strings.Join(frameworkVersions, ", ")),
	}

}

func removePatchBuildDigits(frameworkVersions []string) ([]string, error) {
	var truncatedVersions []string
	for _, version := range frameworkVersions {
		v, err := tasks.ParseVersion(version)
		if err != nil {
			return truncatedVersions, err
		}
		truncatedVersion := strconv.Itoa(v.Major) + "." + strconv.Itoa(v.Minor)
		truncatedVersions = append(truncatedVersions, truncatedVersion)

	}
	return truncatedVersions, nil

}

func checkCompatibilityWithAgentVer(requiredAgentVersions []string, agentVersion string) (bool, error) {
	isCompatible, err := tasks.VersionIsCompatible(agentVersion, requiredAgentVersions)
	if err != nil {
		return false, err
	}

	return isCompatible, nil
}

func checkFrameworkVerIsSupported(frameworkVer string) (bool, []string) {
	requiredAgentVersions, isFrameworkVerSupported := compatibilityVars.DotnetFrameworkSupportedVersions[frameworkVer]
	if isFrameworkVerSupported {
		return true, requiredAgentVersions
	}
	requiredLegacyAgentVersions, isOldFrameworkVerSupported := compatibilityVars.DotnetFrameworkOldVersions[frameworkVer]

	return isOldFrameworkVerSupported, requiredLegacyAgentVersions
}

func checkDependencies(upstream map[string]tasks.Result) (bool, string) {
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		return false, "Did not detect .Net Agent as being installed, this check did not run"
	}

	if upstream["DotNet/Env/TargetVersion"].Status != tasks.Info {
		return false, "Did not detect App Target .Net version, this check did not run"
	}

	if upstream["DotNet/Agent/Version"].Status != tasks.Info {
		return false, "Did not detect .Net Agent version, this check did not run"
	}
	return true, ""
}
