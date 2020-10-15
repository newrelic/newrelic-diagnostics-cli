package env

import (
	"os"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Node/Env/*")
	registrationFunc(NodeEnvNpmVersion{
		cmdExecutor:      tasks.CmdExecutor,
		npmVersionGetter: getNpmVersion,
	}, true)

	registrationFunc(NodeEnvVersion{
		cmdExec: tasks.CmdExecutor}, true)

	registrationFunc(NodeEnvDependencies{
		cmdExec: tasks.CmdExecutor,
	}, true)

	registrationFunc(NodeEnvVersionCompatibility{
		supportedNodeVersions: []string{"6.0+"},
	}, true)

	registrationFunc(NodeEnvNpmPackage{
		Getwd:      os.Getwd,
		fileFinder: tasks.FindFiles,
	}, true)
}
