package env

import (
	"bufio"
	"encoding/json"
	"os/exec"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
) 

// PythonEnvDependencies - This struct defines the project dependencies.
type PythonEnvDependencies struct {
}

// PythonEnvDependenciesPayload - This is the payload.
type PythonEnvDependenciesPayload struct {
	Payload string
}

//MarshalJSON - custom JSON marshaling for this task, in this case we ignore everything.
func (el PythonEnvDependenciesPayload) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
	}{})
}

// Identifier - This returns the Category, Subcategory and Name of this task.
func (t PythonEnvDependencies) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Env/Dependencies")
}

// Explain - Returns the help text for the Python/Env/Dependencies task.
func (t PythonEnvDependencies) Explain() string {
	return "Collect Python application packages"
}

// Dependencies - Returns the dependencies for this task.
func (t PythonEnvDependencies) Dependencies() []string {
	return []string{
		"Python/Config/Agent",
	}
}

// Execute - The core work within this task
func (t PythonEnvDependencies) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Python/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Summary: "Python Agent not installed. This task didn't run.",
			Status:  tasks.None,
		}
	}
	result := getProjectDependencies()
	return result
}

func getProjectDependencies() tasks.Result {
	cmdBuild := exec.Command("pip", "freeze")
	pipFreezeOutput, cmdBuildErr := cmdBuild.CombinedOutput()

	if cmdBuildErr != nil {
		return tasks.Result{
			Summary: "Unable to execute command: $ pip freeze. Error: " + cmdBuildErr.Error(),
			Status:  tasks.Error,
		}
	}

	if pipFreezeOutput == nil {
		return tasks.Result{
			Summary: "Collected pip freeze but output was empty",
			Status:  tasks.Error,
		}
	}

	// This will return the output of 'pip freeze' as a string, with packages separated by newlines
	pipFreezeOutputString := string(pipFreezeOutput)

	// stream payload to zip file to allow troubleshooting with the parsed config as needed
	stream := make(chan string)

	go streamBlob(pipFreezeOutputString, stream)

	log.Debug("Result of running 'pip freeze': ")
	log.Debug(pipFreezeOutputString)

	pipFreezeOutputSlice := filterPipFreeze(pipFreezeOutputString)

	return tasks.Result{
		Summary:     "Collected pip freeze. See pipFreeze.txt for more info.", //where Info line prints from
		Status:      tasks.Info,
		Payload:     pipFreezeOutputSlice,
		FilesToCopy: []tasks.FileCopyEnvelope{tasks.FileCopyEnvelope{Path: "pipFreeze.txt", Stream: stream}},
	}
}

func streamBlob(input string, ch chan string) {
	defer close(ch)

	scanner := bufio.NewScanner(strings.NewReader(input))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}
}

// this function will remove any warnings from pipFreezeOutput so only actionable data is left in the data structure.
func filterPipFreeze(pipFreezeOutput string) []string {
	var filteredOutput []string
	keys := make(map[string]bool)

	pipFreezeOutput = strings.ToLower(pipFreezeOutput)

	pipFreezeOutputSlice := strings.Fields(strings.ToLower(pipFreezeOutput))
	for _, v := range pipFreezeOutputSlice {
		if strings.Contains(v, "==") {
			if _, value := keys[v]; !value {
				keys[v] = true
				filteredOutput = append(filteredOutput, v)

			}
		}
	}

	return filteredOutput
}
