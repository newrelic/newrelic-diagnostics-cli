package config

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraConfigIntegrationsCollect - This struct defined the sample plugin which can be used as a starting point
type InfraConfigIntegrationsCollect struct {
	fileFinder func([]string, []string) []string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigIntegrationsCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/IntegrationsCollect")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigIntegrationsCollect) Explain() string {
	return "Collect New Relic Infrastructure on-host integration configuration and definition files"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraConfigIntegrationsCollect) Dependencies() []string {
	return []string{
		"Infra/Config/Agent",
	}
}

// Execute - Retrieve all yml files from definition and config directories for
// both windows and linux.
func (p InfraConfigIntegrationsCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No Infra Agent detected. Task not executed.",
		}
	}
	var configPaths []string

	if runtime.GOOS == "windows" {
		configPaths = []string{
			`C:\Program Files\New Relic\newrelic-infra\`,
		}
	} else {
		configPaths = []string{
			"/etc/newrelic-infra/",
			"/var/db/newrelic-infra/",
		}
	}

	configPatterns := []string{".+[.]y(a)?ml$"}

	configFiles := p.fileFinder(configPatterns, configPaths)

	if len(configFiles) > 0 {
		var configElements []config.ConfigElement
		var fileCopyEnvelopes []tasks.FileCopyEnvelope
		for _, file := range configFiles {
			dir, fileName := filepath.Split(file)
			configElements = append(configElements, config.ConfigElement{FileName: fileName, FilePath: dir})

			question := fmt.Sprintf("We've found a file that may contain secure information: %s\n", file) +
				"Include this file in nrdiag-output.zip?"
			if tasks.PromptUser(question, options) {
				fileCopyEnvelopes = append(fileCopyEnvelopes, tasks.FileCopyEnvelope{Path: file})
			}
		}
		return tasks.Result{
			Status:      tasks.Success,
			Summary:     fmt.Sprintf("%d on-host integration yml file(s) found", len(configFiles)),
			Payload:     configElements,
			FilesToCopy: fileCopyEnvelopes,
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: "No on-host integration yml files found",
	}
}
