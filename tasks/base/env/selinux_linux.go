package env

import (
	"regexp"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type SEMode string

const (
	NotInstalled SEMode = "Not Installed"
	NotEnforced         = "Not Enforced"
	Enforced            = "Enforced"
	SEUnknown           = "Could not be determined"
)

// BaseEnvCheckSELinux - This struct defined the sample plugin which can be used as a starting point
type BaseEnvCheckSELinux struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	cmdExec func(name string, arg ...string) ([]byte, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvCheckSELinux) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/SELinux")
}

// Explain - Returns the help text for this task
func (p BaseEnvCheckSELinux) Explain() string {
	return "Check for SELinux presence."
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p BaseEnvCheckSELinux) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p BaseEnvCheckSELinux) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	sestatusOutput, sestatusErr := p.cmdExec("sestatus")
	if sestatusErr != nil {
		//if we get an error such as "Command 'sestatus' not found" or "bash: sestatus: command not found", this is a good sign
		if strings.Contains(sestatusErr.Error(), "not found") {
			return tasks.Result{
				Status:  tasks.None,
				Summary: "SELinux does seem to be installed in this environment: " + sestatusErr.Error(),
				Payload: NotInstalled,
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Unable to execute command: sestatus Error: " + sestatusErr.Error(),
			Payload: SEUnknown,
		}
	}

	outputToString := string(sestatusOutput)
	regex := regexp.MustCompile(`(Current mode:\s+enforcing)`)
	matchResults := regex.FindStringSubmatch(outputToString)

	if len(matchResults) < 1 {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Verified SELinux is not enforcing.",
			Payload: NotEnforced,
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: "We have detected that SELinux is enabled with enforcing mode in your environment. If you are having issues installing a New Relic product or you have no data reporting, temporarily disable SELinux, to verify this resolves the issue.",
		Payload: Enforced,
	}

}
