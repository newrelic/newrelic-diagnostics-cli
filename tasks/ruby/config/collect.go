package config

import (
	"os"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RubyConfigCollect - This struct defined the sample plugin which can be used as a starting point
type RubyConfigCollect struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t RubyConfigCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Config/Collect") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (t RubyConfigCollect) Explain() string {
	return "Collect Ruby application gemfiles" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (t RubyConfigCollect) Dependencies() []string {
	return []string{
		"Ruby/Config/Agent", //This identifies this task as dependent on "Ruby/Config/Agent" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results. This helps us know that Ruby is even present.
	}
}

// Execute - The core work within each task
func (t RubyConfigCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	if upstream["Ruby/Config/Agent"].Status == tasks.Success {
		gemfiles, err := findGemfiles()
		if err != nil {
			result.Status = tasks.Error
			result.Summary = "Unable to read working directory. Please check permission and try again."
		}
		if len(gemfiles) == 0 {
			result.Status = tasks.Warning
			result.Summary = "No gemfiles found."
			result.URL = "https://docs.newrelic.com/docs/agents/ruby-agent/installation/install-new-relic-ruby-agent"
		} else {
			result.FilesToCopy = tasks.StringsToFileCopyEnvelopes(gemfiles)
			result.Status = tasks.Success
			result.Payload = gemfiles
			result.Summary = "Successfully added."
		}

	}

	return result
	// Additional functions defined within the task should be added below the standard methods to keep code consistent
}

func findGemfiles() ([]string, error) {
	//return a slice of files (like an array, but any number of elements. Arrays are a defined length)
	gemfiles := []string{"Gemfile$", "Gemfile.lock"}
	localPath, err := os.Getwd()
	if err != nil {
		log.Debug("Error reading local working directory")
		return []string{""}, err
	}

	filepaths := tasks.FindFiles(gemfiles, []string{localPath})
	log.Debug(filepaths)
	return filepaths, nil
}
