package requirements

import (
	"errors"
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/env"
)

// ExampleTemplateMinimalTask - This task checks the OS version against the .Net Agent requirements
type DotnetRequirementsOS struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotnetRequirementsOS) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/OS")
}

// Explain - Returns the help text for each individual task
func (p DotnetRequirementsOS) Explain() string {
	return "Check operating system compatibility with New Relic .NET agent"
}

// Dependencies - Returns the dependencies for ech task.
func (p DotnetRequirementsOS) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
		"Base/Env/HostInfo",
	}
}

// Execute - The core work within each task
func (p DotnetRequirementsOS) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Did not detect .Net Agent as being installed, this check did not run",
		}
	}
	//add check for agent installed
	hostInfo, ok := upstream["Base/Env/HostInfo"].Payload.(env.HostInfo)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Could not resolve payload of dependent task, HostInfo.",
		}
	}

	if len(hostInfo.PlatformVersion) < 1 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Could not get OS version to check compatibility",
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#operating-system",
		}
	}

	if strings.Contains(hostInfo.PlatformVersion, "5.2") && strings.Contains(strings.ToLower(hostInfo.PlatformFamily), "server") {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "OS detected as Windows Server 2003. Last .Net Agent version compatible is 6.11.613",
			URL:     "https://docs.newrelic.com/docs/release-notes/agent-release-notes/net-release-notes/net-agent-612610",
		}

	}

	//OS Major version 6 == Vista or Server 2008
	//OS Major version 10 == Win 10 or Server 2016
	compatibleOs := []string{"6-10.*"}
	osVersionSplit := strings.Split(hostInfo.PlatformVersion, ".")
	var supported bool
	var supportedErr error
	if len(osVersionSplit) > 1 {
		// we only need the first 2 elements here
		osVersion := osVersionSplit[0] + "." + osVersionSplit[1]
		supported, supportedErr = tasks.VersionIsCompatible(osVersion, compatibleOs)
	} else {
		supportedErr = errors.New("Error parsing version: " + hostInfo.PlatformVersion)
	}

	if supportedErr != nil {
		log.Debug("Error parsing version: ", supportedErr)
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Encountered issues getting full OS version to check compatibility",
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#operating-system",
		}
	}
	if supported {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "OS detected as meeting requirements. See HostInfo task Payload for more info on OS",
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#operating-system",
		}
	}

	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "OS not detected as compatible with the .Net Framework Agent",
		URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#operating-system",
	}

}
