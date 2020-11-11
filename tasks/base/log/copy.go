package log

import (
	"fmt"
	"os"

	"path/filepath"
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
		"Base/Config/Validate",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseLogCopy) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	envVars, configElements := initializeDependencies(upstream)

	logFilesPaths := searchForLogPaths(options, envVars, configElements)

	if len(logFilesPaths) < 1 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "New Relic log file(s) not found where the " + tasks.ThisProgramFullName + " was executed, or in default agent log file paths. If you can see New Relic logs being generated, you will need to manually provide these logs if you are working with New Relic Support. Review the following document to send the proper level of logging for Support troubleshooting.",
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/generate-new-relic-agent-logs-troubleshooting",
		}
	}

	logFilesStatuses := tasks.ValidatePaths(logFilesPaths)

	var invalidLogFiles, validLogFiles []string
	var resultSummary string

	for _, file := range logFilesStatuses {
		if !file.IsValid {
			invalidLogFiles = append(invalidLogFiles, file.Path)
			resultSummary += fmt.Sprintf("Warning! the " + tasks.ThisProgramFullName + " cannot collect New Relic log files from the provided path(%s):%s.If you are working with a support ticket, manually provide your New Relic log file for further troubleshooting\n", file.Path, (file.ErrorMsg).Error())
		} else {
			validLogFiles = append(validLogFiles, file.Path)
		}
	}

	if len(invalidLogFiles) > 0 && len(validLogFiles) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: resultSummary,
		}
	}

	//Let's check the age of those valid files and determine which ones to collect
	lastModifiedDate := getLastModifiedDate(options)
	recentLogFiles, oldLogFiles := determineFilesDate(validLogFiles, lastModifiedDate)
	var filesToCopyToResult []tasks.FileCopyEnvelope
	var resultsPayload []LogElement

	if len(recentLogFiles) > 0 {
		for _, path := range recentLogFiles {
			filesToCopyToResult = append(filesToCopyToResult, tasks.FileCopyEnvelope{Path: path, Identifier: p.Identifier().String()})
			dir, fileName := filepath.Split(path)
			resultsPayload = append(resultsPayload, LogElement{fileName, dir})
		}
		resultSummary += fmt.Sprintf("We found at least one recent New Relic log file (modified less than 7 days ago): %s\n", recentLogFiles[0])
	}
	if len(oldLogFiles) > 0 {
		mostRecentOldLogFile := selectMostRecentOldLogs(oldLogFiles)
		filesToCopyToResult = append(filesToCopyToResult, tasks.FileCopyEnvelope{Path: mostRecentOldLogFile, Identifier: p.Identifier().String()})
		dir, fileName := filepath.Split(mostRecentOldLogFile)
		resultsPayload = append(resultsPayload, LogElement{fileName, dir})
		resultSummary += fmt.Sprintf("We found at least one old New Relic log file (modified more than 7 days ago): %s\n", mostRecentOldLogFile)
	}

	if len(filesToCopyToResult) > 0 {
		return tasks.Result{
			Status:      tasks.Success,
			Summary:     resultSummary,
			Payload:     resultsPayload,
			FilesToCopy: filesToCopyToResult,
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: resultSummary,
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

func searchForLogPaths(options tasks.Options, envVars map[string]string, configElements []baseConfig.ValidateElement) []string {
	var logFilesPaths []string
	var secureLogFilesPaths []string

	if options.Options["logpath"] != "" {
		logFilesPaths = []string{options.Options["logpath"]}
	} else {
		logFilesPaths, secureLogFilesPaths = collectFilePaths(envVars, configElements)
		if secureLogFilesPaths != nil {
			for _, secureLog := range secureLogFilesPaths {
				question := fmt.Sprintf("We've found a file that may contain secure information: %s\n", secureLog) +
					"Include this file in nrdiag-output.zip?"
				if tasks.PromptUser(question, options) {
					if !config.Flags.Quiet {
						log.Info("adding file ", secureLog)
					}
					logFilesPaths = append(logFilesPaths, secureLog)
				}
			}
		}
	}
	return logFilesPaths
}

func determineFilesDate(logFilesPaths []string, lastModifiedDate time.Time) ([]string, []string) {
	var recentLogFiles []string
	var oldLogFiles []string

	for _, path := range logFilesPaths {
		if isLogFileRecent(path, lastModifiedDate) {
			recentLogFiles = append(recentLogFiles, path)
		} else {
			oldLogFiles = append(oldLogFiles, path)
		}
	}
	return recentLogFiles, oldLogFiles
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

func selectMostRecentOldLogs(logFilesPaths []string) string {
	var mostRecentTime time.Time
	var fileSelected string

	for index, logFile := range logFilesPaths {
		fileInfo, err := os.Stat(logFile)
		whenFileWasModified := fileInfo.ModTime()
		if err != nil {
			log.Debug("Error reading file", logFile)
			return ""
		}
		if index == 0 || whenFileWasModified.After(mostRecentTime) {
			mostRecentTime = whenFileWasModified
			fileSelected = logFile
		}

	}
	return fileSelected
}
