package log

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	baseConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
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
		"Base/Env/CollectSysProps",
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

	filesToCopy := []tasks.FileCopyEnvelope{}

	var logs []LogElement
	if options.Options["logpath"] != "" {
		logs = append(logs, LogElement{
			Source: LogSourceData{
				FoundBy: logPathDiagnosticsFlagSource,
			},
			IsSecureLocation: false,
		})
	} else {
		//ignoring secure logs for now
		logs = collectFilePaths(envVars, configElements, foundSysProps, options)
	}

	if len(logs) > 0 {
		result.Status = tasks.Success
		// format the output of the result to return the files found and their content

		for _, log := range logs {
			dir, fileName := filepath.Split(log.Source.FullPath)
			log.FileName = fileName
			log.FilePath = dir
			ch, _ := prunedReader(log.Source.FullPath)

			filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: log.Source.FullPath, Stream: ch, Identifier: p.Identifier().String()})

		}
		// now add the results into a single json string
		log.Debug("all logs found", logs)

		result.Payload = logs
		result.Summary = fmt.Sprintf("There were %d file(s) found", len(logs))
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
}

