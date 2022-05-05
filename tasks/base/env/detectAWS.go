package env

import (
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseEnvDetectAWS - This struct defined the sample plugin which can be used as a starting point
type BaseEnvDetectAWS struct {
	httpGetter tasks.HTTPRequestFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvDetectAWS) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/DetectAWS")
}

// Explain - Returns the help text for each individual task
func (p BaseEnvDetectAWS) Explain() string {
	return "Detect if running in AWS environment"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseEnvDetectAWS) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p BaseEnvDetectAWS) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if p.checkAWSMetaData() {
		result = tasks.Result{
			Status:  tasks.Success,
			Summary: "Successfully detected AWS.",
		}
	} else {
		result = tasks.Result{
			Status:  tasks.None,
			Summary: "AWS metadata endpoint timeout",
		}
	}

	return result
}

func (p BaseEnvDetectAWS) checkAWSMetaData() bool {
	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            "http://169.254.169.254/latest/meta-data",
		TimeoutSeconds: 2,
		BypassProxy: true,
	}
	resp, err := p.httpGetter(wrapper)

	if err != nil {
		// HTTP error
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}
