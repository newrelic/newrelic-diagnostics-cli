package env

import (
	"github.com/shirou/gopsutil/process"
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func gatherProcessInfo() (result tasks.Result) {
	processInfos := []ProcessArgs{}
	processes, err := tasks.FindProcessByName("dotnet")
	if err != nil {
		logger.Debug("DotNetCoreEnvProcess - Error finding dotnet process: ", err.Error())
		result.Status = tasks.None
		result.Summary = "Error finding dotnet process."
		return
	}

	if len(processes) == 0 {
		result.Status = tasks.None
		result.Summary = "No dotnet processes detected."
		logger.Debug("DotNetCoreEnvProcess - No dotnet processes detected.")
		return
	}

	for _, p := range processes {
		procInfo := buildProcInfo(p)

		processInfos = append(processInfos, procInfo)
	}

	if len(processInfos) == 0 {
		result.Status = tasks.None
		result.Summary = "No dotnet processes detected."
		return
	}

	result.Status = tasks.Success
	result.Summary = "Gathered command line options of running dotnet processes."
	result.Payload = processInfos

	return
}

func buildProcInfo(p process.Process) (procInfo ProcessArgs) {
	// add process info
	procInfo.Pid = p.Pid
	logger.Debug("DotNetCoreEnvProcess pid - ", p.Pid)

	// get command line args
	cmdLine, err := p.Cmdline()
	if err != nil {
		logger.Debug("Error getting command line args: ", err.Error())
	}
	procInfo.CmdLine = cmdLine
	logger.Debug("DotNetCoreEnvProcess cmdLine - ", cmdLine)

	// get current working directory
	cwd, err := p.Cwd()
	if err != nil {
		logger.Debug("Error getting cwd of process: ", err.Error())
	}
	procInfo.Cwd = cwd
	logger.Debug("DotNetCoreEnvProcess cwd - ", cwd)

	// get env vars
	envVars, err := tasks.GetProcessEnvVars(p.Pid)
	if err != nil {
		logger.Debug("Error getting envVars of process: ", err.Error())
	}
	// use the default filter
	procInfo.EnvVars = envVars.WithDefaultFilter()
	logger.Debug("DotNetCoreEnvProcess filtered envVars - ", procInfo.EnvVars)

	return

}
