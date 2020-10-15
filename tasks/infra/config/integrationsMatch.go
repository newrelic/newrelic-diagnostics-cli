package config

import (
	"fmt"
	"regexp"

	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

// InfraConfigIntegrationsMatch - attempt to match Infra OHAI config and definition files
type InfraConfigIntegrationsMatch struct {
	runtimeOS string
}

// IntegrationFilePair - This struct defines a pair of integration files
type IntegrationFilePair struct {
	Configuration config.ValidateElement
	Definition    config.ValidateElement
}

//MatchedIntegrationFiles - This struct represents the payload for the task
type MatchedIntegrationFiles struct {
	IntegrationFilePairs map[string]*IntegrationFilePair
	Errors               []IntegrationMatchError
}

//IntegrationMatchError - This struct represents any validation errors encountered in the task
type IntegrationMatchError struct {
	IntegrationFile config.ValidateElement
	Reason          string
}

//These variables represent the expected filename patterns for the configuration and definition integration files
var (
	configFileRegex     = regexp.MustCompile(`(.+)\-(config)\.(yaml|yml)`)
	definitionFileRegex = regexp.MustCompile(`(.+)\-(definition)\.(yaml|yml)`)
)

//IntegrationFileType - Establishes enum type for integration file categorical types "config" and "definition"
type IntegrationFileType string

//Implementation of IntegrationFileType enums for integration file categorical types "config" and "definition"
const (
	CONFIG     IntegrationFileType = "config"
	DEFINITION IntegrationFileType = "definition"
)

//These constants are the expected filepaths for the integration file types
var (
	definitionFilepathsLinux     = []string{"/var/db/newrelic-infra/custom-integrations/", "/var/db/newrelic-infra/newrelic-integrations/"}
	definitionFilepathsWindows   = []string{"C:\\Program Files\\New Relic\\newrelic-infra\\custom-integrations\\", "C:\\Program Files\\New Relic\\newrelic-infra\\newrelic-integrations\\"}
	configurationFilepathLinux   = "/etc/newrelic-infra/integrations.d/"
	configurationFilepathWindows = "C:\\Program Files\\New Relic\\newrelic-infra\\integrations.d\\"
)

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigIntegrationsMatch) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/IntegrationsMatch")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigIntegrationsMatch) Explain() string {
	return "Validate New Relic Infrastructure on-host integration configuration and definition file pairs"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraConfigIntegrationsMatch) Dependencies() []string {
	return []string{"Infra/Config/IntegrationsValidate"}
}

// Execute - The core work within each task
func (p InfraConfigIntegrationsMatch) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	// Check if upstream dependency task status was unsuccessful and bail out if so:
	if upstream["Infra/Config/IntegrationsValidate"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: no validated integrations",
		}
	}

	// Grab parsed config files and perform type assertion
	integrationFiles, ok := upstream["Infra/Config/IntegrationsValidate"].Payload.([]config.ValidateElement)

	// If type assertion failed, bail out
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	//Sort files from upstream into a map of integration names (key) with pairs of found config and definition files (IntegrationFilePair)
	//matchFilePairs will also validate filepath of the integration file. Filepath errors get collected in matchErrors
	filePairs, matchErrors := p.matchFilePairs(integrationFiles)

	//Validate found IntegrationFilePairs to make sure they actually are pairs of definition and configuration files.
	//If orphan definition or configuration files are found they are removed from the map and collected as IntegrationMatchErrors
	//Valid filePairs have their yaml integration names validated for a match, else they are also collected into IntegrationMatchErrors
	filePairs, validationErrors := validateIntegrationPairs(filePairs)

	//Merge all found IntegrationMatchErrors
	matchErrors = append(matchErrors, validationErrors...)

	var matchErrorSummary string

	//Build a single string of all IntegrationMatchError reasons separated by newline
	if len(matchErrors) > 0 {
		matchErrorSummary = buildMatchErrorSummary(matchErrors)
	}

	if (len(filePairs) > 0) && (len(matchErrors) == 0) {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Found matching integration files",
			Payload: MatchedIntegrationFiles{
				IntegrationFilePairs: filePairs,
				Errors:               matchErrors,
			},
		}
	} else if (len(filePairs) > 0) && (len(matchErrors) > 0) {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "Found matching integration files with some errors: " + matchErrorSummary,
			Payload: MatchedIntegrationFiles{
				IntegrationFilePairs: filePairs,
				Errors:               matchErrors,
			},
			URL: "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
		}
	}

	//implicit case: (len(configPairs) == 0 && len(matchErrors) >= 0
	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "No matching integration files found" + matchErrorSummary,
		Payload: MatchedIntegrationFiles{
			IntegrationFilePairs: filePairs,
			Errors:               matchErrors,
		},
		URL: "https://docs.newrelic.com/docs/integrations/integrations-sdk/getting-started/integration-file-structure-activation",
	}

}

