package minion

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

type SyntheticsMinionDetect struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p SyntheticsMinionDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Synthetics/Minion/Detect")
}

// Explain - Returns the help text for each individual task
func (p SyntheticsMinionDetect) Explain() string {
	return "Detect if running on New Relic Synthetics private minion (legacy)"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p SyntheticsMinionDetect) Dependencies() []string {
	return []string{}
}

// Execute - Checks for existence of the synthetics-minion.jar file in its working directory to determine if
// it is being executed on a private minion.
func (p SyntheticsMinionDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	onMinion := tasks.FileExists("synthetics-minion.jar")

	if !onMinion {
		result.Status = tasks.None
		result.Summary = "synthetics-minion.jar not found. Not running on synthetics private minion."
		log.Debug("synthetics-minion.jar not found in working directory.")
		return result
	}

	result.Status = tasks.Success
	result.Summary = "synthetics-minion.jar found. Running on synthetics private minion."
	log.Debug("synthetics-minion.jar successfully found in working directory.")
	return result
}
