package template

import (
	"encoding/json"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// ExampleTemplateFullTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateFullTask struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
}

//ExamplePayload - a small struct to store the payload, can be used to hook custom marshaling up as well - optional
//Set to simplistic names (Key, Value) for this example.
type ExamplePayload struct {
	Key         string
	Value       string
	SecretValue string // this value will be omitted by MarshalJSON as we don't want to send it up
}

// MarshalJSON - custom JSON marshaling for this task, this allows you to remove or change data before exporting - completely optional!
func (el ExamplePayload) MarshalJSON() ([]byte, error) {
	// This is a custom JSON Marshalling method. This can be useful for renaming or omitting struct elements from a payload
	// before they're added to the JSON output file.  This is a contrived example - we're omitting SecretValue and renaming
	// the struct elements which we could also have just named Path and Loglevel in the struct definition.
	return json.Marshal(&struct {
		Path     string
		LogLevel string
	}{
		Path:     el.Key,
		LogLevel: el.Value,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateFullTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/FullTask") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateFullTask) Explain() string {
	return "Explanatory help text displayed for this task" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p ExampleTemplateFullTask) Dependencies() []string {
	return []string{
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p ExampleTemplateFullTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task

	// This is what we will use to pass the output from this task back to the core and report to the UI

	result := tasks.Result{
		// Base case result - Use the None status if the check didn't have anything to judge
		Status:  tasks.None,
		Summary: "There were no config files from which to pull the log level",
	}

	validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return

	if !ok { //!ok means the type assertion failed: no []config.ValidateElement found in upstream payload
		result.Summary = "Task did not meet requirements necessary to run: type assertion failure"
		return result
	} else if len(validations) == 0 { //upstream payload is correct type, but upstream task found no valid config files
		return result
	} else {
		payload := []ExamplePayload{}
		for _, validation := range validations { //Iterate through all my results since I may have more than one result to parse through
			// Now I want to check to ensure I have the agent language type I'm concerned about
			logLevel := validation.ParsedResult.FindKey("log_level")
			//This returns a slice of ValidateBlob so I need to walk through them
			if len(logLevel) == 0 {
				result.Status = tasks.Failure
				result.Summary = "Config file doesn't contain log_level"
			}
			for _, value := range logLevel {
				log.Debug("Path to log_level is ", value.PathAndKey())
				log.Debug("value of log_level is ", value.Value())
				payload = append(payload, ExamplePayload{Key: value.PathAndKey(), Value: value.Value()})

				//This is kind of broken, actually... since it loops but only takes the last result. Good thing it's just an example.
				//At least the payload will have everything!
				switch value.Value() {
				case "finest":
					result.Status = tasks.Success //Setting this task's status to Success, we found what we expected/wanted
					result.Summary = "Log level is finest"
				case "info":
					result.Status = tasks.Warning //Setting a task's status to Warning for things that aren't necessarily bad, but might concern us
					result.Summary = "Log level is info, you may want to consider updating the log level to finest before uploading logs to support"
				default:
					result.Status = tasks.Error // User error if something went wrong, but we're not sure if it was our fault or theirs (this isn't the best example)
					result.Summary = "We were unable to determine the log level"
				}
			}
			result.Payload = payload
		}
	}
	return result
	// Additional functions defined within the task should be added below the standard methods to keep code consistent
}
