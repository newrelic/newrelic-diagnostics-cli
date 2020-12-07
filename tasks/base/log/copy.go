package log

import (
	"fmt"
	"os"
	"strings"

	"strconv"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	baseConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// BaseLogCopy - Primary task to search for and find config file. Will optionally take command line input as source
type BaseLogCopy struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseLogCopy) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Log/Copy")
}

// Explain - Returns the help text for each individual task
func (p BaseLogCopy) Explain() string {
	explain := "Collect New Relic log files (has overrides)"
	if config.Flags.ShowOverrideHelp {
		explain += fmt.Sprintf("\n%37s %s", " ", "Override: logpath => set the path of the log file to collect (defaults to finding all logs)")
		explain += fmt.Sprintf("\n%37s %s", " ", "Override: lastModifiedDate => in epochseconds, gathers logs newer than last modified date (defaults to now - 7 days)")
	}
	return explain
}

// Dependencies - No dependencies since this is generally one of the first tasks to run
func (p BaseLogCopy) Dependencies() []string {
	return []string{
		"Base/Env/CollectEnvVars",
		"Base/Env/CollectSysProps",
		"Base/Config/Validate",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseLogCopy) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	logElementsFound := searchForLogPaths(options, upstream)

	if len(logElementsFound) < 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "New Relic log file(s) not found where the " + tasks.ThisProgramFullName + " was executed, or in default agent log file paths. If you can see New Relic logs being generated, you will need to manually provide these logs if you are working with New Relic Support. Review the following document to send the proper level of logging for Support troubleshooting.",
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
		}
	}

	var invalidLogSources []string //will be use to name the log locations we were unable to collect from
	var validLogPaths []string     //will be use to list the filesToCopy into nrdiag.zip
	var failureSummary string

	for idx, logElem := range logElementsFound {
		if logElem.CanCollect {
			logFileStatus := tasks.ValidatePath(logElem.Value)
			if !logFileStatus.IsValid {
				invalidLogSources = append(invalidLogSources, logElem.Source)
				failureSummary += fmt.Sprintf(tasks.ThisProgramFullName+" cannot collect New Relic log files from the provided path(%s):%s\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting\n", logFileStatus.Path, (logFileStatus.ErrorMsg).Error())
				//update our log element to inform that we cannot collect it
				logElementsFound[idx] = LogElement{
					FileName:           logElem.FileName,
					FilePath:           logElem.FilePath,
					Source:             logElem.Source,
					Value:              logElem.Value,
					IsSecureLocation:   logElem.IsSecureLocation,
					CanCollect:         false,
					ReasonToNotCollect: (logFileStatus.ErrorMsg).Error(),
				}
			} else {
				validLogPaths = append(validLogPaths, logElem.Value)
			}
		} else {
			invalidLogSources = append(invalidLogSources, logElem.Source)
			failureSummary += fmt.Sprintf(tasks.ThisProgramFullName+" will not collect the following New Relic log file:%s\n%s\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting\n", logElem.Value, logElem.ReasonToNotCollect)
			logElementsFound[idx] = LogElement{
				FileName:           logElem.FileName,
				FilePath:           logElem.FilePath,
				Source:             logElem.Source,
				Value:              logElem.Value,
				IsSecureLocation:   logElem.IsSecureLocation,
				CanCollect:         false,
				ReasonToNotCollect: logElem.ReasonToNotCollect,
			}
		}
	}

	if len(invalidLogSources) > 0 && len(validLogPaths) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: failureSummary,
			Payload: logElementsFound,
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
		}
	}

	//Let's check the age of those valid files and determine which ones to collect
	lastModifiedDate := getLastModifiedDate(options)
	recentLogFiles, oldLogFiles := determineFilesDate(validLogPaths, lastModifiedDate)
	var filesToCopyToResult []tasks.FileCopyEnvelope
	var successSummary string
	if len(recentLogFiles) > 0 {
		var logPathsList string
		for _, recentLogFile := range recentLogFiles {
			logPathsList = logPathsList + (recentLogFile + "\n")
			filesToCopyToResult = append(filesToCopyToResult, tasks.FileCopyEnvelope{Path: recentLogFile, Identifier: p.Identifier().String()})
		}
		successSummary += fmt.Sprintf("We found at least one recent New Relic log file (modified less than 7 days ago):\n%s", logPathsList)
	}
	if len(oldLogFiles) > 0 {
		mostRecentOldLogFile := selectMostRecentOldLogs(oldLogFiles)
		filesToCopyToResult = append(filesToCopyToResult, tasks.FileCopyEnvelope{Path: mostRecentOldLogFile, Identifier: p.Identifier().String()})
		successSummary += fmt.Sprintf("We found at least one old New Relic log file (modified more than 7 days ago):\n%s", mostRecentOldLogFile)
	}

	if len(filesToCopyToResult) > 0 && len(invalidLogSources) == 0 {
		return tasks.Result{
			Status:      tasks.Success,
			Summary:     successSummary,
			Payload:     logElementsFound,
			FilesToCopy: filesToCopyToResult,
		}
	} else if len(filesToCopyToResult) > 0 && len(invalidLogSources) > 0 {
		warningSummary := fmt.Sprintf("Warning, some log files were not collected:%s\nIf you can see those New Relic logs being generated and they are relevant to this issue, you will need to manually provide those logs if you are working with New Relic Support", strings.Join(invalidLogSources, ", "))
		return tasks.Result{
			Status:      tasks.Warning,
			Summary:     successSummary + "\n" + warningSummary,
			Payload:     logElementsFound,
			FilesToCopy: filesToCopyToResult,
			URL:         "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Payload: logElementsFound,
		Summary: failureSummary,
		URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
	}

}

