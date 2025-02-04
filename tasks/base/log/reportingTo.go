package log

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseLogReportingTo - This struct defined the sample plugin which can be used as a starting point
type BaseLogReportingTo struct {
}
type LogNameReportingTo struct {
	Logfile     string
	ReportingTo []string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseLogReportingTo) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Log/ReportingTo")
}

// Explain - Returns the help text for each individual task
func (t BaseLogReportingTo) Explain() string {
	return "Determine New Relic account id and application id"
}

// Dependencies - Returns the dependencies for ech task.
func (t BaseLogReportingTo) Dependencies() []string {
	return []string{"Base/Log/Copy"}
}

// Execute - The core work within each task
// Calls taskHelpers.ReturnStringInFile with "Reporting to:[^\n]*" specified.
func (t BaseLogReportingTo) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	logElements, ok := upstream["Base/Log/Copy"].Payload.([]LogElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Logs not found",
		}
	}

	var logs []LogElement
	for _, log := range logElements {
		//IsSecureLocation represents non-new relic log files such as docker syslog
		if !log.CanCollect || len(log.FileName) == 0 || len(log.FilePath) == 0 || log.IsSecureLocation {
			continue
		}
		logs = append(logs, log)
	}

	if len(logs) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "New Relic logs not found",
		}
	}

	logsWithReportingTo := readUpstream(logs)
	log.Debug("Full reporting to", logsWithReportingTo)

	if len(logsWithReportingTo) > 0 {

		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Found a reporting to.",
			Payload: logsWithReportingTo,
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: "Logs founds, no instances of reporting to within logs.",
	}
}

func findReportingToLine(filepath string) (string, error) {
	searchString := "Reporting to:[^\n]*"
	reportingTo, err := tasks.ReturnLastStringSubmatchInFile(searchString, filepath)
	if err != nil {
		return "", err
	}
	if len(reportingTo) < 1 {
		return "", nil
	}

	log.Debug("Reporting to", reportingTo)
	sanitizedReportingTo := sanitizeLogEntry(reportingTo[0])

	// First element is the full-line match
	return sanitizedReportingTo, nil
}

func sanitizeLogEntry(logLine string) string {
	lineWithoutText := strings.Replace(logLine, "Reporting to: ", "", -1)
	indexOfQuote := strings.Index(lineWithoutText, "\"")
	if indexOfQuote == -1 {
		return lineWithoutText
	}
	trimmedLogLine := strings.SplitN(lineWithoutText, "\"", 2)

	return strings.Replace(trimmedLogLine[0], "\\", "", -1)

}

func readUpstream(logs []LogElement) []LogNameReportingTo {
	var logResults []LogNameReportingTo
	for _, l := range logs {
		searchResult, err := findReportingToLine(l.FilePath + l.FileName)
		if err != nil {
			log.Debug(err)
			continue
		}

		if len(searchResult) == 0 {
			continue
		}

		logResult := LogNameReportingTo{
			Logfile:     l.FilePath + l.FileName,
			ReportingTo: []string{searchResult},
		}

		logResults = append(logResults, logResult)

	}
	return logResults
}
