package requirements

import (
	"github.com/newrelic/NrDiag/tasks"
)

// DotnetRequirementsChecker - This struct defines this plugin
type DotnetRequirementsRequirementCheck struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (p DotnetRequirementsRequirementCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/RequirementCheck")
}

// Explain - Returns the help text for this task
func (p DotnetRequirementsRequirementCheck) Explain() string {
	return "Validate New Relic .NET agent related diagnostic checks"
}

// Dependencies - Returns the dependencies for this task.
func (p DotnetRequirementsRequirementCheck) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
		"DotNet/Requirements/NetTargetAgentVersionValidate",
		"DotNet/Requirements/OS",
		"Dotnet/Requirements/OwinCheck",
		"DotNet/Requirements/ProcessorType",
		"DotNet/Requirements/Datastores",
		"DotNet/Requirements/MessagingServicesCheck",
	}
}

// Execute - The core work within this task
func (p DotnetRequirementsRequirementCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: ".Net Agent not detected as installed, this check didn't run",
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent",
		}
	}

	var failedTasks []string
	// loop over list of all possible dependencies
	for _, v := range p.Dependencies() {

		if checkTaskStatus(v, upstream[v].Status) {
			continue
		} else {
			failedTasks = append(failedTasks, v)
		}
	}

	if len(failedTasks) > 0 {
		summary := "Detected failed DotNet Agent requirement checks. Failed checks: \n"
		for _, v := range failedTasks {
			summary += v + "\n"

		}
		return tasks.Result{
			Summary: summary,
			Status:  tasks.Failure,
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "All DotNet Agent requirement checks detected as having succeeded.",
	}
}

//Some checks are okay if the status is none or success
func checkTaskStatus(task string, status tasks.Status) bool {

	//map of tasks where a None status is acceptable
	noneSupportedTasks := map[string]bool{
		"DotNet/Requirements/Datastores":             true,
		"DotNet/Requirements/MessagingServicesCheck": true,
		"Dotnet/Requirements/OwinCheck":              true,
	}

	if status == tasks.Success {
		return true
	} else if status == tasks.None && noneSupportedTasks[task] {
		return true
	}
	return false
}
