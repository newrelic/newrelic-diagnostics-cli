package requirements

import (
	"fmt"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const processorRegLoc = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment\PROCESSOR_ARCHITECTURE`

type DotnetRequirementsProcessorType struct {
	getProcessorArch tasks.GetProcessorArchFunc
}

func (p DotnetRequirementsProcessorType) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/ProcessorType")
}

func (p DotnetRequirementsProcessorType) Explain() string {
	return "Check processor architecture compatibility with New Relic .NET agent"
}

func (p DotnetRequirementsProcessorType) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

func (p DotnetRequirementsProcessorType) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	var result tasks.Result
	// abort if it isn't installed
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}

	procType, err := p.getProcessorArch()

	if err != nil {
		log.Debug("Error while getting Processor type", err.Error())

		result.Status = tasks.Error
		result.Summary = "Error while getting Processor type, see debug logs for more details."
		return result

	}

	if procType == "x86" {
		result.Status = tasks.Success
		result.Summary = "Processor detected as x86"
		return result

	}

	if procType == "AMD64" {
		result.Status = tasks.Success
		result.Summary = "Processor detected as x64"
		return result

	}

	result.Status = tasks.Failure
	result.Summary = fmt.Sprintf("Processor detected as %v. .Net Framework Agent only supports x86 and x64 processors", procType)
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#architecture"
	return result

}
