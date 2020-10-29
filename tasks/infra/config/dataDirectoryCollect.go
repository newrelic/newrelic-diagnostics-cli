package config

import (
	"errors"
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// InfraConfigDataDirectoryCollect - This struct defined the sample plugin which can be used as a starting point
type InfraConfigDataDirectoryCollect struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	dataDirectoryGetter     dataDirectoryFunc
	dataDirectoryPathGetter dataDirectoryPathFunc
	osType                  string
}
type dataDirectoryFunc func([]string) ([]tasks.FileCopyEnvelope, error)
type dataDirectoryPathFunc func(string) []string

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigDataDirectoryCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/DataDirectoryCollect")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigDataDirectoryCollect) Explain() string {
	return "Collect New Relic Infrastructure agent data directory"
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p InfraConfigDataDirectoryCollect) Dependencies() []string {
	return []string{
		"Infra/Config/Agent",
	}
}

// Execute - The core work within each task
func (p InfraConfigDataDirectoryCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if upstream["Infra/Config/Agent"].Status == tasks.Success {
		p.osType = runtime.GOOS
		filesToCopy, err := p.dataDirectoryGetter(p.dataDirectoryPathGetter(p.osType))

		if err != nil {
			result.Status = tasks.Error
			result.Summary = "Unable to get Infrastructure data directory: " + err.Error()
			return result
		}

		result.Status = tasks.Success
		result.Summary = "New Relic Infrastructure data directory found"
		result.FilesToCopy = filesToCopy
		return result
	}
	result.Summary = "No New Relic Infrastructure agent detected"
	return result
}

func getDataDir(paths []string) ([]tasks.FileCopyEnvelope, error) {
	if len(paths) == 0 {
		return []tasks.FileCopyEnvelope{}, errors.New("No data directory detected")
	}

	files := tasks.FindFiles([]string{".*"}, paths)
	filePaths := []tasks.FileCopyEnvelope{}
	for _, file := range files {
		filePaths = append(filePaths, tasks.FileCopyEnvelope{Path: file, Identifier: "Infra/Config/DataDirectoryCollect"})
	}

	return filePaths, nil
}

func getDataDirPath(osType string) []string {
	if osType == "linux" {
		return []string{"/var/db/newrelic-infra/data"}
	}
	if osType == "windows" {
		return []string{"C:\\Windows\\system32\\config\\systemprofile\\AppData\\Local\\New Relic", "C:\\ProgramData\\New Relic\\newrelic-infra"}
	}
	return []string{}
}
