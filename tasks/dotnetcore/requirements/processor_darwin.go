package requirements

import (
	"github.com/newrelic/NrDiag/tasks"
)

// DotNetCoreRequirementsProcessorType - This task checks the kernel/processor architecture against the .NET Core Agent requirements
type DotNetCoreRequirementsProcessorType struct {
}

// Identifier - This returns the Category, Subcategory and Name of the task
func (t DotNetCoreRequirementsProcessorType) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Requirements/ProcessorType")
}

// Explain - Returns the help text of the task
func (t DotNetCoreRequirementsProcessorType) Explain() string {
	return "Check processor architecture compatibility with New Relic .NET Core agent"
}

// Dependencies - Returns the dependencies of the task
func (t DotNetCoreRequirementsProcessorType) Dependencies() []string {
	return []string{}
}

// .NET Core Agent doesn't support mac. This task is just here to avoid build errors

// Execute - The core work within each task
func (t DotNetCoreRequirementsProcessorType) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	result.Status = tasks.None
	result.Summary = "Did not pass OS check, skipping this task."
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#architecture"
	return
}
