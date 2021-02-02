package env

import (
	"regexp"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

//func CmdExecutor from tasks/taskHelpers.go is of type CmdExecFunc(name string, arg ...string) ([]byte, error)
type NodeEnvVersion struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/Version")
}

// Explain - Returns the help text for each individual task
func (p NodeEnvVersion) Explain() string {
	return "Determine Nodejs version"
}

// Dependencies - Returns the dependencies for each task.
func (p NodeEnvVersion) Dependencies() []string {
	return []string{"Node/Config/Agent"}
}

// Execute - The core work within each task
func (p NodeEnvVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Node/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Node agent config file not detected. This task did not run",
		}
	}

	version, cmdBuildErr := p.cmdExec("node", "-v")
	if cmdBuildErr != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Unable to execute command:$ node -v. Error: " + cmdBuildErr.Error(),
		}
	}
	//Convert to string because output v10.7.0 is type []byte
	versionToString := string(version)

	// Remove letters from string to make usage taskshelperversion
	reg := regexp.MustCompile("[^0-9.]+")
	processedString := reg.ReplaceAllString(versionToString, "")

	// If we stripped out all letters and nothing was left
	if len(processedString) < 1 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Unexpected output from received node -v: " + versionToString,
		}

	}

	NodeVersion, err := tasks.ParseVersion(processedString)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "An issue occur while parsing node version: " + err.Error(),
		}
	}

	return tasks.Result{
		Status:  tasks.Info,
		Summary: versionToString,
		Payload: NodeVersion,
	}
}
