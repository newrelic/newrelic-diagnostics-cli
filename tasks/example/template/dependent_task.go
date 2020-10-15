package template

import (
	"github.com/newrelic/NrDiag/tasks"
)

// ExampleTemplateDependentTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateDependentTask struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateDependentTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/DependentTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateDependentTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateDependentTask) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
	}
}

// Execute - The core work within each task
func (p ExampleTemplateDependentTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	if upstream["Java/Config/Agent"].Status == tasks.Success {
		result.Status = tasks.Success
		result.Summary = "Looks like Java is installed!"
	} else {
		result.Status = tasks.None
		result.Summary = "Looks like Java is not installed!"
	}

	return result
}
