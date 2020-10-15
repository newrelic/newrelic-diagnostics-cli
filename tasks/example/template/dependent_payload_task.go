package template

import (
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

// ExampleTemplateDependentPayloadTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateDependentPayloadTask struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateDependentPayloadTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/DependentPayloadTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateDependentPayloadTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for ech task.
func (p ExampleTemplateDependentPayloadTask) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
	}
}

// Execute - The core work within each task
func (p ExampleTemplateDependentPayloadTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	configs, ok := upstream["Java/Config/Agent"].Payload.([]config.ValidateElement)
	if !ok {
		result.Status = tasks.Error
		result.Summary = "Could not resolve payload of dependent task."
		return result
	}

	if len(configs) == 0 {
		result.Status = tasks.None
		result.Summary = "No Java config files found."
	} else if len(configs) == 1 {
		result.Status = tasks.Success
		result.Summary = "Exactly on Java config file found."
	} else {
		result.Status = tasks.Warning
		result.Summary = "Multiple Java config files found, it's not clear which is the correct one."
	}

	return result
}
