package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseEnvCollectSysProps - This struct defined the sample plugin which can be used as a starting point
type BaseEnvCollectSysProps struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseEnvCollectSysProps) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/CollectSysProps")
}

// Explain - Returns the help text for each individual task
func (t BaseEnvCollectSysProps) Explain() string {
	return "Collect new relic system properties for running JVM processes"
}

// Dependencies - Returns the dependencies for ech task.
func (t BaseEnvCollectSysProps) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BaseEnvCollectSysProps) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	systemProps := tasks.GetNewRelicSystemProps()

	if len(systemProps) > 0 {
		return tasks.Result{
			Status:  tasks.Info,
			Summary: "Successfully collected some new relic system properties",
			Payload: systemProps, //[]ProcIDSysProps
		}
	}
	return tasks.Result{
		Status:  tasks.None,
		Summary: "No new relic system properties were found",
	}
}
