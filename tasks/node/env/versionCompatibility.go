package env

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/compatibilityVars"
)

// NodeEnvVersionCompatibility - This struct defines Node.js version compatibility
type NodeEnvVersionCompatibility struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvVersionCompatibility) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/VersionCompatibility")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvVersionCompatibility) Explain() string {
	return "Check Nodejs version compatibility with New Relic Nodejs agent"
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvVersionCompatibility) Dependencies() []string {
	return []string{
		"Node/Env/Version",
		"Node/Agent/Version",
	}
}

// Execute - The core work within each task
func (p NodeEnvVersionCompatibility) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Node/Env/Version"].Status != tasks.Info {
		return tasks.Result{
			Summary: "Task did not meet requirements necessary to run: Node is not installed",
			Status:  tasks.None,
		}
	}

	if upstream["Node/Agent/Version"].Status != tasks.Info {
		return tasks.Result{
			Summary: "Node Agent Version not detected. This task didn't run.",
			Status:  tasks.None,
		}
	}

	nodeVersion, ok := upstream["Node/Env/Version"].Payload.(tasks.Ver)
	agentVersion, valid := upstream["Node/Agent/Version"].Payload.(string)
	if !ok || !valid {
		return tasks.Result{
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
			Status:  tasks.None,
		}
	}
	sanitizedNodeVersion := sanitizeNodeVersion(nodeVersion.String())

	requiredAgentVersions, isNodeVersionSupported := compatibilityVars.NodeSupportedVersions[sanitizedNodeVersion]

	if !isNodeVersionSupported {
		if isOddVersion(sanitizedNodeVersion) {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: fmt.Sprintf("Your %s Node Version is not officially supported by the Node Agent because odd versions are considered unstable/experimental releases", sanitizedNodeVersion),
			}
		}
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Your %s Node version is not in the list of supported versions by the Node Agent. Please review our documentation on version requirements", nodeVersion),
			URL:     "https://docs.newrelic.com/docs/agents/python-agent/getting-started/compatibility-requirements-python-agent#basic",
		}
	}

	//we are dealing with a supported node version, but is it compatible with the agent version they are using?
	isAgentVersionCompatible, err := tasks.VersionIsCompatible(agentVersion, requiredAgentVersions)

	if err != nil {
		return tasks.Result{
			Summary: "There was an issue when checking for Node.js Version compatibility: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	if !isAgentVersionCompatible {
		return tasks.Result{
			Summary: fmt.Sprintf("Your current Node.js version, %v, is not compatible with New Relic's Node.js agent", nodeVersion),
			Status:  tasks.Failure,
			URL:     "https://docs.newrelic.com/docs/agents/nodejs-agent/getting-started/compatibility-requirements-nodejs-agent",
		}
	}

	return tasks.Result{
		Summary: fmt.Sprintf("Your current Node.js version, %v, is compatible with New Relic's Node.js agent", nodeVersion),
		Status:  tasks.Success,
	}

}

func sanitizeNodeVersion(nodeVersion string) string {
	versionDigits := strings.Split(nodeVersion, ".")
	return versionDigits[0]
}

func isOddVersion(nodeVersion string) bool {
	i, _ := strconv.Atoi(nodeVersion)
	return i%2 != 0
}
