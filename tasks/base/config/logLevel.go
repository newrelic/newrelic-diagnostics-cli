package config

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseConfigLogLevel - This struct defined the sample plugin which can be used as a starting point
type BaseConfigLogLevel struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseConfigLogLevel) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/LogLevel") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p BaseConfigLogLevel) Explain() string {
	return "Determine New Relic agent logging level" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p BaseConfigLogLevel) Dependencies() []string {
	return []string{
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p BaseConfigLogLevel) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	validations, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if ok {
		log.Debug("correct type")
		//		log.Debug(validations) //This may be useful when debugging to log the entire results to the screen
	}

	if len(validations) == 0 {
		result.Status = tasks.None // Use the None status if the check didn't have anything to judge
		result.Summary = "There were no config files to pull the log level from."
	} else {

		for _, validation := range validations { //Iterate through all my results since I may have more than one result to parse through
			// Now I want to check to ensure I have the agent language type I'm concerned about
			logLevels := validation.ParsedResult.FindKey("log_level")
			//This returns a map of strings so I need to walk through them (or call the one I'm looking for by name)
			if len(logLevels) == 0 {
				result.Status = tasks.Failure
				result.Summary = "Config file doesn't contain log_level"
				result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/configuration/configure-agent"
			}
			for _, logLevel := range logLevels {
				log.Debug("Path to log_level is ", logLevel.Key)
				log.Debug("value of log_level is ", logLevel.Value())
				switch logLevel.Value() {
				case "finest":
					result.Status = tasks.Success //Setting this task's status to Success, we found what we expected/wanted
					result.Summary = "Log level is finest"
				case "info":
					result.Status = tasks.Warning //Setting a task's status to Warning for things that aren't necessarily bad, but might concern us
					result.Summary = "Log level is info, you may want to consider updating the log level to finest before uploading logs to support"
					result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/configuration/configure-agent"
				default:
					result.Status = tasks.Error // User error if something went wrong, but we're not sure if it was our fault or theirs (this isn't the best example)
					result.Summary = "We couldn't figure out the log level."
				}
			}
		}
	}
	return result
	// Additional functions defined within the task should be added below the standard methods to keep code consistent
}
