package daemon

import (
	"fmt"

	"github.com/shirou/gopsutil/process"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// PHPDaemonRunning - Retrieve status of newrelic-daemon processes for PHP Agent
type PHPDaemonRunning struct {
	processFinder     processFinderFunc
	fileExistsChecker fileExistsCheckerFunc
}

type processFinderFunc func(string) ([]process.Process, error)
type fileExistsCheckerFunc func(string) bool

type PHPDaemonInfo struct {
	Mode         string
	ProcessCount int
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p PHPDaemonRunning) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("PHP/Daemon/Running")
}

// Explain - Returns the help text for each individual task
func (p PHPDaemonRunning) Explain() string {
	return "Identify running New Relic PHP daemon process"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well
func (p PHPDaemonRunning) Dependencies() []string {
	return []string{
		"PHP/Config/Agent",
	}
}

// Execute - The core work within each task
func (p PHPDaemonRunning) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["PHP/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "PHP Agent was not detected on this host. Skipping daemon detection.",
		}
	}

	daemonInfo, err := p.getDaemonInfo()
	if err != nil {
		log.Debug("Error getting process", err)
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Error detecting newrelic-daemon process: %s", err.Error()),
		}
	}

	//number of daemons != 2 which is a failure state. Display action message depending on daemon mode.
	if daemonInfo.ProcessCount != 2 {
		baseFailureSummaryMsg := fmt.Sprintf("There is incorrect number of newrelic-daemon processes running - (%v).", daemonInfo.ProcessCount)

		if daemonInfo.Mode == "manual" {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: fmt.Sprintf("%s Please make sure the daemons were started as outlined in the following document:", baseFailureSummaryMsg),
				Payload: daemonInfo,
				URL:     "https://docs.newrelic.com/docs/agents/php-agent/advanced-installation/starting-php-daemon-advanced",
			}
		}

		// agent daemon mode
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("%s Please restart your web server to start up the daemons.", baseFailureSummaryMsg),
			Payload: daemonInfo,
		}
	}

	/* newrelic-daemon processes found - there should be two, a watcher and a worker.
	only one process is a failure state, more that 2 as well.*/
	log.Debug("Found two newrelic-daemon processes!")
	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("Two New Relic daemon processes found in %s mode.", daemonInfo.Mode),
		Payload: daemonInfo,
	}

}

func (p PHPDaemonRunning) getDaemonInfo() (PHPDaemonInfo, error) {

	daemonInfo := PHPDaemonInfo{}

	daemonProcesses, err := p.processFinder("newrelic-daemon")
	if err != nil {
		return PHPDaemonInfo{}, err
	}

	daemonInfo.ProcessCount = len(daemonProcesses)

	// https://docs.newrelic.com/docs/agents/php-agent/configuration/proxy-daemon-newreliccfg-settings#proxy-settings
	if p.fileExistsChecker("/etc/newrelic/newrelic.cfg") {
		daemonInfo.Mode = "manual"
	} else {
		daemonInfo.Mode = "agent"
	}

	return daemonInfo, nil

}
