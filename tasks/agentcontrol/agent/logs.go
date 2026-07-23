package agentcontrol

import (
	"fmt"
	"path/filepath"
	"runtime"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const acWindowsLogDir = `C:\ProgramData\New Relic\newrelic-agent-control\log\agent-control`

// AgentControlLogCollect collects agent-control logs.
// On Linux: reads from the systemd journal via journalctl (agent-control writes to stdout by default).
// On Windows: collects the default log file.
type AgentControlLogCollect struct {
	cmdExec tasks.CmdExecFunc
}

func (p AgentControlLogCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("AgentControl/Agent/Logs")
}

func (p AgentControlLogCollect) Explain() string {
	return "Collect New Relic agent-control log files"
}

func (p AgentControlLogCollect) Dependencies() []string {
	return []string{"AgentControl/Config/Agent"}
}

func (p AgentControlLogCollect) Execute(_ tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["AgentControl/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Agent Control not detected on system.",
		}
	}

	switch runtime.GOOS {
	case "linux":
		return p.collectLinuxLogs()
	case "windows":
		return p.collectWindowsLogs()
	default:
		return tasks.Result{
			Status:  tasks.None,
			Summary: fmt.Sprintf("Log collection for agent-control is not supported on %s.", runtime.GOOS),
		}
	}
}

func (p AgentControlLogCollect) collectLinuxLogs() tasks.Result {
	output, err := p.cmdExec("journalctl", "-u", "newrelic-agent-control", "--no-pager")
	if err != nil {
		log.Debug("journalctl failed: " + err.Error())
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Could not retrieve agent-control logs via journalctl: " + err.Error(),
		}
	}

	stream := make(chan string)
	go tasks.StreamBlob(string(output), stream)

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Collected agent-control logs from journalctl.",
		FilesToCopy: []tasks.FileCopyEnvelope{{
			Path:       "newrelic-agent-control.log",
			Stream:     stream,
			Identifier: "AgentControl/Agent/Logs",
		}},
	}
}

func (p AgentControlLogCollect) collectWindowsLogs() tasks.Result {
	if !tasks.FileExists(acWindowsLogDir) {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("agent-control log directory not found: %s", acWindowsLogDir),
		}
	}

	matches, err := filepath.Glob(filepath.Join(acWindowsLogDir, "*.newrelic-agent-control.log"))
	if err != nil || len(matches) == 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("No agent-control log files found in %s", acWindowsLogDir),
		}
	}

	log.Debug(fmt.Sprintf("Found %d agent-control log file(s) on Windows", len(matches)))
	return tasks.Result{
		Status:      tasks.Success,
		Summary:     fmt.Sprintf("Collected %d agent-control log file(s) from %s.", len(matches), acWindowsLogDir),
		FilesToCopy: tasks.StringsToFileCopyEnvelopes(matches),
	}
}
