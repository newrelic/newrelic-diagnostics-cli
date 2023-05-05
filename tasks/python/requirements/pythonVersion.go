package requirements

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
)

//https://github.com/edmorley/newrelic-python-agent/blame/master/newrelic/setup.py#L100

type UnsupportedVersions struct {
	requiredAgentVersion   string
	minimumRequiredVersion string
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
	var unsupportedPythonVersions []string
	var supportedPythonVersions []string
	unsupportedAgentVersionsMap := make(map[string]UnsupportedVersions)

	for _, version := range pyVersions {
		sanitizedPyVersion := removePyVersionPatch(version)
		requiredAgentVersion, isPythonVersionSupported := compatibilityVars.PythonVersionAgentSupportability[sanitizedPyVersion]
		if !isPythonVersionSupported {
			unsupportedPythonVersions = append(unsupportedPythonVersions, version)
		} else {
			isAgentVersionCompatible, err := tasks.VersionIsCompatible(agentVersion, []string{requiredAgentVersion})
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
				result := matchExpression.FindStringSubmatch(requiredAgentVersion)
				minimumRequiredVersion := result[0]

				unsupportedAgentVersionsMap[version] = UnsupportedVersions{requiredAgentVersion: requiredAgentVersion, minimumRequiredVersion: minimumRequiredVersion}

			} else {
				supportedPythonVersions = append(supportedPythonVersions, version)
			}
		}
	}
	if len(supportedPythonVersions) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("None of your versions of Python (%s) are supported by the Python Agent. Please review our documentation on version requirements", strings.Join(pyVersions, ",")),
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}
	var warningStr string
	if len(unsupportedPythonVersions) > 0 {
		warningStr += fmt.Sprintf("Some of your versions of Python (%s) are not supported by the Python Agent. Please review our documentation on version requirements.\n", strings.Join(unsupportedPythonVersions, ","))
	}
	if len(unsupportedAgentVersionsMap) > 0 {
		for v, agentVersionMap := range unsupportedAgentVersionsMap {
			warningStr += fmt.Sprintf("Your %s Python version is not supported by this specific Python Agent Version (%s). You'll have to use a different version of the Python Agent, %s as the minimum, to ensure the agent works as expected.\n", v, agentVersion, agentVersionMap.minimumRequiredVersion)
		}
		if len(supportedPythonVersions) > 0 {
			warningStr += fmt.Sprintf("Your %s Python version(s) are supported by our Python Agent", strings.Join(supportedPythonVersions, ","))
		}
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: warningStr,
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("Your %s Python version(s) are supported by the Python Agent.", strings.Join(supportedPythonVersions, ",")),
	}

}

func removePyVersionPatch(pyVersion string) string {

	versionDigits := strings.Split(pyVersion, ".")
	if len(versionDigits) > 2 {
		return versionDigits[0] + "." + versionDigits[1]
	}
	return pyVersion
}
