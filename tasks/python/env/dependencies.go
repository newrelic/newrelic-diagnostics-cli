package env

import (
	"encoding/json"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/repository"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// PythonEnvDependencies - This struct defines the project dependencies.
type PythonEnvDependencies struct {
	iPipEnvVersion repository.IPipEnvVersion
}

// PythonEnvDependenciesPayload - This is the payload.
type PythonEnvDependenciesPayload struct {
	Payload string
}

// MarshalJSON - custom JSON marshaling for this task, in this case we ignore everything.
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
	result := t.getProjectDependencies()
	return result
}

func (t PythonEnvDependencies) getProjectDependencies() tasks.Result {
	var errorsToReturn []string
	var summariesToReturn []string
	var payloadToReturn []string
	var fileToCopyToReturn []tasks.FileCopyEnvelope
	result_1 := t.iPipEnvVersion.CheckPipVersion("pip")
	if result_1.Status == tasks.Error {
		errorsToReturn = append(errorsToReturn, result_1.Summary)
	} else {
		summariesToReturn = append(summariesToReturn, result_1.Summary)
		if slice, ok := result_1.Payload.([]string); ok {
			payloadToReturn = append(payloadToReturn, slice...)
		}
		fileToCopyToReturn = append(fileToCopyToReturn, result_1.FilesToCopy...)
	}
	result_2 := t.iPipEnvVersion.CheckPipVersion("pip3")
	if result_2.Status == tasks.Error {
		errorsToReturn = append(errorsToReturn, result_2.Summary)
	} else {
		summariesToReturn = append(summariesToReturn, result_2.Summary)
		if slice, ok := result_2.Payload.([]string); ok {
			payloadToReturn = append(payloadToReturn, slice...)
		}
		fileToCopyToReturn = append(fileToCopyToReturn, result_2.FilesToCopy...)
	}
	errorStr := strings.Join(errorsToReturn, "\n")
	successStr := strings.Join(summariesToReturn, "\n")
	payloadWithoutDup := removeDuplicates(payloadToReturn)
	if len(errorsToReturn) == 2 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: errorStr,
		}
	} else if len(errorsToReturn) > 0 {

		return tasks.Result{
			Status:      tasks.Warning,
			Summary:     errorStr + "\n" + successStr,
			Payload:     payloadWithoutDup,
			FilesToCopy: fileToCopyToReturn,
		}
	}

	return tasks.Result{
		Status:      tasks.Success,
		Summary:     successStr,
		Payload:     payloadWithoutDup,
		FilesToCopy: fileToCopyToReturn,
	}
}

func removeDuplicates(slice []string) []string {
	uniqueStrings := make(map[string]bool)
	result := []string{}

	for _, str := range slice {
		if _, exists := uniqueStrings[str]; !exists {
			uniqueStrings[str] = true
			result = append(result, str)
		}
	}

	return result
}
