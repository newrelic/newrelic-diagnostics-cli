// +build windows

package profiler

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetProfilerInstrumentationPossible -
type DotNetProfilerInstrumentationPossible struct {
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetProfilerInstrumentationPossible) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/InstrumentationPossible")
}

// Explain - Returns the help text for each individual task
func (p DotNetProfilerInstrumentationPossible) Explain() string {
	return "Validate registry keys and environment variables required for .NET application profiling"
}

// Dependencies - This task has no dependencies
func (p DotNetProfilerInstrumentationPossible) Dependencies() []string {

	return []string{
		"DotNet/Agent/Installed",
		"DotNet/Profiler/W3svcRegKey",
		"DotNet/Profiler/EnvVarKey",
		"DotNet/Profiler/WasRegKey",
	}

}

//Uses results of wasRegKey,w3svcRegKey, and envVarKey tasks to determin if they are set in such a way that instrumentation is possible
func (p DotNetProfilerInstrumentationPossible) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if (upstream["DotNet/Agent/Installed"].Status == tasks.Failure) || (upstream["DotNet/Agent/Installed"].Status == tasks.None) {

		result.Status = tasks.None
		result.Summary = "Did not detect .Net Agent as being installed, this check did not run"
		return result
	}

	wasSet := upstream["DotNet/Profiler/WasRegKey"].Status
	w3svcSet := upstream["DotNet/Profiler/W3svcRegKey"].Status
	envVarSet := upstream["DotNet/Profiler/EnvVarKey"].Status

	if wasSet == tasks.Success && w3svcSet == tasks.Success && envVarSet == tasks.Success {

		result.Status = tasks.Success
		result.Summary = "IIS and System wide .Net instrumentation keys are correctly set"
		return result
	} else if wasSet == tasks.Success && w3svcSet == tasks.Success {

		result.Status = tasks.Success
		result.Summary = "IIS .Net instrumentation keys are correctly set"
		return result
	} else if envVarSet == tasks.Success {

		result.Status = tasks.Success
		result.Summary = "System wide .Net instrumentation keys are correctly set"
		return result
	} else {
		result.Status = tasks.Failure
		result.Summary = "Keys needed for .Net instrumentation are not set"
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/troubleshooting/no-data-appears-net"
		return result
	}

}
