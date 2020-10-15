package agent

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/newrelic/NrDiag/suites"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// EOLVersions prior to: Node 1.14.1, Java 3.6.0 (except 2.21.7), .NET 5.1, PHP 5.0.0.115, Python 2.42.0, Ruby 3.9.6
//To satisfy these requirements, here we're specifying the last version released prior the versions listed above
var EOLVersions = map[string][]string{
	"Node":   []string{"1.0.0-1.14.0"},
	"Java":   []string{"1.3.0-2.21.4", "3.0.0-3.5.1"},
	"Python": []string{"1.0.2.130-2.40.0.34"},
	"Ruby":   []string{"3.0.0-3.9.5.251"},
	"PHP":    []string{"2.0.2.65-4.23.4.113"},
	"DotNet": []string{"2.0.6-5.0.136.0"},
}

type agentVersion struct {
	name    string
	version string
}

// BaseAgentEOL is the basic struct for our task
type BaseAgentEOL struct {
	suiteManager *suites.SuiteManager
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseAgentEOL) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Agent/EOL")
}

// Explain - Returns the help text for each individual task
func (p BaseAgentEOL) Explain() string {
	return "Detect end of life (EOL) New Relic agents"
}

//Dependencies - As this task is a "Base" task, we wanted it to run on every suite. But we didn't want it to
//kick off tasks for agents unrelated to the selected suite (e.g. You run --suites java but you see results for Node/Agent/Version)
//This is our current solution due to time constraints
func (p BaseAgentEOL) Dependencies() []string {

	defaultDependencies := []string{
		"Node/Agent/Version",
		"Java/Agent/Version",
		"Python/Agent/Version",
		"Ruby/Agent/Version",
		"PHP/Agent/Version",
	}
	if runtime.GOOS == "windows" {
		defaultDependencies = append(defaultDependencies, "DotNet/Agent/Version")
	}

	if len(p.suiteManager.SelectedSuites) < 1 {
		return defaultDependencies
	}

	var suiteDependencies []string
	for _, suite := range p.suiteManager.SelectedSuites {
		switch suite.Identifier {
		case "node":
			suiteDependencies = append(suiteDependencies, "Node/Agent/Version")
		case "java":
			suiteDependencies = append(suiteDependencies, "Java/Agent/Version")
		case "python":
			suiteDependencies = append(suiteDependencies, "Python/Agent/Version")
		case "ruby":
			suiteDependencies = append(suiteDependencies, "Ruby/Agent/Version")
		case "php":
			suiteDependencies = append(suiteDependencies, "PHP/Agent/Version")
		case "dotnet":
			if runtime.GOOS == "windows" {
				suiteDependencies = append(suiteDependencies, "DotNet/Agent/Version")
			}
		case "all":
			return defaultDependencies
		}
	}
	return suiteDependencies
}

// Execute - The core work within each task
func (p BaseAgentEOL) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	//decisionQueue is a slice for unpacking multiple versions from a single task
	//we'll like to eventually standardize on all version tasks returning []Ver payloads
	//currently Ruby/Agent/Version is the only compliant task
	decisionQueue := []agentVersion{}

	//Errors encountered retrieving payload contents
	payloadErrors := []agentVersion{}

	for identifier, val := range upstream {
		identifierComponents := strings.Split(identifier, "/")
		if len(identifierComponents) != 3 {
			log.Debug("Error parsing agent name from ", identifier)
			continue
		}
		agentName := identifierComponents[0]

		if (val.Status == tasks.Success) || (val.Status == tasks.Info) {

			switch v := val.Payload.(type) {
			case string:
				decisionQueue = append(decisionQueue, agentVersion{agentName, v})
			case tasks.Ver: // PHP/Agent/Version payload type
				decisionQueue = append(decisionQueue, agentVersion{agentName, v.String()})
			case []tasks.Ver:
				if len(v) > 0 {
					for _, agentVer := range v {
						decisionQueue = append(decisionQueue, agentVersion{agentName, agentVer.String()})
					}
				} else {
					payloadErrors = append(payloadErrors, agentVersion{agentName, ""})
				}
			default:
				payloadErrors = append(payloadErrors, agentVersion{agentName, ""})
			}
		}
	}

	successes, errors, failures := processDecisionQueue(decisionQueue)
	errors = append(errors, payloadErrors...)
	taskStatus, taskSummary := genSummaryStatus(successes, errors, failures)

	return tasks.Result{
		Status:  taskStatus,
		Summary: taskSummary,
		URL:     "https://discuss.newrelic.com/t/important-upcoming-changes-to-supported-agent-versions/72280",
	}
}

func isItEOL(version string, agentName string) (bool, error) {

	unsupportedVersions := EOLVersions[agentName]
	isItUnsupported, err := tasks.VersionIsCompatible(version, unsupportedVersions)
	if err != nil {
		return false, err
	}
	return isItUnsupported, nil
}

//processDecisionQueue evaluates EOL status about accumulated agentVersions
func processDecisionQueue(agentVersionQueue []agentVersion) (successes, errors, failures []agentVersion) {

	for _, currentAgent := range agentVersionQueue {
		eolResult, eolResultErr := isItEOL(currentAgent.version, currentAgent.name)

		if eolResultErr != nil {
			errors = append(errors, currentAgent)
			continue
		}

		if eolResult {
			failures = append(failures, currentAgent)
			continue
		}

		successes = append(successes, currentAgent)
	}
	return successes, errors, failures
}

func genSummaryStatus(successes, errors, failures []agentVersion) (tasks.Status, string) {
	total := len(successes) + len(errors) + len(failures)

	if total == 0 {
		return tasks.None, "No New Relic agent versions detected. This task did not run"
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("We detected %d New Relic agent(s) running on your system:", total))

	if len(successes) > 0 {
		summary.WriteString(fmt.Sprintf("\n%d New Relic agent(s) whose version is within the scope of support", len(successes)))
	}

	if len(failures) > 0 {
		summary.WriteString(fmt.Sprintf("\n%d New Relic agent(s) whose version has reached EOL:", len(failures)))

		for _, a := range failures {
			summary.WriteString(fmt.Sprintf("\n\t%s agent %s", a.name, a.version))
		}

		return tasks.Failure, summary.String()
	}

	if len(errors) > 0 {
		summary.WriteString(fmt.Sprintf("\n%d New Relic agent(s) whose EOL status could not be determined:", len(errors)))
		for _, a := range errors {
			summary.WriteString(fmt.Sprintf("\n\t%s agent %s", a.name, a.version))
		}
		return tasks.Error, summary.String()
	}

	return tasks.Success, summary.String()

}
