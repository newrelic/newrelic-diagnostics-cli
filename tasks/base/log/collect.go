package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/newrelic/NrDiag/config"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	baseConfig "github.com/newrelic/NrDiag/tasks/base/config"
)

// BaseLogCollect - Primary task to search for and find config file. Will optionally take command line input as source
type BaseLogCollect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseLogCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Log/Collect")
}

// Explain - Returns the help text for each individual task
func (p BaseLogCollect) Explain() string {
	explain := "Collect New Relic log files (has overrides)"
	if config.Flags.ShowOverrideHelp {
		explain += fmt.Sprintf("\n%37s %s", " ", "Override: logpath => set the path of the log file to collect (defaults to finding all logs)")
	}
	return explain
}

// Dependencies - No dependencies since this is generally one of the first tasks to run
func (p BaseLogCollect) Dependencies() []string {
	// no dependencies!
	return []string{
		"Base/Env/CollectEnvVars",
		"Base/Config/Validate",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseLogCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Could not get envVars from upstream")
	}

	configElements, ok := upstream["Base/Config/Validate"].Payload.([]baseConfig.ValidateElement)
	if !ok {
		log.Debug("type assertion failure")
	}

	results := []LogElement{}
	filesToCopy := []tasks.FileCopyEnvelope{}

	var logs []string
	if options.Options["logpath"] != "" {
		logs = []string{options.Options["logpath"]}
	} else {
		//ignoring secure logs for now
		logs, _ = collectFilePaths(envVars, configElements)
	}

	if logs != nil {
		result.Status = tasks.Success
		// format the output of the result to return the files found and their content

		for _, path := range logs {
			dir, fileName := filepath.Split(path)

			c := LogElement{fileName, dir}
			results = append(results, c)
			ch, _ := prunedReader(path)

			filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: path, Stream: ch, Identifier: p.Identifier().String()})

		}
		// now add the results into a single json string
		log.Debug("results", results)

		result.Payload = results
		result.Summary = fmt.Sprintf("There were %d file(s) found", len(results))
		result.FilesToCopy = filesToCopy

	} else {
		// Log File not found, does it exist?
		result.Status = tasks.Failure
		result.Summary = "Log File not found, please check working directory"
		result.URL = "https://docs.newrelic.com/docs/new-relic-diagnostics#run-diagnostics"
	}

	return result
}

func prunedReader(path string) (c chan string, err error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	logChannel := make(chan string, 10)

	go pruneLog(file, logChannel)
	return logChannel, nil
}

func pruneLog(file *os.File, logChannel chan string) {
	defer file.Close()

	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	var err error
	var line string

	for {
		line, err = reader.ReadString('\n')

		if err != nil {
			break
		}
		logChannel <- line
	}

	if err != io.EOF {
		log.Debug("Log prune failed: ", err)
	}

	close(logChannel)

	return
}

func limitLength(s string, length int) string {
	if len(s) < length {
		return s
	}

	return s[:length]
}
