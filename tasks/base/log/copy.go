package log

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

	logElements := dedupeLogPaths(logElementsFound)

	var invalidLogPaths []string //will be use to name the log locations we were unable to collect from
	var validLogPaths []string   //will be use to list the filesToCopy into nrdiag.zip
	var failureSummary string

	for idx, logElem := range logElements {
		if logElem.CanCollect {
			logFileStatus := tasks.ValidatePath(logElem.Source.FullPath)
			if !logFileStatus.IsValid {
				invalidLogPaths = append(invalidLogPaths, logElem.Source.FullPath)
				failureSummary += fmt.Sprintf(tasks.ThisProgramFullName+" cannot collect New Relic log files from the provided path(%s):%s\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting\n", logFileStatus.Path, (logFileStatus.ErrorMsg).Error())
				//update our log element to inform that we cannot collect it
				logElements[idx] = setLogElement(logElem.FileName, logElem.FilePath, logElem.Source, logElem.IsSecureLocation, false, (logFileStatus.ErrorMsg).Error())
			} else {
				validLogPaths = append(validLogPaths, logElem.Source.FullPath)
			}
		} else { //cases where customer rejected the prompt to collect the file
			invalidLogPaths = append(invalidLogPaths, logElem.Source.FullPath)
			failureSummary += fmt.Sprintf("%s\nIf you are working with a support ticket, manually provide your New Relic log file for further troubleshooting\n", logElem.ReasonToNotCollect)
			logElements[idx] = setLogElement(logElem.FileName, logElem.FilePath, logElem.Source, logElem.IsSecureLocation, false, logElem.ReasonToNotCollect)
		}
	}

	hasInvalidLogs := len(invalidLogPaths) > 0
	hasValidLogs := len(validLogPaths) > 0

	if hasValidLogs {
		var filesToCopyToResult []tasks.FileCopyEnvelope
		var successSummary = "Succesfully collected one or more New Relic Log file(s). Those file names will be listed in the nrdiag-output.json, under the payload section with the field 'CanCollect' set to true.\n"
		for _, validPath := range validLogPaths {
			filesToCopyToResult = append(filesToCopyToResult, tasks.FileCopyEnvelope{
				Path:       validPath,
				Identifier: p.Identifier().String(),
			})
		}
		//Look for NET log files. There are too many so we'll only include one file in the payload. By now all files should had been captured as part of filesToCopyToResult
		var resultPayload interface{}
		if hasDotnetLogs(logElements) {
			resultPayload = filterDotnetLogElements(logElements)
		} else {
			resultPayload = logElements
		}

		if hasInvalidLogs {
			warningSummary := fmt.Sprintf("Warning, some log files were not collected:%s\nIf those logs are relevant to this issue, you will need to manually provide those logs if you are working with New Relic Support.", strings.Join(invalidLogPaths, ", "))
			return tasks.Result{
				Status:      tasks.Warning,
				Summary:     successSummary + "\n" + warningSummary,
				Payload:     resultPayload,
				FilesToCopy: filesToCopyToResult,
				URL:         "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
			}
		}
		//we only have valid logs
		return tasks.Result{
			Status:      tasks.Success,
			Summary:     successSummary,
			Payload:     resultPayload,
			FilesToCopy: filesToCopyToResult,
		}
	}
	// we have no valid logs, only invalid
	return tasks.Result{
		Status:  tasks.Failure,
		Summary: failureSummary,
		Payload: logElements,
		URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
	}
}

func filterDotnetLogElements(logElements []LogElement) []LogElement {
	var filteredLogElements []LogElement
	profilerRgx := regexp.MustCompile(profilerLogName)
	foundFirstDotnetLog := false
	for _, log := range logElements {
		if profilerRgx.MatchString(log.FileName) {
			if !foundFirstDotnetLog {
				filteredLogElements = append(filteredLogElements, setLogElement(log.FileName, log.FilePath, log.Source, log.IsSecureLocation, true, dotnetLogsDownsizeExplanation))
				foundFirstDotnetLog = true
			}
			continue
		}
		filteredLogElements = append(filteredLogElements, log)
	}
	// filteredLogElements = append(filteredLogElements, setLogElement())
	return filteredLogElements
}
func hasDotnetLogs(logElements []LogElement) bool {
	profilerRgx := regexp.MustCompile(profilerLogName)
	for _, log := range logElements {
		if profilerRgx.MatchString(log.FileName) {
			return true
		}
	}
	return false
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
		log.Debug("type assertion failure for Base/Config/Validate Payload")
	}
	//attempt to find system properties related to logs in payload
	foundSysProps := make(map[string]string)
	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		proccesses, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			for _, process := range proccesses {
				for _, sysPropKey := range logSysProps {
					sysPropVal, isPresent := process.SysPropsKeyToVal[sysPropKey]
					if isPresent {
						foundSysProps[sysPropKey] = sysPropVal
					}
				}
			}
		}
	}

	var logFilesFound []LogElement
	if options.Options["logpath"] != "" {
		logSourceData := LogSourceData{
			FoundBy:  logPathDiagnosticsFlagSource,
			FullPath: options.Options["logpath"],
			KeyVals: map[string]string{
				"logpath": options.Options["logpath"],
			},
		}
		dir, fileName := filepath.Split(options.Options["logpath"])
		logFilesFound = append(logFilesFound, setLogElement(fileName, dir, logSourceData, false, true, ""))
	} else {
		logFilesFound = collectFilePaths(envVars, configElements, foundSysProps, options) //At this point foundSysPropPath may be not be have an assigned value but we'll check for length on the other end
		for i, logFileFound := range logFilesFound {
			if logFileFound.IsSecureLocation {
				question := fmt.Sprintf("We've found a file that may contain secure information: %s\n", logFileFound.Source.FullPath) +
					"Include this file in nrdiag-output.zip?"
				if !(tasks.PromptUser(question, options)) {
					//update logElement with new data
					reasonCannotCollect := "User opted out when " + tasks.ThisProgramFullName + " asked if it can collect this file that may contain secure information."
					logFilesFound[i] = setLogElement(logFileFound.FileName, logFileFound.FilePath, logFileFound.Source, true, false, reasonCannotCollect)
				}
			}

		}
	}
	return logFilesFound
}

func determineFilesDate(logFilePaths []string, lastModifiedDate time.Time) ([]string, []string) {
	var recentLogFilePaths []string
	var oldLogFilePaths []string
	profilerRgx := regexp.MustCompile(profilerLogName)
	for _, logFilePath := range logFilePaths {
		//overwrite lastModifiedDate for .NET profiler logs because it produces too many files. We'll only collect profiler logs produced in the last 4 days instead of 7.
		if profilerRgx.MatchString(logFilePath) {
			lastModifiedDate = time.Now().AddDate(0, 0, -profilerMaxNumDays)
		}
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
		lastModifiedDate = time.Now().AddDate(0, 0, -defaultMaxNumDays)
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
