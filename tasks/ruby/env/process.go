package env

import (
	"github.com/shirou/gopsutil/process"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RubyEnvProcess - This struct defined the sample plugin which can be used as a starting point
type RubyEnvProcess struct {
}

type rubyPidEnvVars struct {
	Proc    process.Process
	Cwd     string
	EnvVars map[string]string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t RubyEnvProcess) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Env/Process")
}

// Explain - Returns the help text for each individual task
func (t RubyEnvProcess) Explain() string {
	return "Identify running Ruby processes"
}

// Dependencies - Returns the dependencies for ech task.
func (t RubyEnvProcess) Dependencies() []string {
	return []string{
		"Ruby/Config/Agent",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within each task
func (t RubyEnvProcess) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	//Check to ensure the agent was detected
	if upstream["Ruby/Config/Agent"].Status != tasks.Success {
		log.Debug("Ruby/Config/Agent status was not successful")
		return
	}

	//Type assert env vars back out
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Failed to get environment variables from upstream")
	}

	//Get list of running processes
	processes := getRubyProcesses()

	var procs []rubyPidEnvVars
	for _, process := range processes {
		cwd, _ := process.Cwd()
		procs = append(procs, rubyPidEnvVars{Proc: process, Cwd: cwd, EnvVars: envVars})
	}
	result.Payload = procs
	result.Status = tasks.Success

	//Return structure data pairing pid with active env vars

	return
}

func getRubyProcesses() []process.Process {
	processes, err := tasks.FindProcessByName("ruby")
	if err != nil {
		log.Debug("Failed to get list of processes")

	}

	return processes
}
