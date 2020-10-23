package env

import (
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
)

type supportabilityStatus int

//Constants for use by the supportabilityStatus enum
const (
	NotSupported supportabilityStatus = iota
	Supported
	Unknown
)

// NodeEnvOsCheck
type NodeEnvOsCheck struct {
	upstream map[string]tasks.Result
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvOsCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/OsCheck")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvOsCheck) Explain() string {
	return "This task checks the OS for compatibility with the Node Agent."
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvOsCheck) Dependencies() []string {
	return []string{"Node/Config/Agent", "Base/Env/HostInfo"}
}

// Execute - The core work within each task
func (p NodeEnvOsCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	p.upstream = upstream
	var result tasks.Result

	if upstream["Node/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Node agent not detected"
		return result
	}

	switch isSupported := p.checkOs(upstream); isSupported {
	case Supported:
		result.Status = tasks.Success
		result.URL = `https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent`
		result.Summary = "This OS is compatible with the Node Agent."
	case NotSupported:
		result.Status = tasks.Failure
		result.Summary = "This OS is not compatible with the Node Agent."
		result.URL = `https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent`
	case Unknown:
		result.Status = tasks.Warning
		result.URL = `https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent`
		result.Summary = "Could not determine OS version."
	}
	return result
}

func (p NodeEnvOsCheck) checkOs(upstream map[string]tasks.Result) supportabilityStatus {

	osInfo := p.upstream["Base/Env/HostInfo"].Payload.(env.HostInfo)
	osNameLower := strings.ToLower(osInfo.OS)
	osPlatformVersion := osInfo.PlatformVersion

	if osNameLower == "linux" {
		return Supported
	}

	//Supported OS but missing OS version to determine actual supportability
	if osPlatformVersion == "" && (osNameLower == "windows" || osNameLower == "darwin") {
		return Unknown
	}

	if osNameLower == "windows" {
		if strings.Contains(osInfo.PlatformFamily, "server") {
			re := regexp.MustCompile(`^[0-9.]+`)
			version := re.FindString(osPlatformVersion)
			if compatible, _ := tasks.VersionIsCompatible(version, []string{"6+"}); compatible {
				return Supported
			}
		}
		return NotSupported
	}

	if osNameLower == "darwin" {
		if compatible, _ := tasks.VersionIsCompatible(osPlatformVersion, []string{"10.7+"}); compatible {
			return Supported
		}
	}
	return NotSupported
}
