package requirements

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
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
	return []string{
		"DotNetCore/Agent/Installed",
		"DotNetCore/Requirements/OS",
	}
}

// Execute - The core work within each task
func (t DotNetCoreRequirementsProcessorType) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Core Agent was not detected, skipping this task."
		return result
	}

	if upstream["DotNetCore/Requirements/OS"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Did not pass OS check, skipping this task."
		return result
	}

	result = getType()
	return
}

func getType() (result tasks.Result) {
	procType, err := tasks.GetProcessorArch()

	if err != nil {
		log.Debug("Error while getting Processor type", err.Error())
		result.Status = tasks.Error
		result.Summary = "Error while getting Processor type, see debug logs for more details."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#architecture"
		return
	}

	procType = strings.ToLower(procType)
	log.Debug("DotNetCoreRequirementsProcessorType - proc type is", procType)

	if procType == "x86_64" || procType == "amd64" {
		result.Status = tasks.Success
		result.Summary = "Processor detected as x64."
		return
	}

	result.Status = tasks.Failure
	result.Summary = "Processor not detected as x86_64 or amd64. .NET Core Agent only supports x86_64 and amd64 processors on Linux."
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#architecture"
	return

}
