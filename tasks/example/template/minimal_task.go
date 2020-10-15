package template

import (
	"github.com/newrelic/NrDiag/tasks"
)

// ExampleTemplateMinimalTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateMinimalTask struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateMinimalTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/MinimalTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateMinimalTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateMinimalTask) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p ExampleTemplateMinimalTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	return result
}
