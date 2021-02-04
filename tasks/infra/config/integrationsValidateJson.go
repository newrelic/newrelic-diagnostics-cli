package config

import (
	"encoding/json"
	"fmt"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraConfigIntegrationsValidateJson - Validate config and definition files collected from Infra OHAIs
type InfraConfigIntegrationsValidateJson struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigIntegrationsValidateJson) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/IntegrationsValidateJson")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigIntegrationsValidateJson) Explain() string {
	return "Validate json of New Relic Infrastructure on-host integration configuration and definition file(s)"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraConfigIntegrationsValidateJson) Dependencies() []string {
	return []string{
		"Infra/Config/IntegrationsValidate",
	}
}

type jsonValidationError struct {
	fileName  string
	fieldName string
}

var yamlKeyMap = map[string][]string{
	"rabbitmq-config.yml":   []string{"queues", "queues_regexes", "exchanges", "exchanges_regexes", "vhosts", "vhosts_regexes"},
	"redis-config.yml":      []string{"keys"},
	"f5-config.yml":         []string{"partition_filter"},
	"mongodb-config.yml":    []string{"filters"},
	"kafka-config.yml":      []string{"zookeeper_hosts"},
	"postgresql-config.yml": []string{"collection_list"},
}

// Execute - Retrieve all yaml files from definition and config directories for
// both windows and linux.
func (p InfraConfigIntegrationsValidateJson) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	//Check if upstream dependency task status was unsuccessful and bail out if so:
	if upstream["Infra/Config/IntegrationsValidate"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: no validated integrations",
		}
	}

	//Grab validated yml files from Payload
	validatedYamlFiles, ok := upstream["Infra/Config/IntegrationsValidate"].Payload.([]config.ValidateElement)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	yamlInMapCount := 0
	jsonErrors := []jsonValidationError{}

	//Loop through each yaml files in the payload
	for _, yamlFile := range validatedYamlFiles {
		//If ymlFile exists in our yamlKeyMap
		if yamlKeyMap[yamlFile.Config.FileName] != nil {
			yamlInMapCount++
			//Loop through its respective keys in yamlKeyMap
			for _, field := range yamlKeyMap[yamlFile.Config.FileName] {
				foundFields := yamlFile.ParsedResult.FindKey(field)
				//Validate json value, if invalid, append to invalidJson array
				for _, key := range foundFields {
					jsonValue := key.Value()

					//Validate raw JSON value
					if !isValidJson(jsonValue) {
						jsonErrors = append(jsonErrors, jsonValidationError{
							fileName:  yamlFile.Config.FilePath + yamlFile.Config.FileName,
							fieldName: field,
						})
					}
				}
			}
		}
	}

	// Abort if no `*-config.yml` files found
	if yamlInMapCount == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Unable to locate *-config.yml files with known JSON fields to validate.",
		}
	}

	// If invalidJsonErrors > 0, loop through all invalidJsonErrors, concatenate summary for all invalidJsonError,
	//and return tasks.Result failure, notifying user of yaml files with invalid json value fields
	if len(jsonErrors) > 0 {
		summaryString := "Found invalid JSON values in following yml files:\n"

		for _, jsonError := range jsonErrors {
			summaryString += fmt.Sprintf("\t%s: '%s' field contains invalid JSON\n", jsonError.fileName, jsonError.fieldName)
		}

		return tasks.Result{
			Status:  tasks.Failure,
			Summary: summaryString,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("Found and validated %d `*-config.yml` files.", yamlInMapCount),
	}

}

//Helper function to validate raw JSON value
func isValidJson(rawJson string) bool {
	return json.Valid([]byte(rawJson))
}
