package requirements

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
)

// DotNetCoreRequirementsOS - This task checks the OS version against the .Net Core Agent requirements
type DotNetCoreRequirementsOS struct {
}

// Identifier - This returns the Category, Subcategory and Name of the task
func (t DotNetCoreRequirementsOS) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Requirements/OS")
}

// Explain - Returns the help text of the task
func (t DotNetCoreRequirementsOS) Explain() string {
	return "Check operating system compatibility with New Relic .NET Core agent"
}

// Dependencies - Returns the dependencies of the task
func (t DotNetCoreRequirementsOS) Dependencies() []string {
	return []string{
		"DotNetCore/Agent/Installed",
		"Base/Env/HostInfo",
	}
}

// Execute - The core work within each task
func (t DotNetCoreRequirementsOS) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Did not detect .Net Core Agent as being installed, skipping this task."
		return
	}

	hostInfo, ok := upstream["Base/Env/HostInfo"].Payload.(env.HostInfo)

	if !ok {
		result.Status = tasks.Error
		result.Summary = "Could not resolve payload of dependent task, HostInfo."
		return
	}

	result = checkOS(hostInfo)
	return
}

const osRequirementsURL = "https://docs.newrelic.com/docs/apm/agents/net-agent/getting-started/net-agent-compatibility-requirements/#operating-system-core"

const architectureRequirementsURL = "https://docs.newrelic.com/docs/apm/agents/net-agent/getting-started/net-agent-compatibility-requirements/#architecture-core"

func checkOS(hostInfo env.HostInfo) (result tasks.Result) {
	if len(hostInfo.PlatformVersion) < 1 {
		result.Status = tasks.Warning
		result.Summary = "Could not detect OS version to check compatibility."
		result.URL = osRequirementsURL
		return
	}

	switch hostInfo.OS {
	case "darwin":
		result.Status = tasks.Failure
		result.Summary = "MacOS is not supported by the .NET Core agent."
		result.URL = osRequirementsURL
		return
	case "linux":
		result = checkLinux(hostInfo.Platform, hostInfo.PlatformVersion)
		return
	case "windows":
		result = checkWindows(hostInfo.PlatformVersion)
		return
	}

	log.Debug("Unable to determine OS.")
	result.Status = tasks.Warning
	result.Summary = "Could not detect OS to check compatibility."
	result.URL = osRequirementsURL
	return
}

func checkWindows(osVersion string) (result tasks.Result) {
	osVerMaj, _, _, _ := tasks.GetVersionSplit(osVersion)

	//OS Major version 6 == Vista or Server 2008
	//OS Major version 10 == Win 10 or Server 2016
	if osVerMaj >= 6 && osVerMaj <= 10 {
		result.Status = tasks.Success
		result.Summary = "OS detected as meeting requirements. See HostInfo task Payload for more info on OS."
		result.URL = osRequirementsURL
		return
	}

	if osVerMaj == -1 {
		result.Status = tasks.Error
		result.Summary = "Unable to get full OS version to check compatibility."
		result.URL = osRequirementsURL
		return
	}

	result.Status = tasks.Failure
	result.Summary = "OS not detected as compatible with the .Net Core Agent."
	result.URL = osRequirementsURL

	return
}

// knownLinuxPlatforms - distributions (as reported by HostInfo) that .NET supports.
// The agent supports any Linux distribution that .NET supports, so we no longer check distribution versions:
// https://learn.microsoft.com/en-us/dotnet/core/install/linux
var knownLinuxPlatforms = map[string]bool{
	"ubuntu":    true,
	"debian":    true,
	"linuxmint": true,
	"opensuse":  true,
	"suse":      true,
	"redhat":    true,
	"fedora":    true,
	"centos":    true,
	"oracle":    true,
	"alpine":    true,
	"amazon":    true,
	"rocky":     true,
	"almalinux": true,
}

func checkLinux(osPlatform string, osVersion string) (result tasks.Result) {
	if !knownLinuxPlatforms[osPlatform] {
		log.Debug("Unknown Linux Platform '" + osPlatform + "'")
		result.Status = tasks.Info
		result.Summary = "Unrecognized Linux platform '" + osPlatform + " " + osVersion + "'. The .NET agent supports any Linux distribution that is supported by .NET; verify that yours is."
		result.URL = osRequirementsURL
		return
	}

	result.Status = tasks.Success
	result.Summary = "OS detected as meeting requirements. See HostInfo task Payload for more info on OS."
	result.URL = osRequirementsURL
	return
}
