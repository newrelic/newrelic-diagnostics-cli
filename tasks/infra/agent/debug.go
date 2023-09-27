package agent

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"

	pb "gopkg.in/cheggaaa/pb.v1"
)

var debugLoggingTimeoutMin = 5
var infraLogConfigDocLink = "https://docs.newrelic.com/docs/infrastructure/install-infrastructure-agent/configuration/infrastructure-agent-configuration-settings/#log"

// InfraAgentDebug - This struct defines the Infrastructure agent version task
type InfraAgentDebug struct {
	blockWithProgressbar func(int) (bool, int)
	cmdExecutor          func(string, ...string) ([]byte, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraAgentDebug) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Agent/Debug")
}

// Explain - Returns the help text for each individual task
func (p InfraAgentDebug) Explain() string {
	return "Dynamically enable New Relic Infrastructure agent debug logging by running newrelic-infra-ctl"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p InfraAgentDebug) Dependencies() []string {
	return []string{
		"Infra/Log/Collect",
		"Infra/Agent/Version",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within each task
func (p InfraAgentDebug) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Log/Collect"].Status == tasks.Error || upstream["Infra/Log/Collect"].Status == tasks.Failure {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "Failed to collect New Relic Infrastructure log files. Review the output of the Infra/Log/Collect task for more information.",
			URL:     infraLogConfigDocLink,
		}
	}

	if upstream["Infra/Log/Collect"].Status == tasks.None || !upstream["Infra/Log/Collect"].HasPayload() {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "No New Relic Infrastructure log files detected. If your log files are in a custom location, re-run the " + tasks.ThisProgramFullName + " after setting the NRIA_LOG_FILE environment variable.",
			URL:     infraLogConfigDocLink,
		}
	}

	infraLogs, ok := upstream["Infra/Log/Collect"].Payload.([]string)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	if len(infraLogs) < 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "No New Relic Infrastructure log files detected. If your log files are in a custom location, re-run the " + tasks.ThisProgramFullName + " after setting the NRIA_LOG_FILE environment variable.",
			URL:     infraLogConfigDocLink,
		}
	}

	if upstream["Infra/Agent/Version"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "New Relic Infrastructure agent not detected on system. Please ensure the Infrastructure agent is installed and running.",
		}
	}

	infraVersion, ok := upstream["Infra/Agent/Version"].Payload.(tasks.Ver)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	ctlSupported := infraVersion.IsGreaterThanEq(InfraCtlVerRequirement)
	if !ctlSupported {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Infrastructure debug CTL binary not available in detected version of Infrastructure Agent(%s). Minimum required Infrastructure Agent version is: %s", infraVersion.String(), InfraCtlVerRequirement.String()),
			URL:     "https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/",
		}
	}

	infraCtlCmd, err := GetInfraCtlCmd(upstream)
	if err != nil {
		if err.Error() == tasks.AssertionErrorSummary {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Unable to determine path of New Relic Infrastructure CTL binary via environment variables: %s", err.Error()),
		}

	}

	// Sleep here to give stdout priority to go func handling task output to avoid printing the following line
	// before dependency task result. E.g:
	// 		Enabling debug level logging for New Relic Infrastructure agent...
	// 		Success  Infra/Config/Version
	time.Sleep(time.Millisecond)

	log.Info("\nEnabling debug level logging for New Relic Infrastructure agent...")

	cmdOutBytes, err := p.cmdExecutor(infraCtlCmd)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Error executing newrelic-infra-ctl: %s\n\t%s", err.Error(), string(cmdOutBytes)),
			URL:     "https://docs.newrelic.com/docs/infrastructure/install-infrastructure-agent/manage-your-agent/troubleshoot-running-infrastructure-agent/",
		}
	}

	log.Debugf("newrelic-infra-ctl output: %s", string(cmdOutBytes))

	log.Infof("Debug logging enabled. Collecting logs for %d minutes... (Ctrl-C to end early)\n", debugLoggingTimeoutMin)

	interrupted, interruptedSeconds := p.blockWithProgressbar(debugLoggingTimeoutMin)

	if interrupted {
		log.Info("Log collection interrupted.\n")
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl, however the collection was interrupted after %s", time.Duration(interruptedSeconds*int(time.Second))),
		}
	}

	log.Info("Done!\n")

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl",
	}
}

func blockWithProgressbar(minutes int) (bool, int) {
	seconds := minutes * 60
	bar := pb.New(seconds)
	bar.ShowTimeLeft = false
	bar.ShowCounters = false
	bar.ShowElapsedTime = true
	bar.SetRefreshRate(time.Second)
	bar.SetUnits(pb.U_DURATION)
	bar.Width = 60

	bar.Start()
	defer bar.Finish()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for i := 0; i < seconds; i++ {
		select {
		case <-ticker.C:
			bar.Increment()
		case <-ctx.Done():
			stop()
			ticker.Stop()
			return true, i
		}
	}
	return false, 0
}
