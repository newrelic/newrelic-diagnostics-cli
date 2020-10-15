package env

import (
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RubyEnvVersion - This struct defines the Ruby version
type RubyEnvVersion struct {
	cmdExecutor tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p RubyEnvVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Env/Version")
}

// Explain - Returns the help text for each individual task
func (p RubyEnvVersion) Explain() string {
	return "Determine Ruby version"
}

// Dependencies - Returns the dependencies for each task.
func (p RubyEnvVersion) Dependencies() []string {
	return []string{
		"Ruby/Config/Agent",
	}
}

// Execute - The core work within each task
func (p RubyEnvVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result //pass the result back to core and report to UI

	if upstream["Ruby/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Ruby Agent not installed. This task didn't run."
		return result
	}
	result = p.checkRubyVersion()
	return result
}

func (p RubyEnvVersion) checkRubyVersion() (result tasks.Result) {
	version, cmdBuildErr := p.cmdExecutor("ruby", "-v")

	if cmdBuildErr != nil {
		result.Status = tasks.Error
		result.Summary = "Unable to execute command: $ ruby -v. Error: " + cmdBuildErr.Error()
		result.URL = "https://docs.newrelic.com/docs/agents/ruby-agent/getting-started/ruby-agent-requirements-supported-frameworks#ruby_versions"
		return
	}
	versionString := string(version)
	versionString = strings.TrimLeft(versionString, "Ruby ")
	versionString = strings.TrimSpace(versionString)
	result.Summary = versionString //where Info line prints from
	result.Status = tasks.Info
	result.Payload = versionString

	log.Debug("Ruby version found. Version is: " + versionString)

	return
}
