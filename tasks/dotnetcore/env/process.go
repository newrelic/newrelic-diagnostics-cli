package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetCoreEnvProcess - This task gathers information on any running dotnet processes
type DotNetCoreEnvProcess struct {
}

// Identifier - This returns the Category, Subcategory and Name of the task
func (t DotNetCoreEnvProcess) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Env/Process")
}

// Explain - Returns the help text of the task
func (t DotNetCoreEnvProcess) Explain() string {
	return "Collect .NET Core process command line options"
}

// Dependencies - Returns the dependencies of the task
func (t DotNetCoreEnvProcess) Dependencies() []string {
	return []string{
		"DotNetCore/Agent/Installed",
		"DotNetCore/Env/Versions",
	}
}

// ProcessArgs - contains .net core process information
type ProcessArgs struct {
	Pid     int32
	CmdLine string
	Cwd     string
	EnvVars map[string]string
}

// Execute - The core work within each task
func (t DotNetCoreEnvProcess) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	if upstream["DotNetCore/Env/Versions"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Did not detect .Net Core as being installed, skipping this task."
		logger.Debug("Did not detect .Net Core as being installed, skipping this task.")
		return
	}

	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Did not detect .Net Core Agent as being installed, skipping this task."
		logger.Debug("Did not detect .Net Core Agent as being installed, skipping this task.")

		return
	}

	result = gatherProcessInfo()
	return
}
