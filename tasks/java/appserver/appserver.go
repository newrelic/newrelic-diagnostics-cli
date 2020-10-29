package appserver

import (
	"os"
	"runtime"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Java/Appserver/*")

	registrationFunc(JavaAppServerWebSphere{
		osGetwd:             os.Getwd,
		osGetExecutable:     os.Executable,
		fileFinder:          tasks.FindFiles,
		findStringInFile:    tasks.FindStringInFile,
		returnSubstring:     tasks.ReturnLastStringSubmatchInFile,
		versionIsCompatible: tasks.VersionIsCompatible,
	}, true)
	registrationFunc(JavaAppserverJbossEapCheck{
		runtimeOs:             runtime.GOOS,
		fileExists:            tasks.FileExists,
		returnSubstringInFile: tasks.ReturnLastStringSubmatchInFile,
		findFiles:             tasks.FindFiles,
	}, true)
	registrationFunc(JavaAppserverJBossAsCheck{
		getCmdline:            getCmdlineFromProcess,
		findFiles:             tasks.FindFiles,
		findProcessByName:     tasks.FindProcessByName,
		returnSubstringInFile: tasks.ReturnLastStringSubmatchInFile,
	}, true)
}
