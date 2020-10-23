package config

import (
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	logger.Debug("Registering DotNetCore/Config/*")

	registrationFunc(DotNetCoreConfigAgent{}, true)
}
