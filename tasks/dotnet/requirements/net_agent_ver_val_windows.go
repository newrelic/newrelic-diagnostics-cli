package requirements

import (
	"strings"

	"github.com/newrelic/NrDiag/tasks"
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
	return "Check application's .NET version compatibility with New Relic .NET agent"
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
	appTarget := strings.Replace(upstream["DotNet/Env/TargetVersion"].Summary, " ", "", -1)
	targetSplitList := strings.Split(appTarget, ",")

	//The .Net target version check can pick up multiple config files
	//If there are multiple versions of .Net targeted it indicates an issue

	if !checkConsistentTargets(targetSplitList) {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Detected multiple versions for .Net Target or unable to determine targets. .Net Targets detected as " + appTarget + ", Agent version detected as " + agentVer,
		}
	}
	dotnetTargetVersion, err := tasks.ParseVersion(targetSplitList[0])
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error parsing Target .Net version " + err.Error(),
		}
	}

	//If the app targets a below 2.0, none of our Agents will support it
	minTargetVersion := []string{"2+"}
	isCompatible, _ := dotnetTargetVersion.CheckCompatibility(minTargetVersion)
	if !isCompatible {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "Target .Net version detected as below 2.0. This version of .Net is not supported by any agent versions",
		}
	}

	agentVersion, err := tasks.ParseVersion(agentVer)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error parsing Agent version",
		}
	}

	//Only the major version of the Agent matters for this check
	if checkVersionOk(dotnetTargetVersion, agentVersion.Major) {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: ".Net target and Agent version compatible. .Net Target detected as " + targetSplitList[0] + ", Agent version detected as " + agentVer,
		}
	}
	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "App detected as targeting a version of .Net below 4.5 with an Agent version of 7 or above. .Net Target detected as " + targetSplitList[0] + ", Agent version detected as " + agentVer,
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

func checkVersionOk(appTarget tasks.Ver, AgentVersion int) bool {
	//This is the only combination that will cause issues, so everything else is fine
	// minAppTarget := tasks.Ver{Major: 4, Minor: 4, Patch: 999} //This allows us to compare to less than 4.5 with less than or equal
	minAppTarget := []string{"4.5+"}
	isCompatible, _ := appTarget.CheckCompatibility(minAppTarget)
	if !isCompatible && AgentVersion >= 7 {
		return false
	}
	return true
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
