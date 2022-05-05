//go:build linux || darwin
// +build linux darwin

package jvm

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterNixWith - will register any plugins in this package
func RegisterNixWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Java/JVM/*_nix")

	registrationFunc(JavaJVMPermissions{}, true)
}
