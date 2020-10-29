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

func checkOS(hostInfo env.HostInfo) (result tasks.Result) {
	if len(hostInfo.PlatformVersion) < 1 {
		result.Status = tasks.Warning
		result.Summary = "Could not detect OS version to check compatibility."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	switch hostInfo.OS {
	case "darwin":
		result.Status = tasks.Failure
		result.Summary = "MacOS is not supported by the .NET Core agent."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
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
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
	return
}

func checkWindows(osVersion string) (result tasks.Result) {
	osVerMaj, _, _, _ := tasks.GetVersionSplit(osVersion)

	//OS Major version 6 == Vista or Server 2008
	//OS Major version 10 == Win 10 or Server 2016
	if osVerMaj >= 6 && osVerMaj <= 10 {
		result.Status = tasks.Success
		result.Summary = "OS detected as meeting requirements. See HostInfo task Payload for more info on OS."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	if osVerMaj == -1 {
		result.Status = tasks.Error
		result.Summary = "Unable to get full OS version to check compatibility."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	result.Status = tasks.Failure
	result.Summary = "OS not detected as compatible with the .Net Core Agent."
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"

	return
}

func checkLinux(osPlatform string, osVersion string) (result tasks.Result) {
	checkPassed := false
	osVerMaj, osVerMin, _, _ := tasks.GetVersionSplit(osVersion)

	if osVerMaj == -1 {
		result.Status = tasks.Error
		result.Summary = "Unable to get full OS version to check compatibility."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	switch osPlatform {
	case "ubuntu":
		checkPassed = checkUbuntu(osVerMaj, osVerMin)
	case "debian":
		checkPassed = checkDebian(osVerMaj, osVerMin)
	case "linuxmint":
		checkPassed = checkMint(osVerMaj)
	case "opensuse":
		checkPassed = checkOpenSuse(osVerMaj, osVerMin)
	case "suse":
		checkPassed = checkSles(osVerMaj)
	case "redhat":
		checkPassed = checkRhel(osVerMaj)
	case "fedora":
		checkPassed = checkFedora(osVerMaj)
	case "centos":
		checkPassed = checkCentOs(osVerMaj)
	case "oracle":
		checkPassed = checkOracle(osVerMaj)
	default:
		log.Debug("Unknown Linux Platform '" + osPlatform + "'")
		result.Status = tasks.Warning
		result.Summary = "Unknown Linux Platform '" + osPlatform + "'. Unable to check compatability."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	if checkPassed == true {
		result.Status = tasks.Success
		result.Summary = "OS detected as meeting requirements. See HostInfo task Payload for more info on OS."
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
		return
	}

	result.Status = tasks.Failure
	result.Summary = "OS not detected as compatible with the .Net Core Agent. See HostInfo task Payload for more info on OS."
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-core-20-agent#operating-system"
	return
}

/*
	supported versions: Comment last updated: 3/1/2018
	- https://github.com/dotnet/core/blob/master/release-notes/2.0/2.0-supported-os.md#linux
	- Ubuntu: 17.10, 16.04, 14.04
	- Mint: 18, 17
	- Debian: 9, 8.7+
	- openSUSE: 42.2+
	- SLES: 	12
	- Red Hat Enterprise Linux: 7
	- CentOS: 7
	- Oracle Linux: 7
	- Fedora: 26, 27
*/

func checkUbuntu(osVerMaj int, osVerMin int) (retVal bool) {
	// Ubuntu: 17.10, 16.04, 14.04
	if osVerMaj == 14 || osVerMaj == 16 {
		if osVerMin == 04 {
			retVal = true
			return
		}
	}
	if osVerMaj == 17 {
		if osVerMin == 10 {
			retVal = true
			return
		}
	}
	retVal = false
	return
}

func checkDebian(osVerMaj int, osVerMin int) (retVal bool) {
	// Debian: 9, 8.7+
	if osVerMaj == 9 {
		retVal = true
		return
	}
	if osVerMaj == 8 {
		if osVerMin >= 7 {
			retVal = true
			return
		}
	}

	retVal = false
	return
}

func checkMint(osVerMaj int) (retVal bool) {
	// Mint: 18, 17
	if osVerMaj == 18 || osVerMaj == 17 {
		retVal = true
		return
	}

	retVal = false
	return
}

func checkOpenSuse(osVerMaj int, osVerMin int) (retVal bool) {
	// openSUSE: 42.2+
	if osVerMaj == 42 {
		if osVerMin >= 2 {
			retVal = true
			return
		}
	}

	retVal = false
	return
}

func checkSles(osVerMaj int) (retVal bool) {
	// SLES: 12
	if osVerMaj == 12 {
		retVal = true
		return
	}

	retVal = false
	return
}

func checkRhel(osVerMaj int) (retVal bool) {
	// Red Hat Enterprise Linux: 7
	if osVerMaj == 7 {
		retVal = true
		return
	}
	retVal = false
	return
}

func checkCentOs(osVerMaj int) (retVal bool) {
	// CentOS: 7
	if osVerMaj == 7 {
		retVal = true
		return
	}
	retVal = false
	return
}

func checkOracle(osVerMaj int) (retVal bool) {
	// Oracle Linux: 7
	if osVerMaj == 7 {
		retVal = true
		return
	}
	retVal = false
	return
}

func checkFedora(osVerMaj int) (retVal bool) {
	// Fedora: 26, 27
	if osVerMaj == 26 || osVerMaj == 27 {
		retVal = true
		return
	}
	retVal = false
	return
}
