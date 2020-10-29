package config

import (
	"fmt"
	"os"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraConfigIntegrationsValidate - Validate config and definition files collected from Infra OHAIs
type InfraConfigIntegrationsValidate struct {
	fileReader func(string) (*os.File, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigIntegrationsValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/IntegrationsValidate")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigIntegrationsValidate) Explain() string {
	return "Validate yml formatting of New Relic Infrastructure on-host integration configuration and definition file(s)"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraConfigIntegrationsValidate) Dependencies() []string {
	return []string{
		"Infra/Config/IntegrationsCollect",
	}
}

type validationError struct {
	fileLocation string
	errorText    string
}

// Execute - Retrieve all yaml files from definition and config directories for
// both windows and linux.
func (p InfraConfigIntegrationsValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Infra/Config/IntegrationsCollect"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No On-host Integration config and definitions files were collected. Task not executed.",
		}
	}

	yamlLocations, ok := upstream["Infra/Config/IntegrationsCollect"].Payload.([]config.ConfigElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	if len(yamlLocations) == 0 { //upstream payload is correct type, but upstream task found no config/definition yamls
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No On-host Integration config or definition files were found. Task not executed.",
		}
	}

	validatedYamls, validationErrors := p.validateYamls(yamlLocations)

	if len(validationErrors) > 0 {
		taskFailureSummary := "Error validating on-host integration configuration files:\n"

		for _, e := range validationErrors {
			taskFailureSummary += fmt.Sprintf("%s: %s\n", e.fileLocation, e.errorText)
		}

		return tasks.Result{
			Status:  tasks.Failure,
			Summary: taskFailureSummary,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("Successfully validated %v yaml file(s)", len(validatedYamls)),
		Payload: validatedYamls,
	}

}

//validateYamls open, lint and parse Yaml files, return slice of errors and slice of successfully parsed Yamls
func (p InfraConfigIntegrationsValidate) validateYamls(yamlLocations []config.ConfigElement) ([]config.ValidateElement, []validationError) {
	validationErrors := []validationError{}
	validatedYamls := []config.ValidateElement{}

	//Read and lint yamls
	for _, yamlLocation := range yamlLocations {

		yamlPath := yamlLocation.FilePath + yamlLocation.FileName
		log.Debug("validating " + yamlPath)

		//Read in yaml file
		file, err := p.fileReader(yamlPath)
		if err != nil {
			validationErrors = append(validationErrors,
				validationError{
					fileLocation: yamlPath,
					errorText:    "Unable to read yaml: " + err.Error(),
				})
			continue
		}
		defer file.Close()

		//Parse raw yaml to tasks.ValidateBlob
		parsedConfig, err := config.ParseYaml(file)
		if err != nil { //This indicates there was a parsing error on the yaml file
			validationErrors = append(validationErrors,
				validationError{
					fileLocation: yamlPath,
					errorText:    "Unable to parse yaml: " + err.Error(),
				})
			continue
		}

		//Append parsed yaml blob to collection of validated yamls
		validatedYamls = append(validatedYamls,
			config.ValidateElement{
				Config:       yamlLocation,
				ParsedResult: parsedConfig,
			})
	}
	return validatedYamls, validationErrors
}
