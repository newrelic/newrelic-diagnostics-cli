//go:build windows
// +build windows

package profiler

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

const envNetAgentRegPath = `System\CurrentControlSet\Control\Session Manager\Environment`
const nrNetAgentInstallPathValue = `C:\Program Files\New Relic\.NET Agent\`
const nrNetAgentCurrentClsid = `{71DA0A04-7777-4EC6-9643-7D28B46A8A41}`
const nrNetAgentOldClsid = `{FF68FEB9-E58A-4B75-A2B8-90CE7D915A26}`

type DotNetProfilerEnvVarKey struct {
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetProfilerEnvVarKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/EnvVarKey")
}

// Explain - Returns the help text for each individual task
func (p DotNetProfilerEnvVarKey) Explain() string {
	return "Validate environment variables required for system wide/non-IIS .NET application profiling"
}

// Dependencies - This task has no dependencies
func (p DotNetProfilerEnvVarKey) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute -  This does not take options
func (p DotNetProfilerEnvVarKey) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
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

	return validateEnvInstrumentationKeys()

}

// queries the registry for env keys and validates them against expected standards
func validateEnvInstrumentationKeys() (result tasks.Result) {

	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, envNetAgentRegPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug("Error opening Environment Reg Key. Error = ", err.Error())
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Could not open Environment Reg Key " + err.Error(),
		}
	}

	defer regKey.Close()
	regValue, _, regErr := regKey.GetStringValue("COR_ENABLE_PROFILING")

	profilingEnabled := false
	installPathSet := false
	clsidSet := false
	if regErr != nil {
		log.Debug("Error opening COR_ENABLE_PROFILING Reg Sub Key. Error = ", regErr.Error())

	} else if regValue == `1` {

		profilingEnabled = true
	}

	regValue, _, regErr = regKey.GetStringValue("NEWRELIC_INSTALL_PATH")

	if regErr != nil {
		log.Debug("Error opening NEWRELIC_INSTALL_PATH Reg Sub Key. Error = ", regErr.Error())
	} else if regValue == nrNetAgentInstallPathValue {
		installPathSet = true
	}

	regValue, _, regErr = regKey.GetStringValue("COR_PROFILER")

	if regErr != nil {
		log.Debug("Error opening COR_PROFILER Reg Sub Key. Error = ", regErr.Error())
	} else if regValue == nrNetAgentOldClsid || regValue == nrNetAgentCurrentClsid {
		clsidSet = true
	}

	if profilingEnabled && installPathSet && clsidSet {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Environment variables needed for system wide/non-IIS .NET application profiling are correctly set",
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: "Some of these environment variables needed for system wide/non-IIS .NET application profiling are not set:\nCOR_ENABLE_PROFILING=1\n" + `NEWRELIC_INSTALL_PATH=C:\Program Files\New Relic\.NET Agent` + "\nCOR_PROFILER={XXXXXXXXXX}\n" + `If you are attempting to instrument a non-IIS .NET framework application, re-run the .NET agent installer and select the "Instrument all .NET framework applications" option on the "Custom Setup" screen.`,
		URL:     "https://docs.newrelic.com/docs/agents/net-agent/installation/install-net-agent-windows#enabling-the-agent",
	}

}
