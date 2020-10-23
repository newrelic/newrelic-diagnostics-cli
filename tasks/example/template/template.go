package template

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Example/Template/*")

	// every task in a package needs to be registered to run
	// if you pass "false" as the second parameter it will be runnable, but not included by default
	registrationFunc(ExampleTemplateFullTask{}, false)
	registrationFunc(ExampleTemplateMinimalTask{}, false)
	registrationFunc(ExampleTemplateDependentTask{}, false)
	registrationFunc(ExampleTemplateDependentPayloadTask{}, false)
	registrationFunc(ExampleTemplateCustomPayloadTask{}, false)
	registrationFunc(ExampleTemplateCustomPayloadJSONTask{}, false)
	registrationFunc(ExampleTemplateCopyFilesTask{}, false)
	registrationFunc(ExampleTemplateInfoTask{}, false)
}
