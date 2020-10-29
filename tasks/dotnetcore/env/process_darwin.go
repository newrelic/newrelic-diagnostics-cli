package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func gatherProcessInfo() (result tasks.Result) {
	result.Status = tasks.None
	result.Summary = "The .NET Core agent is not supported on Mac, skipping this task."

	return
}
