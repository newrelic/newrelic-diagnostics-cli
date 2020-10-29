package env

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var expectedAzureEnvVarKey = "WEBSITE_SITE_NAME"

// BaseEnvDetectAzure - This struct defined the sample plugin which can be used as a starting point
type BaseEnvDetectAzure struct{}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvDetectAzure) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/DetectAzure")
}

// Explain - Returns the help text for each individual task
func (p BaseEnvDetectAzure) Explain() string {
	return "Detect if running in Azure environment"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseEnvDetectAzure) Dependencies() []string {
	return []string{"Base/Env/CollectEnvVars"}
}

// Execute - The core work within each task
func (p BaseEnvDetectAzure) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Base/Env/CollectEnvVars"].Status == tasks.Warning {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Unable to gather environment variables, this task did not run",
		}
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)

	if !ok {
		log.Debug("Could not get envVars from upstream")
	}

	_, azureEnvVarIsPresent := envVars[expectedAzureEnvVarKey]

	if azureEnvVarIsPresent {
		return tasks.Result{
			Status:  tasks.Info,
			Summary: "Identified this as an Azure environment.",
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: "Detected that this is not an Azure environment.",
	}
}
