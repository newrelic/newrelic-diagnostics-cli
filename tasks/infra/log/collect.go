package log

import (
	"fmt"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraLogCollect - This struct defined the sample plugin which can be used as a starting point
type InfraLogCollect struct {
	validatePaths func([]string) []tasks.CollectFileStatus
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraLogCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Log/Collect")
}

// Explain - Returns the help text for each individual task
func (p InfraLogCollect) Explain() string {
	return "Collect New Relic Infrastructure agent log files"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraLogCollect) Dependencies() []string {
	return []string{
		"Infra/Config/Agent",
		"Base/Config/Validate",
	}
}

// Execute - Returns result containing the log_file value(s) parsed from any found newrelic-infra.yml files previously collected.
func (p InfraLogCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Not executing task. Infra agent not found.",
		}
	}

	if upstream["Base/Config/Validate"].Status != tasks.Success && upstream["Base/Config/Validate"].Status != tasks.Warning {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Not executing task. Infra config file not found.",
		}
	}

	configElements, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure.",
		}
	}

	logFilePaths := getLogFilePaths(configElements)

	if len(logFilePaths) < 1 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "New Relic Infrastructure configuration file did not specify log file path",
		}
	}

	//check for log paths that provide files that are not accessible for collection
	fileStatuses := p.validatePaths(logFilePaths)

	var invalidFilePaths []string
	var validFilePaths []string
	var resultSummary string

	for _, file := range fileStatuses {
		if !file.IsValid {
			invalidFilePaths = append(invalidFilePaths, file.Path)
			resultSummary += fmt.Sprintf("The log file path found in the New Relic config file (%q) did not provide a file that was accessible to us:\n%q\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting", file.Path, (file.ErrorMsg).Error())
			continue
		}
		validFilePaths = append(validFilePaths, file.Path)
		resultSummary += fmt.Sprintf("Success, logs found! We were able to access the following New Relic log file through the path provided in your New Relic config file:%s\n", file.Path)
	}

	if len(invalidFilePaths) == 0 {
		return tasks.Result{
			Status:      tasks.Success,
			Summary:     resultSummary,
			Payload:     logFilePaths,
			FilesToCopy: tasks.StringsToFileCopyEnvelopes(logFilePaths),
		}
	}

	//We found at least one invalid log files which demands a warning, but let's check if there any valid log file for FilesToCopy
	if len(validFilePaths) > 0 {
		return tasks.Result{
			Status:      tasks.Warning,
			Summary:     resultSummary,
			Payload:     validFilePaths,
			FilesToCopy: tasks.StringsToFileCopyEnvelopes(validFilePaths),
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: resultSummary,
	}
}

// getLogFilePaths - Retrives log_file paths from a slice of config validate elements
func getLogFilePaths(configElements []config.ValidateElement) []string {
	filePaths := []string{}
	//Loop over parsed config elements
	for _, configFile := range configElements {

		//Check if current config element is desired filename
		if configFile.Config.FileName == "newrelic-infra.yml" {

			//Check desired config element for log_file key
			foundKeys := configFile.ParsedResult.FindKey("log_file")

			//Loop over found log_file keys in parsed config
			for _, key := range foundKeys {
				//Extract log_file value
				filePaths = append(filePaths, key.Value())
			}
		}
	}
	return filePaths
}
