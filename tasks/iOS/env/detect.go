package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// iOSEnvDetect - This struct defined the sample plugin which can be used as a starting point
type iOSEnvDetect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p iOSEnvDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("iOS/Env/Detect")
}

// Explain - Returns the help text for each individual task
func (p iOSEnvDetect) Explain() string {
	return "Detect if running in iOS environment"
}

// Dependencies - Returns the dependencies for each task.
func (p iOSEnvDetect) Dependencies() []string {
	return []string{"Base/Config/Collect"}
}

// Execute - The core work within each task
func (p iOSEnvDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "iOS environment not detected",
	}

	if upstream["Base/Config/Collect"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Android environment not detected",
		}
	}

	configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	for _, value := range configs {
		if value.FileName == "AppDelegate.swift" {
			result.Status = tasks.Info
			result.Summary = "iOS environment detected"
		}
	}
	return result
}
