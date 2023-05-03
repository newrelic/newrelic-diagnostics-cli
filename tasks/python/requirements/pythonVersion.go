package requirements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
)

//https://github.com/edmorley/newrelic-python-agent/blame/master/newrelic/setup.py#L100

// PythonRequirementsPythonVersion  - This struct defines the Python Version requirement
type PythonRequirementsPythonVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t PythonRequirementsPythonVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Requirements/PythonVersion")
}

// Explain - Returns the help text for this task
func (t PythonRequirementsPythonVersion) Explain() string {
	return "Check Python version compatibility with New Relic Python agent"
}

// Dependencies - Returns the dependencies for the Python/Requirements/PythonVersion task.
func (t PythonRequirementsPythonVersion) Dependencies() []string {
	return []string{
		"Python/Env/Version",
		"Python/Agent/Version",
	}
}

// Execute - The core work within this task
func (t PythonRequirementsPythonVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Python/Env/Version"].Status == tasks.Error {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Python version not detected. This task didn't run.",
		}
	}

	if upstream["Python/Agent/Version"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Python Agent version not detected. This task didn't run.",
		}
	}

	pyVersions := upstream["Python/Env/Version"].Payload.([]string)
	agentVersion := upstream["Python/Agent/Version"].Payload.(string)
	var requiredAgentVersions []string
	var unsupportedAgentVersions []string
	for _, version := range pyVersions {
		sanitizedPyVersion := removePyVersionPatch(version)
		requiredAgentVersion, isPythonVersionSupported := compatibilityVars.PythonVersionAgentSupportability[sanitizedPyVersion]
		if !isPythonVersionSupported {
			unsupportedAgentVersions = append(unsupportedAgentVersions, requiredAgentVersion...)
		} else {
			requiredAgentVersions = append(requiredAgentVersions, requiredAgentVersion...)
		}
	}

	if len(requiredAgentVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("None of your versions of Python (%s) are in the list of supported versions by the Python Agent. Please review our documentation on version requirements", strings.Join(pyVersions, ",")),
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}

	isAgentVersionCompatible, err := tasks.VersionIsCompatible(agentVersion, requiredAgentVersions)

	if err != nil {
		var errMsg string = err.Error()
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("We ran into an error while parsing your current agent version %s. %s", agentVersion, errMsg),
		}
	}

	if !isAgentVersionCompatible {
		//requiredAgentVersions is a single string that contains a range of versions. Let's just get one end of the range
		matchExpression := regexp.MustCompile(`([0-9.]+)`)
		result := matchExpression.FindStringSubmatch(requiredAgentVersions[0])
		minimumRequiredVersion := result[0]

		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Your %s Python version is not supported by this specific Python Agent Version. You'll have to use a different version of the Python Agent, %s as the minimum, to ensure the agent works as expected.", strings.Join(requiredAgentVersions, ","), minimumRequiredVersion),
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}

	if len(unsupportedAgentVersions) > 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("Some of your versions of Python (%s) are supported by the Python Agent while other versions (%s) aren't. Please review our documentation on version requirements", strings.Join(requiredAgentVersions, ","), strings.Join(unsupportedAgentVersions, ",")),
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Your Python version is supported by the Python Agent.",
	}

}

func removePyVersionPatch(pyVersion string) string {

	versionDigits := strings.Split(pyVersion, ".")
	if len(versionDigits) > 2 {
		return versionDigits[0] + "." + versionDigits[1]
	}
	return pyVersion
}
