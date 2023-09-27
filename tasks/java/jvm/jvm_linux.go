package jvm

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterLinuxWith - will register any plugins in this package
func RegisterLinuxWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Java/JVM/*_nix")

	registrationFunc(JavaJVMPermissions{}, true)
}
