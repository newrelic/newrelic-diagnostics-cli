package docs

// This is an example task referenced in /docs/unit-testing.md

import (
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
)

// BaseLogCount - This struct defined the sample plugin which can be used as a starting point
type BaseLogCount struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseLogCount) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Log/Count")
}

// Explain - Returns the help text for each individual task
func (p BaseLogCount) Explain() string {
	return "Count log files collected."
}

// Dependencies - Returns the dependencies for each task.
func (p BaseLogCount) Dependencies() []string {
	return []string{"Base/Log/Collect"}
}

// Execute - The core work within each task
func (p BaseLogCount) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	// Check if upstream depedency task status is a failure
	if upstream["Base/Log/Collect"].Status == tasks.Failure {
		result = tasks.Result{
			Status:  tasks.None,
			Summary: "There were no log files to count",
		}
		return result
	}

	// type assertion
	logs, ok := upstream["Base/Log/Collect"].Payload.([]log.LogElement)

	// if type assertion failed
	if !ok {
		result = tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
		return result
	}

	logCount := len(logs)

	result = tasks.Result{
		Status:  tasks.Info,
		Payload: logCount,
	}

	// one or more log files found by upstream dependency
	if logCount > 0 {
		result.Summary = strconv.Itoa(logCount) + " log file(s) collected"
	} else {
		//no log files found by upstream dependency
		result.Summary = "No log files collected"
	}

	return result
}
