package requirements

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Python/Requirements/*")

	registrationFunc(PythonRequirementsPythonVersion{}, true)
	registrationFunc(PythonRequirementsWebframework{}, true)
	registrationFunc(PythonRequirementsPythonVersion{}, true)

}
