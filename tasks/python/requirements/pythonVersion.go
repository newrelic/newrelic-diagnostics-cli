package requirements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newrelic/NrDiag/tasks"
)

//https://github.com/edmorley/newrelic-python-agent/blame/master/newrelic/setup.py#L100
var pythonVersionAgentSupportability = map[string][]string{
	//the keys are the python version and the values are the agent versions that support that specific version
	"3.8": []string{"5.2.3.131+"},
	"3.7": []string{"3.4.0.95+"},
	"3.6": []string{"2.80.0.60+"},
	"3.5": []string{"2.78.0.57+"},
	"3.4": []string{"2.42.0.35-4.20.0.120"},
	"3.3": []string{"2.42.0.35-3.4.0.95"},
	"2.7": []string{"2.42.0.35+"},
	"2.6": []string{"2.42.0.35-3.4.0.95"},
}

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

	if upstream["Python/Env/Version"].Status != tasks.Info {
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

	pyVersion := upstream["Python/Env/Version"].Payload.(string)
	agentVersion := upstream["Python/Agent/Version"].Payload.(string)

	sanitizedPyVersion := removePyVersionPatch(pyVersion)

	requiredAgentVersions, isPythonVersionSupported := pythonVersionAgentSupportability[sanitizedPyVersion]

	if !isPythonVersionSupported {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Your %s Python version is not in the list of supported versions by the Python Agent. Please review our documentation on version requirements", pyVersion),
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
			Summary: fmt.Sprintf("Your %s Python version is not supported by this specific Python Agent Version. You'll have to use a different version of the Python Agent, %s as the minimum, to ensure the agent works as expected.", pyVersion, minimumRequiredVersion),
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
