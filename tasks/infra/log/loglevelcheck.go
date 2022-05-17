package log

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraLogLevelCheck - This struct defined the sample plugin which can be used as a starting point
type InfraLogLevelCheck struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t InfraLogLevelCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Log/LevelCheck") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (t InfraLogLevelCheck) Explain() string {
	return "Check if New Relic Infrastructure agent logging level is set to debug" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (t InfraLogLevelCheck) Dependencies() []string {
	return []string{
		"Infra/Config/Agent", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (t InfraLogLevelCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Infrastructure Agent config not present"
		return result
	}
	validations, ok := upstream["Infra/Config/Agent"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if !ok {
		result.Status = tasks.Error
		result.Summary = tasks.AssertionErrorSummary
		return result
	}
	if len(validations) == 0 {
		result.Status = tasks.None // Use the None status if the check didn't have anything to judge
		result.Summary = "There were no config files to pull the log level from."
	}

	validationLevel := logLevelCheck(validations)
	if validationLevel != "1" {
		result.Status = tasks.Warning
		result.Summary = "Infrastructure logging level not set to verbose. If troubleshooting an Infrastructure issue, please set verbose: 1 in newrelic-infra.yml."
		result.URL = "https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/troubleshooting/generate-logs-troubleshooting-infrastructure"
	} else if validationLevel == "1" {
		result.Status = tasks.Success
		result.Summary = "Infrastructure logging level is set to verbose."
	}
	return result
	// Additional functions defined within the task should be added below the standard methods to keep code consistent

}

func logLevelCheck(configs []config.ValidateElement) string {
	for _, config := range configs {
		if config.Config.FileName == "newrelic-infra.yml" {
			result := config.ParsedResult.FindKey("verbose")
			if len(result) == 1 {
				return result[0].Value()
			}
		}
	}
	return ""

}
