package agent

import (
	"bufio"
	"errors"
	"os"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// iOSAgentVersion - This struct defined the sample plugin which can be used as a starting point
type iOSAgentVersion struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t iOSAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("iOS/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (t iOSAgentVersion) Explain() string {
	return "Determine New Relic iOS agent version"
}

// Dependencies - Returns the dependencies for ech task.
func (t iOSAgentVersion) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"iOS/Env/Detect",
	}
}

// Execute - The core work within each task
func (t iOSAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	if upstream["iOS/Env/Detect"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Task did not meet requirements necessary to run: iOS environment not detected."
		return result
	}

	configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement) // This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if !ok {
		result.Status = tasks.None
		result.Summary = "Task did not meet requirements necessary to run: type assertion failure"
		return result
	}

	pathResult, err := confirmHeaderFile(configs)
	if err != nil {
		result.Status = tasks.None
		result.Summary = err.Error()
		return result
	}
	log.Debug(pathResult)

	versionResult, err := parseHeaderFile(pathResult)
	if err != nil {
		result.Status = tasks.Warning
		result.Summary = "Unable to confirm iOS agent version: " + err.Error()
		return result
	}
	log.Debug(versionResult)
	result.Status = tasks.Info
	result.Summary = "Mobile iOS agent version: " + versionResult
	result.Payload = versionResult
	return result
}

func confirmHeaderFile(configs []config.ConfigElement) (string, error) {
	for _, currentConfig := range configs {
		if currentConfig.FileName == "NewRelic.h" {
			return currentConfig.FilePath + currentConfig.FileName, nil
		}
	}
	return "", errors.New("agent header file not found")
}

func parseHeaderFile(headerFilePath string) (string, error) {
	var lineRef = "// Using New Relic Agent Version: " // This is the first line of the header file, the next chars after this string should be the version.

	fileHandle, err := os.Open(headerFilePath)
	if err != nil {
		return "", err
	}
	defer fileHandle.Close()

	fileScanner := bufio.NewScanner(fileHandle)
	for fileScanner.Scan() {
		if strings.Contains(fileScanner.Text(), lineRef) {
			return strings.Trim(fileScanner.Text(), lineRef), nil
		}
	}
	return "", errors.New("version reference not found within header file")
}