//This method identifies Integration files from ValidateElements that are config or definition files by pattern match
//Validates found Integration file to expected filepath, if invalid validate element is put in IntegrationMatchError struct
//Adds validate element to map in a IntegrationFilePair struct where keys are integration name derived from filename
func (p InfraConfigIntegrationsMatch) matchFilePairs(integrationFiles []config.ValidateElement) (map[string]*IntegrationFilePair, []IntegrationMatchError) {
	filePairs := make(map[string]*IntegrationFilePair)
	matchErrors := []IntegrationMatchError{}

	for _, file := range integrationFiles {
		var integrationName string
		fileName := file.Config.FileName
		configRes := configFileRegex.FindAllStringSubmatch(fileName, -1)

		if len(configRes) > 0 {
			isValidFilePath, matchError := p.isValidIntegrationFilePath(file, CONFIG)

			if !isValidFilePath {
				matchErrors = append(matchErrors, matchError)
				continue
			}

			//set first capture group, all characters before '-' as integration name
			integrationName = configRes[0][1]
			val, ok := filePairs[integrationName]
			if ok {
				val.Configuration = file
				continue
			}
			filePairs[integrationName] = &IntegrationFilePair{
				Configuration: file,
			}
		}

		definitionRes := definitionFileRegex.FindAllStringSubmatch(fileName, -1)

		if len(definitionRes) > 0 {
			isValidFilePath, matchError := p.isValidIntegrationFilePath(file, DEFINITION)

			if !isValidFilePath {
				matchErrors = append(matchErrors, matchError)
				continue
			}

			//set first capture group, all characters before '-' as integration name
			integrationName = definitionRes[0][1]

			val, ok := filePairs[integrationName]
			if ok {
				val.Definition = file
				continue
			}
			filePairs[integrationName] = &IntegrationFilePair{
				Definition: file,
			}
		}
	}

	return filePairs, matchErrors
}

//Validates that ValidateElement filepath matches expected filepath for the provided IntegrationFileType
func (p InfraConfigIntegrationsMatch) isValidIntegrationFilePath(integrationFile config.ValidateElement, fileType IntegrationFileType) (bool, IntegrationMatchError) {
	var isValidPath bool
	matchError := IntegrationMatchError{}

	fileName := integrationFile.Config.FileName
	filePath := integrationFile.Config.FilePath

	var definitionFilepaths []string
	var configFilepath string

	if p.runtimeOS == "windows" {
		definitionFilepaths = definitionFilepathsWindows
		configFilepath = configurationFilepathWindows
	} else {
		definitionFilepaths = definitionFilepathsLinux
		configFilepath = configurationFilepathLinux
	}

	if fileType == CONFIG {
		isValidPath = (filePath == configFilepath)
	} else if fileType == DEFINITION {

		for _, path := range definitionFilepaths {
			isValidPath = (filePath == path)
			if isValidPath == true {
				break
			}
		}
	}

	if !isValidPath {
		matchError = IntegrationMatchError{
			IntegrationFile: integrationFile,
			Reason:          fmt.Sprintf("Filepath '%s' not a valid location for this Integration file '%s'", filePath, fileName),
		}
	}
	return isValidPath, matchError
}

