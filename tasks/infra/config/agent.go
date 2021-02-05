package config

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// InfraConfigAgent - This struct defined the sample plugin which can be used as a starting point
type InfraConfigAgent struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name              string
	upstream          map[string]tasks.Result
	validationChecker validationFunc
	configChecker     configFunc
	binaryChecker     binaryFunc
}

type validationFunc func([]config.ValidateElement) ([]config.ValidateElement, bool)
type configFunc func([]config.ConfigElement) ([]config.ConfigElement, bool)
type binaryFunc func() (bool, string)

// InfraConfig - defines the payload returned by this task
type InfraConfig struct {
	config.ValidateElement
}

var infraKeys = []string{
	"New Relic Infrastructure configuration file",
	"custom_attributes",
	"license_key",
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/Agent") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p InfraConfigAgent) Explain() string {
	return "Detect New Relic Infrastructure agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p InfraConfigAgent) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p InfraConfigAgent) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task

	if upstream["Base/Config/Validate"].HasPayload(){
		validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
		if !ok {
			return tasks.Result{
				Status: tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}
		infraValidation, checkValidationTrue := p.validationChecker(validations)

		if checkValidationTrue {
			log.Debug("Identified Infra from validated config file, setting Infra to true")
			return tasks.Result{
				Status: tasks.Success,
				Summary: "Infra agent identified as present on system from validated config file",
				Payload: infraValidation,
			}
		}
	}

	if upstream["Base/Config/Collect"].Status == tasks.Success {
		configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement)

		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}

		infraConfig, checkConfigTrue := p.configChecker(configs)

		if checkConfigTrue {
			log.Debug("Identified Infra from config file parsing, setting Infra to true")
			//Map config elements into ValidationElements so we always return a ValidationElement
			var validationResults []config.ValidateElement

			for _, configItem := range infraConfig {
				//This defines the mocked validate element we'll put in the results that is empty except the config element
				infraItem := config.ValidateElement{Config: configItem, Status: tasks.None}
				validationResults = append(validationResults, infraItem)
			}
			return tasks.Result{
				Status: tasks.Success,
				Summary: "Infra agent identified as present on system from parsed config file",
				Payload: validationResults,
			}
		}

	}

	//Last check for the existence of the newrelic-infra binary as a last ditch effort
	binaryFound, binaryFilename := p.binaryChecker()
	if binaryFound {
		log.Debug("Identified Infra from binary, setting Infra to true")
		return tasks.Result{
			Status: tasks.Success,
			Summary: "Infra agent identified as present on system from existence of binary file: " + binaryFilename,
		}
	}
	log.Debug("No Infra agent found on system")
	return tasks.Result{
		Status: tasks.None,
		Summary: "No Infra agent found on system",
	}
}

// This uses the validation output since a valid yml should produce data that can be read by the FindString function to look for pertinent values
func checkValidation(validations []config.ValidateElement) ([]config.ValidateElement, bool) {

	var infraValidate []config.ValidateElement

	for _, validation := range validations {
		for _, key := range infraKeys {
			if validation.Config.FileName != "newrelic-infra.yml" {
				continue
			}

			log.Debug("parsed result is ", validation.ParsedResult)
			attributes := validation.ParsedResult.FindKey(key)
			if len(attributes) > 0 {
				log.Debug("found ", attributes, "in validated yml. Infra agent detected")
				infraValidate = append(infraValidate, validation)
				break
			}
		}
	}

	//Check for one or more ValidateElements
	if len(infraValidate) > 0 {
		log.Debug(len(infraValidate), " infra ValidateElements found")
		return infraValidate, true
	}

	log.Debug("no infra configuration elements found, setting false")
	return infraValidate, false
}

// This uses the config elements to manually check to see if a Infra agent by stepping through the files
func checkConfig(configs []config.ConfigElement) ([]config.ConfigElement, bool) {
	var infraConfig []config.ConfigElement

	// compile keys into a single regex and then loop through each file
	regexString := "("
	for i, key := range infraKeys {
		if i == 0 {
			regexString += key //This doesn't add a | for the first string
		} else {
			regexString += "|" + key
		}
	}
	regexString += ")"
	log.Debug("regex to search is ", regexString)

	for _, configItem := range configs {
		if configItem.FileName != "newrelic-infra.yml" {
			continue
		}

		file := configItem.FilePath + configItem.FileName

		if tasks.FindStringInFile(regexString, file) {
			log.Debug("Found match in ", file)
			infraConfig = append(infraConfig, configItem)
		}

	}

	if len(infraConfig) > 0 {
		log.Debug("infra key found in config file, setting infra true")
		return infraConfig, true
	}

	log.Debug("no infra configuration elements found, setting false")
	return infraConfig, false

}

func checkForBinary() (bool, string) {

	binaryNames := []string{
		"newrelic-infra",
		"newrelic-infra.exe",
	}

	for _, binaryName := range binaryNames {

		if tasks.FileExists(binaryName) {
			log.Debug("Binary file found, setting true")
			return true, binaryName
		}
	}
	log.Debug("Done search for binary files, setting false")
	return false, ""
}
