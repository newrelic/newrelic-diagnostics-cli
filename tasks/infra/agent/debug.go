package agent

import (
	"fmt"
	"time"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"

	pb "gopkg.in/cheggaaa/pb.v1"
)

//Time we wait for debug log collection in minutes
var debugLoggingTimeoutMin = 3

//https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/new-relic-infrastructure-agent-140
var ctlVerRequirementLinux tasks.Ver = tasks.Ver{
	Major: 1,
	Minor: 4,
	Patch: 0,
	Build: 0,
}

//https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/new-relic-infrastructure-agent-170-0
var ctlVerRequirementWindows tasks.Ver = tasks.Ver{
	Major: 1,
	Minor: 7,
	Patch: 0,
	Build: 0,
}

// InfraAgentDebug - This struct defines the Infrastructure agent version task
type InfraAgentDebug struct {
	blockWithProgressbar func(int)
	cmdExecutor          func(string, ...string) ([]byte, error)
	runtimeOS            string
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
		"Base/Log/Copy",
		"Infra/Agent/Version",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within each task
func (p InfraAgentDebug) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	// Because this task doesn't run by default and customers will need deliberately request it be run, we don't want this task
	// to fail silently (None result) when there is an upstream dependency issue.
	if upstream["Base/Log/Copy"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "No New Relic Infrastructure log files detected. If your log files are in a custom location, re-run the " + tasks.ThisProgramFullName + " after setting the NRIA_LOG_FILE environment variable.",
			URL:     "https://docs.newrelic.com/docs/infrastructure/install-configure-manage-infrastructure/configuration/infrastructure-configuration-settings#log-file",
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
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	var targetInfraVersion tasks.Ver
	if p.runtimeOS == "windows" {
		targetInfraVersion = ctlVerRequirementWindows
	} else {
		targetInfraVersion = ctlVerRequirementLinux
	}

	ctlSupported := infraVersion.IsGreaterThanEq(targetInfraVersion)
	if !ctlSupported {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Infrastructure debug CTL binary not available in detected version of Infrastructure Agent(%s). Minimum required Infrastructure Agent version is: %s", infraVersion.String(), targetInfraVersion.String()),
			URL:     "https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes",
		}
	}
	
	infraCtlCmd := "newrelic-infra-ctl"

	//For windows we have determine exact location of binary, using environment variables
	if p.runtimeOS == "windows" {
		if upstream["Base/Env/CollectEnvVars"].Status != tasks.Info {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: "Unable to determine path of New Relic Infrastructure CTL binary via environment variables: upstream dependency failure",
			}
		}

		envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: "Task did not meet requirements necessary to run: type assertion failure",
			}
		}

		programFilesPath, ok := envVars["ProgramFiles"]
		if !ok {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: "Unable to determine path of New Relic Infrastructure CTL binary via environment variables: ProgramFiles environment variable not",
			}
		}

		infraCtlCmd = programFilesPath + `\New Relic\newrelic-infra\newrelic-infra-ctl.exe`
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
			URL:     "https://docs.newrelic.com/docs/infrastructure/install-configure-manage-infrastructure/manage-your-agent/troubleshoot-running-agent",
		}
	}

	log.Debugf("newrelic-infra-ctl output: %s", string(cmdOutBytes))

	log.Infof("Debug logging enabled. Collecting logs for %d minutes...\n", debugLoggingTimeoutMin)

	p.blockWithProgressbar(debugLoggingTimeoutMin)

	log.Info("Done!\n")

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Successfully enabled New Relic Infrastructure debug logging with newrelic-infra-ctl",
	}
}


func blockWithProgressbar(minutes int) {
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

	ticker := time.NewTicker(time.Second)

	for i := 0; i < seconds; i++ {
		select {
		case <-ticker.C:
			bar.Increment()
		}
	}
	ticker.Stop()
}
