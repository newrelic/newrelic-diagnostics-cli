package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func gatherProcessInfo() (result tasks.Result) {
	processInfos := []ProcessArgs{}
	processes, err := tasks.GetProcInfoByName("%dotnet%")
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

func buildProcInfo(p tasks.ProcInfoStruct) (procInfo ProcessArgs) {
	// add process info
	procInfo.Pid = int32(p.ProcessId)
	logger.Debug("DotNetCoreEnvProcess pid - ", procInfo.Pid)

	// get command line args
	procInfo.CmdLine = p.CommandLine
	logger.Debug("DotNetCoreEnvProcess cmdLine - ", procInfo.CmdLine)

	// get current working directory
	procInfo.Cwd = "" // ProcInfoStruct has ExecutablePath but that is not CWD
	logger.Debug("DotNetCoreEnvProcess cwd - Not implemented on Windows")

	// get env vars
	envVars, err := tasks.GetProcessEnvVars(procInfo.Pid) // not implemented for windows yet
	if err != nil {
		logger.Debug("Error getting envVars of process: ", err.Error())
	}
	// use the default filter
	procInfo.EnvVars = envVars.WithDefaultFilter()
	logger.Debug("DotNetCoreEnvProcess filtered envVars - ", procInfo.EnvVars)

	return

}
