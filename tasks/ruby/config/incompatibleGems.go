package config

import (
	"fmt"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// List of incompatible gems: keys are displayed to customer in task Result
var incompatibleGems = map[string]string{
	"db-charmer":            "^[^# ]*(gem 'db-charmer').*",
	"escape_utils":          "^[^# ]*(gem 'escape_utils').*",
	"right_http_connection": "^[^# ]*(gem 'right_http_connection').*",
	"ar-octopus":            "^[^# ]*(gem 'ar-octopus').*",
}

// RubyConfigIncompatibleGems - This task handles incompatible gems that are across the board incompatible regardless of Ruby Agent version and gem version
type RubyConfigIncompatibleGems struct {
}

type BadGemAndPath struct {
	GemName     string
	GemfilePath string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t RubyConfigIncompatibleGems) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Config/IncompatibleGems")
}

// Explain - Returns the help text for each individual task
func (t RubyConfigIncompatibleGems) Explain() string {
	return "Check gem compatibility with New Relic Ruby agent"
}

// Dependencies - Returns the dependencies for ech task.
func (t RubyConfigIncompatibleGems) Dependencies() []string {
	return []string{
		"Ruby/Config/Collect",
	}
}

// Execute - The core work within each task
func (t RubyConfigIncompatibleGems) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	if upstream["Ruby/Config/Collect"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Either no Gemfile or newrelic.yml was found"
		return result
	}

	gemfiles, ok := upstream["Ruby/Config/Collect"].Payload.([]string)
	if !ok {
		result.Status = tasks.Error
		result.Summary = "Error getting Gemfile list; expecting type of Slice of Strings."
		return result
	}
	incompatibleGems := checkGems(gemfiles)
	if len(incompatibleGems) > 0 {
		result.Status = tasks.Failure
		result.URL = "https://docs.newrelic.com/docs/agents/ruby-agent/troubleshooting/incompatible-gems"
		result.Payload = incompatibleGems
	}
	if len(gemfiles) > 0 && len(incompatibleGems) == 0 {
		result.Status = tasks.Success
	}
	result.Summary = displayResults(incompatibleGems)
	return result
}

func checkGems(gemfiles []string) (incompatibleGem []BadGemAndPath) {
	for _, gemfile := range gemfiles {
		for name, gemRegex := range incompatibleGems {
			if tasks.FindStringInFile(gemRegex, gemfile) {
				log.Debug("Incompatible gem", name, "was found")
				incompatibleGem = append(incompatibleGem, BadGemAndPath{GemName: name, GemfilePath: gemfile})
			}
		}
	}
	return
}

func displayResults(incompatibleGems []BadGemAndPath) (summary string) {
	
	if len(incompatibleGems) == 0 {
		return "There were no incompatible gems found."
	}
	summary = fmt.Sprintf("We detected %d Ruby gem(s) incompatible with the New Relic Ruby agent:", len(incompatibleGems))
	for _, gem := range incompatibleGems {
		summary += fmt.Sprintf("\n%s - %s", gem.GemName, gem.GemfilePath)
	}
	return summary
}
