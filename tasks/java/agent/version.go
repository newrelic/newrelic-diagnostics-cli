package agent

import (
	"errors"
	"regexp"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// JavaAgentVersion - This struct defined the sample plugin which can be used as a starting point
type JavaAgentVersion struct {
	wdGetter     workingDirectoryGetterFunc
	cmdExec      tasks.CmdExecFunc
	findTheFiles func([]string, []string) []string
}

type workingDirectoryGetterFunc func() (string, error)

// Identifier - This returns the Category, Subcategory and Name of each task
func (p JavaAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (p JavaAgentVersion) Explain() string {
	return "Determine New Relic Java agent version"
}

// Dependencies - Returns the dependencies for ech task.
func (p JavaAgentVersion) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
	}
}

// Execute - The core work within each task
func (p JavaAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Java/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Java agent not detected. Task did not run.",
		}
	}

	paths := p.findJavaJar()
	jarNames := []string{
		"newrelic.jar",
	}

	jars := p.findTheFiles(jarNames, paths)

	// For all jars found, return first successfully parsed version
	// otherwise return last error
	var result tasks.Result
	for _, jarPath := range jars {
		version, err := p.getAgentVersion(jarPath)

		if err != nil {
			result = tasks.Result{
				Summary: "Error parsing the Java Agent version: " + err.Error(),
				Status:  tasks.Error,
			}
			continue
		}

		return tasks.Result{
			Summary: "Java Agent version " + version + " found on path : " + jarPath,
			Status:  tasks.Info,
			Payload: version,
		}
	}

	return result
}

func (p JavaAgentVersion) findJavaJar() []string {

	path, err := p.wdGetter()
	if err != nil {
		log.Debug("Error getting current working directory")
		return nil
	}
	paths := []string{
		path,
		path + "/newrelic",
	}
	return paths
}

func (p JavaAgentVersion) getAgentVersion(jarLocation string) (string, error) {

	version, cmdBuildErr := p.cmdExec("java", "-jar", jarLocation, "-v")

	if cmdBuildErr != nil {
		log.Debug("Error running java agent version -", cmdBuildErr)
		log.Debug("Error was ", string(version))
		return "", cmdBuildErr
	}
	r, _ := regexp.Compile(`(\d+\.\d+\.\d+)(\s+)?$`)
	match := r.FindStringSubmatch(string(version))
	if len(match) > 1 {
		return match[1], nil
	}

	return "", errors.New("unable to determine Java Agent version from output: '" + string(version) + "'")
}