//Valdates that a map of IntegrationFilePair actually have defined Configuration and Definition files.
//If IntegrationFilePair only has one file, it is removed from the map and captued as an IntegrationMatchError
//If IntegrationFilePair has two files, this will validate that their integration names as parsed from the yaml match
func validateIntegrationPairs(filePairs map[string]*IntegrationFilePair) (map[string]*IntegrationFilePair, []IntegrationMatchError) {
	matchErrors := []IntegrationMatchError{}

	for integration, pair := range filePairs {
		// if pair is missing configuration
		if pair.Configuration.Config == (config.ValidateElement{}.Config) {
			matchErrors = append(matchErrors, IntegrationMatchError{
				IntegrationFile: pair.Definition,
				Reason:          fmt.Sprintf("Definition file '%s%s' does not have matching Configuration file", pair.Definition.Config.FilePath, pair.Definition.Config.FileName),
			})
			delete(filePairs, integration)
			continue
			// if pair is missing definition
		} else if pair.Definition.Config == (config.ValidateElement{}.Config) {
			matchErrors = append(matchErrors, IntegrationMatchError{
				IntegrationFile: pair.Configuration,
				Reason:          fmt.Sprintf("Configuration file '%s%s' does not have matching Definition file", pair.Configuration.Config.FilePath, pair.Configuration.Config.FileName),
			})
			delete(filePairs, integration)
			continue
		}
		// we have two files, do their names match?
		isMatchedPair, nameMatchError := validateConfigPairNames(pair)
		if !isMatchedPair {
			matchErrors = append(matchErrors, nameMatchError...)
			delete(filePairs, integration)
		}
	}

	return filePairs, matchErrors
}

//Validates that IntegrationPair ValidateElements have matching integration names parsed from their yamls
func validateConfigPairNames(filePair *IntegrationFilePair) (bool, []IntegrationMatchError) {
	var (
		definitionName string
		configName     string
	)
	errorsEncountered := []IntegrationMatchError{}

	configKeys := filePair.Configuration.ParsedResult.FindKey("integration_name")

	if len(configKeys) > 0 {
		configName = configKeys[0].Value()
	} else {
		errorsEncountered = append(errorsEncountered, IntegrationMatchError{
			IntegrationFile: filePair.Configuration,
			Reason:          fmt.Sprintf("Integration Configuration File '%s%s' is missing key 'integration_name'", filePair.Configuration.Config.FilePath, filePair.Configuration.Config.FileName),
		})
	}
	defKeys := filePair.Definition.ParsedResult.FindKey("name")

	if len(defKeys) > 0 {
		definitionName = defKeys[0].Value()
	} else {
		errorsEncountered = append(errorsEncountered, IntegrationMatchError{
			IntegrationFile: filePair.Definition,
			Reason:          fmt.Sprintf("Integration Definition File '%s%s' is missing key 'name'", filePair.Definition.Config.FilePath, filePair.Definition.Config.FileName),
		})
	}
	// no need to check a match if one of these is empty
	if definitionName != "" && configName != "" {
		if definitionName != configName {
			errorsEncountered = append(errorsEncountered, IntegrationMatchError{
				IntegrationFile: filePair.Configuration,
				Reason: fmt.Sprintf("Integration Configuration file '%s%s' 'integration_name': '%s', expected to match Definition file '%s%s' 'name': '%s'",
					filePair.Configuration.Config.FilePath, filePair.Configuration.Config.FileName, configName,
					filePair.Definition.Config.FilePath, filePair.Definition.Config.FileName, definitionName),
			})
		}

	}

	if len(errorsEncountered) == 0 {
		return true, errorsEncountered
	}

	return false, errorsEncountered
}

//Builds string of all provided IntegrationMatchError reasons separated by /n
func buildMatchErrorSummary(matchErrors []IntegrationMatchError) string {
	var errorSummary string

	for _, matchError := range matchErrors {
		errorSummary = errorSummary + "\n" + matchError.Reason
	}

	return errorSummary
}
