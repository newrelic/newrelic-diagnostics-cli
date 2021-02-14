package w3wp

import (
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type DotNetW3wpCollect struct {
	name string
}

func (p DotNetW3wpCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/W3wp/Collect")
}

func (p DotNetW3wpCollect) Explain() string {
	return "Collect list of W3wp processes" //This is the customer visible help text that describes what this particular task does
}

func (p DotNetW3wpCollect) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
		"Base/Env/CheckWindowsAdmin",
	}
}

func (p DotNetW3wpCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{}
	adminPerms := true // assume admin permissions, set to false later if needed

	// get results of Base/Env/CheckWindowsAdmin to see what permissions we have
	if upstream["Base/Env/CheckWindowsAdmin"].Status == tasks.Warning {
		adminPerms = false // no admin permissions
	}

	// get results from DotNet/Agent/IsInstalled and make sure status wasn't Failure
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

	w3wpProcesses, err := tasks.FindProcessByName("w3wp.exe")
	if err != nil {
		if !adminPerms {
			result.Status = tasks.Error
			result.Summary = " Unable to check for w3wp.exe processes due to permissions. If possible re-run from an Admin cmd prompt or PowerShell."
			result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
			return result
		} else {
			result.Status = tasks.Warning
			logger.Debug("Unknown error:", err)
			logger.Debug("w3wpProcesses:", w3wpProcesses)
			result.Summary = " There was an unknown error while checking for w3wp.exe processes. Please retry with the '-v' option and send the output to support."
			result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
			return result
		}
	}

	numOfProcesses := len(w3wpProcesses)

	if numOfProcesses == 0 {
		if !adminPerms {
			result.Status = tasks.Error
			result.Summary = " Unable to check for w3wp.exe processes due to permissions. If possible re-run from an Admin cmd prompt or PowerShell."
			result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics#windows-run"
			return result
		} else {
			result.Status = tasks.Warning
			result.Summary = " The .NET Agent is installed, but there are no w3wp.exe processes running. Are your apps in IIS and receiving traffic?"
			result.URL = "https://docs.newrelic.com/docs/agents/net-agent/installation-configuration/install-net-agent"
			return result
		}
	}
	result.Status = tasks.Success
	result.Payload = w3wpProcesses
	logger.Debug(" There are ", numOfProcesses, " w3wp.exe processes running with the following PIDs:")
	for _, proc := range w3wpProcesses {
		logger.Debug(proc.Pid)
	}

	return result
}
