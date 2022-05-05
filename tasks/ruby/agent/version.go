package agent

import (
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

const agentGemName = "newrelic_rpm"

// RubyAgentVersion - This struct defined the sample plugin which can be used as a starting point
type RubyAgentVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t RubyAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (t RubyAgentVersion) Explain() string {
	return "Determine New Relic Ruby agent version"
}

// Dependencies - Returns the dependencies for ech task.
func (t RubyAgentVersion) Dependencies() []string {
	return []string{
		"Ruby/Config/Agent",
	}
}

// Execute - The core work within each task
func (t RubyAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	//Was the agent config found?
	if upstream["Ruby/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Summary: "Task did not meet requirements necessary to run: Ruby agent is not installed",
			Status:  tasks.None,
		}
	}

	cmdBuild := exec.Command("gem", "list")
	cmdOutput, cmdBuildErr := cmdBuild.CombinedOutput()

	//Error executing `gem list` ?
	if cmdBuildErr != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Unable to execute command: $ gem list. Error: " + cmdBuildErr.Error(),
		}
	}

	//Gem list output -> array of lines
	gemList := strings.Split(string(cmdOutput), "\n")

	for _, gem := range gemList {

		//skip gem if not "newrelic_rpm" is not found in line
		if !regexp.MustCompile(agentGemName).MatchString(gem) {
			continue
		}

		versions, err := parseGemVersions(gem)

		if err != nil {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: fmt.Sprintf(`Error parsing '%s' gem version from: %s`, agentGemName, gem),
			}
		}

		return tasks.Result{
			Status:  tasks.Info,
			Summary: tasks.VersionsJoin(versions, ", "),
			Payload: versions,
		}

	}

	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "Ruby Agent detected on system, but failed to find newrelic_rpm Ruby gem",
		URL:     "https://docs.newrelic.com/docs/agents/ruby-agent/installation/install-new-relic-ruby-agent",
	}
}

// parseAgentVersion takes raw line from gem list and outputs
// parsed task.Ver(s) in a slice
//
// In: newrelic_rpm (6.5.0.357, 5.4.0.347)
// Out: [tasks.Ver{6.5.0.357}, tasks.Ver{5.4.0.347}]
func parseGemVersions(gemListLine string) ([]tasks.Ver, error) {
	//parse version value with parentheses e.g. newrelic_rpm (3.12.0.288)
	versionString := regexp.MustCompile(`\((.*?)\)`).FindStringSubmatch(gemListLine)

	//Was a "(val)" substring not found?
	if len(versionString) < 2 {
		return nil, errors.New("unable to parse version from string")
	}

	rawVersions := strings.Split(versionString[1], ",")
	parsedVersions := []tasks.Ver{}

	for _, v := range rawVersions {

		parsedVersion, err := tasks.ParseVersion(strings.TrimSpace(v))
		if err != nil {
			return nil, err
		}

		parsedVersions = append(parsedVersions, parsedVersion)
	}

	return parsedVersions, nil
}
