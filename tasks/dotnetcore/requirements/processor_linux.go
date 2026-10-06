package requirements

import (
	"fmt"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
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
		"Base/Env/HostInfo",
		"DotNetCore/Env/Versions",
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

	procType, err := tasks.GetProcessorArch()
	if err != nil {
		log.Debug("Error while getting Processor type", err.Error())
		result.Status = tasks.Error
		result.Summary = "Error while getting Processor type, see debug logs for more details."
		result.URL = architectureRequirementsURL
		return
	}

	//the platform and .NET versions only matter for ARM64, so missing values are tolerated here
	hostInfo, _ := upstream["Base/Env/HostInfo"].Payload.(env.HostInfo)
	dotnetVersions, _ := upstream["DotNetCore/Env/Versions"].Payload.([]string)

	return checkLinuxArch(procType, hostInfo.Platform, dotnetVersions)
}

func checkLinuxArch(procType string, platform string, dotnetVersions []string) (result tasks.Result) {
	procType = strings.ToLower(procType)
	log.Debug("DotNetCoreRequirementsProcessorType - proc type is", procType)

	if procType == "x86_64" || procType == "amd64" {
		result.Status = tasks.Success
		result.Summary = "Processor detected as x64."
		return
	}

	if procType != "aarch64" && procType != "arm64" {
		result.Status = tasks.Failure
		result.Summary = "Processor not detected as x64 or ARM64. The .NET agent only supports x64 and ARM64 processors on Linux."
		result.URL = architectureRequirementsURL
		return
	}

	var warnings []string
	if strings.EqualFold(platform, "alpine") {
		warnings = append(warnings, "The .NET agent does not support ARM64 on Alpine Linux.")
	}

	var preNet5Versions []string
	for _, version := range dotnetVersions {
		if major, _, _, _ := tasks.GetVersionSplit(version); major != -1 && major < 5 {
			preNet5Versions = append(preNet5Versions, version)
		}
	}
	if len(preNet5Versions) > 0 {
		warnings = append(warnings, fmt.Sprintf("The .NET agent only supports ARM64 on .NET 5 or later. Unsupported .NET version(s) detected: %s", strings.Join(preNet5Versions, ", ")))
	}

	if len(warnings) > 0 {
		result.Status = tasks.Warning
		result.Summary = "Processor detected as ARM64. " + strings.Join(warnings, " ")
		result.URL = architectureRequirementsURL
		return
	}

	result.Status = tasks.Success
	result.Summary = "Processor detected as ARM64."
	return
}
