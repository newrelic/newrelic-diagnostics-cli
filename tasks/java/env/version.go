package env

import (
	"os/exec"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// JavaEnvVersion - This struct defined the sample plugin which can be used as a starting point
type JavaEnvVersion struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p JavaEnvVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Env/Version")
}

// Explain - Returns the help text for this task
func (p JavaEnvVersion) Explain() string {
	return "Determine JRE/JVM version"
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p JavaEnvVersion) Dependencies() []string {
	return []string{
		"Java/Config/Agent", //This identifies this task as dependent on "Java/Config/Agent" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
// Check the upstream status; Err will issue a warning that Java was not found in the path; Otherwise, we found the version -add summary and status to the result
func (p JavaEnvVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	if upstream["Java/Config/Agent"].Status == tasks.Success {
		version, err := getJREVersion()
		if err != nil {
			result.Summary = "Java not found in PATH"
			result.Status = tasks.Warning //Do we want a URL in this one?
		} else {
			result.Summary = version
			result.Status = tasks.Info
			result.Payload = version
		}
		log.Debug(version)
	}

	return result

}

//Execute command to the JRE. return the output as a string; if we throw an error return the error
func getJREVersion() (string, error) {

	cmdBuild := exec.Command("java", "-version")

	version, cmdBuildErr := cmdBuild.CombinedOutput()
	if cmdBuildErr != nil {
		log.Debug("Error running java -version", cmdBuildErr)
		log.Debug("Error was ", string(version))
		return cmdBuildErr.Error(), cmdBuildErr
	}
	return string(version), nil
}
