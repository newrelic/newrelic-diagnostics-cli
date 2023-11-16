package log

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraLogCollect - This struct defined the sample plugin which can be used as a starting point
type InfraLogCollect struct {
	validatePaths func([]string) []tasks.CollectFileStatus
	findFiles     func([]string, []string) []string
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
		"Base/Env/CollectEnvVars",
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

	configElements := []config.ValidateElement{}
	if upstream["Base/Config/Validate"].HasPayload() {
		getConfigElements, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)
		if ok {
			configElements = getConfigElements
		}
	}

	envVars := make(map[string]string)
	if upstream["Base/Env/CollectEnvVars"].Status == tasks.Info && upstream["Base/Env/CollectEnvVars"].HasPayload() {
		getEnvVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
		if ok {
			envVars = getEnvVars
		}
	}

	logFilePaths := p.getLogFilePaths(configElements, envVars)

	if len(logFilePaths) < 1 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "New Relic Infrastructure agent log file path not found in the configuration file or environment variables.",
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
			resultSummary += fmt.Sprintf("The log file path found (%s) did not provide a file that was accessible to us:\n%s\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting", file.Path, file.ErrorMsg.Error())
			continue
		}
		validFilePaths = append(validFilePaths, file.Path)
		resultSummary += fmt.Sprintf("Success, logs found! We were able to access the following New Relic Infrastructure agent log file:%s\n", file.Path)
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

// getLogFilePaths - Retrieves log files paths from a slice of config validate elements
func (p InfraLogCollect) getLogFilePaths(configElements []config.ValidateElement, envVars map[string]string) []string {
	filePaths := []string{}
	if len(configElements) < 1 && len(envVars) < 1 {
		return filePaths
	}

	// By default, windows creates a log in ProgramData
	if runtime.GOOS == "windows" {
		sysProgramData, ok := envVars["ProgramData"]
		if ok && sysProgramData != "" {
			filePaths = append(filePaths, sysProgramData+`\New Relic\newrelic-infra\newrelic-infra.log`) //Windows, agent version 1.0.944 or higher
		}
	}

	// Make sure to check the env variable too
	nriaLogFile, ok := envVars["NRIA_LOG_FILE"]
	if ok && nriaLogFile != "" {
		filePaths = append(filePaths, nriaLogFile)
	}

	//Loop over parsed config elements
	for _, configFile := range configElements {

		//Check if current config element is desired filename
		if configFile.Config.FileName == "newrelic-infra.yml" {

			//Check desired config element for log/file new key configuration option
			logFile := configFile.ParsedResult.FindKeyByPath("/log/file").Value()
			//New log configuration allows log rotation with custom timestamp on log filename
			if logFile != "" {
				// search for log, rotated and gz files (filename is made with the base name of the logfile)
				searchPattern := []string{"^" + strings.TrimSuffix(filepath.Base(logFile), filepath.Ext(logFile)) + ".+" + `(\.gz|\.zip|\` + filepath.Ext(logFile) + ")"}
				// Gather all files
				filePaths = append(filePaths, p.findFiles(searchPattern, []string{filepath.Dir(logFile)})...)
			} else {
				//Check desired config element for log_file old key configuration option
				foundKeys := configFile.ParsedResult.FindKey("log_file")

				//Loop over found log_file keys in parsed config
				for _, key := range foundKeys {
					//Extract log_file value
					filePaths = append(filePaths, key.Value())
				}
			}

		}
	}
	return filePaths
}
