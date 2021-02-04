package agent

import (
	"fmt"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// PHPAgentVersion - Check the version of the PHP Agent according to the logs
type PHPAgentVersion struct {
	returnLastMatchInFile func(search string, filepath string) ([]string, error)
}

//PHPAgentVersionPayload - a small struct to store the payload
type PHPAgentVersionPayload struct {
	Version  string
	Major    int
	Minor    int
	Patch    int
	Build    int
	Codename string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p PHPAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("PHP/Agent/Version")
}

// Explain - Returns the help text for this task
func (p PHPAgentVersion) Explain() string {
	return "Determine New Relic PHP agent version"
}

// Our task won't run if we don't detect the Agent
func (p PHPAgentVersion) Dependencies() []string {
	// Dependent on the PHP agent config to find the logfile location

	return []string{
		"PHP/Config/Agent",
	}
}

// Execute - The core work within each task
func (p PHPAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	/* This is a type assertion to cast my upstream results back into data I know
	  the structure of and can now work with. In this case, I'm casting it back
	to the []validateElements{} I know it should return */

	validations, ok := upstream["PHP/Config/Agent"].Payload.([]config.ValidateElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	if len(validations) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "There were no logs found to check for the agent version.",
		}
	}
	/* a config.ValidateElement will be created for each .ini file found
	we'll loop over them and examine each one  */
	agentVersions := []string{}
	for _, validation := range validations {
		// Now let's check to ensure we have PHP
		logFile := validation.ParsedResult.FindKey("newrelic.logfile")
		//This returns a slice of ValidateBlob(s) so I need to walk through them
		if len(logFile) == 0 {
			continue // didn't find a log file in this one - move on
		}
		// If you made it this far, :youhavephp (congrats)
		for _, value := range logFile {
			// get the filename from this ValidateBlob - and strip " and ' chars
			fileLocation := tasks.TrimQuotes(value.Value())
			regex := `info: New Relic (?P<version>(?P<major>\d+)(\.)?(?P<minor>\d+)(\.)?(?P<patch>\d+)(\.)?(?P<build>\d+)) \("(?P<codename>[^"]+)"`
			//dependency injection here
			regexMatches, err := p.returnLastMatchInFile(regex, fileLocation)
			if len(regexMatches) == 0 || err != nil {
				continue
			}
			// populate our map with version and the matches map
			agentVersions = append(agentVersions, regexMatches[1])
		}
	}
	if len(agentVersions) == 1 {
		parsedVer, err := tasks.ParseVersion(agentVersions[0])
		if err != nil {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: err.Error(),
			}
		}
		return tasks.Result{
			Status:  tasks.Info,
			Summary: parsedVer.String(),
			Payload: parsedVer,
		}
	} else if len(agentVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Unable to determine PHP Agent version from log file",
		}
	}
	// Other result - multiple agent versions found
	return tasks.Result{
		Status:  tasks.Warning,
		Summary: fmt.Sprintf("Expected 1, but found %d versions of the PHP Agent", len(agentVersions)),
	}
}
