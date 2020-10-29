package requirements

import (
	"regexp"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var supportedVersions = map[string][]string{

	// any version supported
	"bottle":   []string{"0+"},
	"cherrypy": []string{"0+"},
	"django":   []string{"0+"},
	"falcon":   []string{"0+"},
	"flask":    []string{"0+"},
	"pylons":   []string{"0+"},
	"pyramid":  []string{"0+"},
	"sanic":    []string{"0+"},
	"web2py":   []string{"0+"},

	// specific versions supported
	"tornado": []string{"6.0+"},
	"aiohttp": []string{"2.2+"},
}

// PythonRequirementsWebframework - This struct defines the Webframework requirement.
type PythonRequirementsWebframework struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task.
func (t PythonRequirementsWebframework) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Requirements/Webframework")
}

// Explain - Returns the help text for the PythonRequirementsWebframework task.
func (t PythonRequirementsWebframework) Explain() string {
	return "Check web framework compatibility with New Relic Python agent"
}

// Dependencies - Returns the dependencies for this task.
func (t PythonRequirementsWebframework) Dependencies() []string {
	return []string{
		"Python/Env/Dependencies",
	}
}

// Execute - The core work within this task.
func (t PythonRequirementsWebframework) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Python/Env/Dependencies"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No dependencies found. This task didn't run.",
		}
	}

	// Get list of dependencies from upstream payload.
	pipFreezeOutput, ok := upstream["Python/Env/Dependencies"].Payload.([]string)

	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	return checkWebframework(pipFreezeOutput)
}

// Function to check if compatible framework is present in pip freeze

func checkWebframework(pipFreezeOutput []string) (result tasks.Result) {
	// Iterate through customer's list of dependencies to identify if compatible framework(s) present
	// Initialize nil slice to store supported frameworks found in pipfreeze
	// ```rubyValidation, ok := upstream["Ruby/Config/Agent"].Payload.([]config.ValidateElement)```
	var compatibleCustomerFrameworks []string
	// seeding the output for validation
	// pipFreezeOutput = append(pipFreezeOutput, "tornado==3.5", "aiohttp==3.5")
	for _, pipFreezeOutputItem := range pipFreezeOutput {

		framework, frameworkVersion := extractFrameworkDetails(pipFreezeOutputItem)
		framework = strings.ToLower(framework)
		versionParam := supportedVersions[framework]

		if versionParam != nil {
			v, err := tasks.VersionIsCompatible(frameworkVersion, versionParam)

			if err != nil {
				log.Debugf("Encountered %s while attempting to parse the framework version.", err)
			}
			if v == true {
				compatibleCustomerFrameworks = append(compatibleCustomerFrameworks, pipFreezeOutputItem)

			}
		}
	}
	if len(compatibleCustomerFrameworks) >= 1 {
		log.Debug("Web framework found is compatible.")
		return tasks.Result{
			Summary: "Found compatible web framework.",
			Status:  tasks.Success,
			Payload: compatibleCustomerFrameworks,
		}

	}
	log.Debug("Web framework found is not compatible.")
	return tasks.Result{
		Status:  tasks.Warning,
		Summary: "No compatible web framework found.",
		URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/instrumented-python-packages#web-frameworks, https://docs.newrelic.com/docs/agents/python-agent/supported-features/python-background-tasks",
	}
}

// Function to split pip freeze line and return major, minor and rev
func extractFrameworkDetails(pipFreezeOutputItem string) (string, string) {
	// Separate package name from version
	splitOutputItem := strings.Split(pipFreezeOutputItem, "==")
	if len(splitOutputItem) > 1 {
		framework, version := splitOutputItem[0], splitOutputItem[1]
		reg, err := regexp.Compile("[^0-9.]")
		if err != nil {
			log.Debug("Regex failed to compile", err)
			return "", ""
		}
		version = reg.ReplaceAllString(version, "")
		return framework, version

	}
	return "", ""
}
