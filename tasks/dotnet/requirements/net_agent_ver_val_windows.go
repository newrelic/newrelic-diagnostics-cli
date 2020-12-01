package requirements

import (
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
	log "go.datanerd.us/p/support-tools/nr-diagnostics/logger"
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

	depsOK, failureSummary := checkDependencies(upstream)

	if !depsOK {
		return tasks.Result{
			Status:  tasks.None,
			Summary: failureSummary,
		}
	}

	agentVer := upstream["DotNet/Agent/Version"].Summary
	dotNetAgentVersion, good := tasks.ParseVersion(agentVer) //Examples of how this string looks like: 8.30.0.0 or 8.3.360.0. TODO: will our compatibilityVars.DotnetFrameworkSupportedVersions complain if we are passing a 4 digit string rather than something like 7.0.0???
	if good != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error parsing Target .Net Agent version " + good.Error(),
		}
	}
	dotNetVersions := strings.Replace(upstream["DotNet/Env/TargetVersion"].Summary, " ", "", -1) //gets a string containing multiple dotnet versions
	dotNetVersionsList := strings.Split(dotNetVersions, ",")

	//The .Net target version check can pick up multiple config files
	//If there are multiple versions of .Net targeted it indicates an issue

	//we arbitrarily grab the first because we assume they are all the same versions
	//UPDATE going through and checking compatibility for each dotnet framework version
	var compatibleVersions []string
	var incompatibleVersions []string
	var warningMessage string
	var successMessage string

	if !checkConsistentTargets(dotNetVersionsList) {
		//TODO: check if any of them is not compatible with the agent, otherwise no need for a warning. Here is an example ticket of how we get multiple versions yet all of them are compatible with the agent:https://newrelic.zendesk.com/agent/tickets/414788
		//UPDATE: this is just to return a warning if multiple dotnet versions are detected
		warningMessage += "Detected multiple target .NET versions.\nThe target .NET versions detected as: " + dotNetVersions + " and Agent version detected as: " + agentVer + "\n"

	}
	for _, version := range dotNetVersionsList {

		//check if version is in the supported framework version
		requiredAgentVersions, isTargetVersionSupported := compatibilityVars.DotnetFrameworkSupportedVersions[version]
		if !isTargetVersionSupported {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: "The detected Target .NET version is not supported by any .NET agent version: " + version,
			}
		}

		//check if the dotnetTarget version can be parsed
		/*dotnetTargetVersion, ok := tasks.ParseVersion(version)
		if ok != nil {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: "Error parsing Target .Net version " + ok.Error(),
			}
		}*/

		//check if the dotnet agent version and dotnet framework version are compatible

		isDotnetAgentVersionCompatible, err := dotNetAgentVersion.CheckCompatibility(requiredAgentVersions)
		if err != nil {
			var errMsg string = err.Error()
			log.Debug(errMsg)
			return tasks.Result{
				Status:  tasks.Error,
				Summary: fmt.Sprintf("We ran into an error while parsing your current agent version %s. %s", agentVer, errMsg),
			}
		}
		if isDotnetAgentVersionCompatible {
			compatibleVersions = append(compatibleVersions, version)
			successMessage += "Compatible Version detected: " + version + "\n"
		} else {
			incompatibleVersions = append(incompatibleVersions, version)
			warningMessage += "Incompatible Version detected: " + version + "\n"
		}
	}

	if len(incompatibleVersions) > 0 && len(compatibleVersions) > 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: warningMessage + successMessage,
		}
	}
	if len(incompatibleVersions) > 0 && len(compatibleVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: warningMessage,
		}
	}
	return tasks.Result{
		Status:  tasks.Success,
		Summary: successMessage,
	}
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

func checkConsistentTargets(targetSplitList []string) bool {
	if len(targetSplitList) > 1 {
		for _, v := range targetSplitList {
			for _, subV := range targetSplitList {
				if v != subV {
					return false
				}
			}
		}
	}
	return true
}
