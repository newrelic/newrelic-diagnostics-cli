package template

import (
	"github.com/newrelic/NrDiag/tasks"
)

// ExampleTemplateInfoTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateInfoTask struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateInfoTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/InfoTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateInfoTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateInfoTask) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p ExampleTemplateInfoTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.Info,
		Summary: "I succeeded in doing nothing.",
		Payload: "The return data that I want to store",
	}

	return result
}
