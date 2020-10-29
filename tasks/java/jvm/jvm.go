package jvm

import (
	"runtime"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Java/JVM/*")

	registrationFunc(JavaJVMVendorsVersions{
		findProcessByName: tasks.FindProcessByName,
		cmdExec:           tasks.CmdExecutor,
		runtimeGOOS:       runtime.GOOS,
		getCmdLineArgs:    getCmdLineArgs,
	}, true)

}
