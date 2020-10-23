package agent

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// AndroidAgentDetect - This struct defined the sample plugin which can be used as a starting point
type AndroidAgentDetect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p AndroidAgentDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Android/Agent/Detect")
}

// Explain - Returns the help text for each individual task
func (p AndroidAgentDetect) Explain() string {
	return "Detect if running in Android environment"
}

// Dependencies - Returns the dependencies for ech task.
func (p AndroidAgentDetect) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
	}
}

// Execute - Returns success if AndroidManifest.xml file is found in Base/Config/Collect
func (p AndroidAgentDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

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
			Summary: "Error occurred reading the Base/Config/Collect payload",
		}
	}

	for _, value := range configs {
		if value.FileName == "AndroidManifest.xml" {
			return tasks.Result{
				Status:  tasks.Info,
				Summary: "Android environment detected",
			}
		}
	}
	return tasks.Result{
		Status:  tasks.None,
		Summary: "Android environment not detected",
	}
}