func initializeDependencies(upstream map[string]tasks.Result) (map[string]string, []baseConfig.ValidateElement) {
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Could not get envVars from upstream")
	}

	configElements, ok := upstream["Base/Config/Validate"].Payload.([]baseConfig.ValidateElement)
	if !ok {
		log.Debug("type assertion failure")
	}
	return envVars, configElements
}

func searchForLogPaths(options tasks.Options, upstream map[string]tasks.Result) []LogElement {

	//get payload from env vars
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Could not get envVars from upstream")
	}
	//get payload from config files
	configElements, ok := upstream["Base/Config/Validate"].Payload.([]baseConfig.ValidateElement)
	if !ok {
		log.Debug("type assertion failure")
	}
	//attempt to find payload in system properties
	var foundSysPropPath string
	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		proccesses, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			for _, process := range proccesses {
				sysPropVal, isPresent := process.SysPropsKeyToVal[logSysProp]
				if isPresent {
					foundSysPropPath = sysPropVal
				}
			}
		}
	}

	var logFilesFound []LogElement
	if options.Options["logpath"] != "" {
		logFilesFound = append(logFilesFound, LogElement{
			Source:           "Defined by user through the " + tasks.ThisProgramFullName + " flag: logpath",
			Value:            options.Options["logpath"],
			IsSecureLocation: false,
		})
	} else {
		logFilesFound = collectFilePaths(envVars, configElements, foundSysPropPath) //At this point foundSysPropPath may be not be have an assigned value but we'll check for length on the other end
		for i, logFileFound := range logFilesFound {
			if logFileFound.IsSecureLocation {
				question := fmt.Sprintf("We've found a file that may contain secure information: %s\n", logFileFound.Value) +
					"Include this file in nrdiag-output.zip?"
				if !(tasks.PromptUser(question, options)) {
					logFilesFound[i] = LogElement{
						Source:             logFileFound.Source,
						Value:              logFileFound.Value,
						CanCollect:         false,
						ReasonToNotCollect: "User opted out when " + tasks.ThisProgramFullName + " asked if it can collect this file that may contain secure information.",
					}
				}
			}

		}
	}
	return logFilesFound
}

func determineFilesDate(logFilePaths []string, lastModifiedDate time.Time) ([]string, []string) {
	var recentLogFilePaths []string
	var oldLogFilePaths []string

	for _, logFilePath := range logFilePaths {

		if isLogFileRecent(logFilePath, lastModifiedDate) {
			recentLogFilePaths = append(recentLogFilePaths, logFilePath)
		} else {
			oldLogFilePaths = append(oldLogFilePaths, logFilePath)
		}
	}
	return recentLogFilePaths, oldLogFilePaths
}

func getLastModifiedDate(options tasks.Options) time.Time {
	var lastModifiedDate time.Time
	if options.Options["lastModifiedDate"] != "" {
		i, err := strconv.ParseInt(options.Options["lastModifiedDate"], 10, 64)
		lastModifiedDate = time.Unix(i, 0)
		if err != nil {
			log.Info("Error parsing input time override: ", options.Options["lastModifiedDate"])
		}
		log.Debug("setting override lastModifiedDate to:", lastModifiedDate)
	} else {
		lastModifiedDate = time.Now().AddDate(0, 0, -7)
		log.Debug("Default last modified date is:", lastModifiedDate)
	}
	return lastModifiedDate

}

func isLogFileRecent(inputFilePath string, minimumModTime time.Time) bool {
	fileInfo, err := os.Stat(inputFilePath)
	if err != nil {
		log.Debug("Error reading file", inputFilePath)
		return true
	}

	return fileInfo.ModTime().After(minimumModTime)
}

func selectMostRecentOldLogs(logFilePaths []string) string {
	var mostRecentTime time.Time
	var logFilePathSelected string

	for index, logFilePath := range logFilePaths {
		fileInfo, err := os.Stat(logFilePath)
		whenFileWasModified := fileInfo.ModTime()
		if err != nil {
			log.Debug("Error reading file", logFilePath)
			return ""
		}
		if index == 0 || whenFileWasModified.After(mostRecentTime) {
			mostRecentTime = whenFileWasModified
			logFilePathSelected = logFilePath
		}

	}
	return logFilePathSelected
}
