package env

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"os"
)

// RegisterWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Env/*")

	registrationFunc(DotNetEnvVersions{}, true)
	registrationFunc(DotNetEnvTargetVersion{
		osGetwd:            os.Getwd,
		findFiles:          tasks.FindFiles,
		returnStringInFile: tasks.ReturnStringSubmatchInFile,
	}, true)
}
