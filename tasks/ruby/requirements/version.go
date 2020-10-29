package requirements

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"strconv"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

//https://github.com/edmorley/newrelic-python-agent/blame/master/newrelic/setup.py#L100
var rubyVersionAgentSupportability = map[string][]string{
	//the keys are the ruby version and the values are the agent versions that support that specific version
	"2.7":   []string{"6.9.0.363+"},
	"2.6":   []string{"5.7.0.350+"},
	"2.5":   []string{"4.8.0.341+"},
	"2.4":   []string{"3.18.0.329+"},
	"2.3":   []string{"3.9.9.275+"},
	"2.2":   []string{"3.9.9.275+"},
	"2.1":   []string{"3.9.9.275+"},
	"2.0":   []string{"3.9.6.257+"},
	"1.9.3": []string{"3.9.6.257-3.18.1.330"},
	"1.9.2": []string{"3.9.6.257-3.18.1.330"},
	"1.8.7": []string{"3.9.6.257-3.18.1.330"},
}

// RubyRequirementsVersion  - This struct defines the Ruby Version requirement
type RubyRequirementsVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t RubyRequirementsVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Ruby/Requirements/Version")
}

// Explain - Returns the help text for this task
func (t RubyRequirementsVersion) Explain() string {
	return "Check Ruby version compatibility with New Relic Ruby agent"
}

// Dependencies - Returns the dependencies for the Python/Requirements/PythonVersion task.
func (t RubyRequirementsVersion) Dependencies() []string {
	return []string{
		"Ruby/Env/Version",
		"Ruby/Agent/Version",
	}
}

// Execute - The core work within this task
func (t RubyRequirementsVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Ruby/Env/Version"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Ruby version not detected. This task didn't run.",
		}
	}

	if upstream["Ruby/Agent/Version"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Ruby Agent version not detected. This task didn't run.",
		}
	}

	rubyVersion := upstream["Ruby/Env/Version"].Payload.(string) //ruby 2.4.0p0 (2016-12-24 revision 57164) [x86_64-linux] --> regex for ruby\s([^a-z]+)
	agentVersions := upstream["Ruby/Agent/Version"].Payload.([]tasks.Ver)

	sanitizedRubyVersion, err := sanitizeRubyVersionPayload(rubyVersion)

	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "While parsing the Ruby Version, we encountered an error: " + err.Error(),
		}
	}

	requiredAgentVersions, isRubyVersionSupported := rubyVersionAgentSupportability[sanitizedRubyVersion]

	if !isRubyVersionSupported {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Your %s Ruby version is not in the list of supported versions by the Ruby Agent. Please review our documentation on version requirements", rubyVersion),
			URL:     "https://docs.newrelic.com/docs/agents/ruby-agent/getting-started/ruby-agent-requirements-supported-frameworks",
		}
	}
	var compatibleVersions []string
	var incompatibleVersions []string
	var warningMessage string
	var successMessage string
	for _, version := range agentVersions {
		isAgentVersionCompatible, err := version.CheckCompatibility(requiredAgentVersions)
		if err != nil {
			var errMsg string = err.Error()
			log.Debug(errMsg)
			return tasks.Result{
				Status:  tasks.Error,
				Summary: fmt.Sprintf("We ran into an error while parsing your current agent version %s. %s", agentVersions, errMsg),
			}
		}
		if isAgentVersionCompatible {
			compatibleVersions = append(compatibleVersions, version.String())
			successMessage += "Compatible Version detected: " + version.String() + "\n"
		} else {
			incompatibleVersions = append(incompatibleVersions, version.String())
			warningMessage += "Incompatible Version detected: " + version.String() + "\n"

		}

	}
	if len(incompatibleVersions) > 0 && len(compatibleVersions) > 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: warningMessage + successMessage,
		}
	}
	if len(incompatibleVersions) > 0 && len(compatibleVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: warningMessage,
		}
	}
	return tasks.Result{
		Status:  tasks.Success,
		Summary: successMessage,
	}

}

func sanitizeRubyVersionPayload(rubyVersionPayload string) (string, error) {
	matchExpression := regexp.MustCompile(`ruby\s([^a-z]+)`)
	result := matchExpression.FindStringSubmatch(rubyVersionPayload)
	if len(result) == 0 {
		return "", errors.New("No found result for Ruby Version when parsing for payload " + rubyVersionPayload)
	}
	rubyVersion := result[1]
	versionDigits := strings.Split(rubyVersion, ".")

	majorInt, err := strconv.Atoi(versionDigits[0])

	if err != nil {
		return "", errors.New("No found result for Ruby Version when parsing for payload " + rubyVersionPayload)
	}

	if majorInt >= 2 {
		return versionDigits[0] + "." + versionDigits[1], nil
	}
	if majorInt == 1 {
		return versionDigits[0] + "." + versionDigits[1] + "." + versionDigits[2], nil
	}
	return rubyVersion, nil
}
