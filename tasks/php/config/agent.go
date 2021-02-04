package config

import (
	"path/filepath"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// PHPConfigAgent - This struct defined the sample plugin which can be used as a starting point
type PHPConfigAgent struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

var phpKeys = []string{
	"newrelic.daemon.utilization.detect_docker",
	"newrelic.enabled",
	"newrelic.license",
	"utilization.detect_docker",
	"utilization.detect_aws",
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p PHPConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("PHP/Config/Agent") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p PHPConfigAgent) Explain() string {
	return "Detect New Relic PHP agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p PHPConfigAgent) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p PHPConfigAgent) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if !ok {
		return tasks.Result{
			Status: tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	phpValidation, checkValidationTrue := checkValidation(validations)

	if checkValidationTrue {
		log.Debug("Identified PHP from validated config file, setting PHP to true")
		result.Status = tasks.Success
		result.Summary = "PHP agent identified as present on system"
		result.Payload = phpValidation
		return result
	}
	//If this fails to identify the language, now check the raw file itself

	configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if !ok {
		return tasks.Result{
			Status: tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	phpConfig, checkConfigTrue := checkConfig(configs)

	if checkConfigTrue {
		log.Debug("Identified PHP from config file parsing, setting PHP to true")
		result.Status = tasks.Success
		result.Summary = "PHP agent identified as present on system"
		//Map config elements into ValidationElements so we always return a ValidationElement
		var validationResults []config.ValidateElement

		for _, configItem := range phpConfig {
			phpItem := config.ValidateElement{Config: configItem, Status: tasks.None} //This defines the mocked validate element we'll put in the results that is empty expect the config element
			validationResults = append(validationResults, phpItem)
		}
		return result
	}

	log.Debug("No PHP agent found on system")
	result.Status = tasks.None
	result.Summary = "No PHP agent found on system"
	return result

}

// This uses the validation output since a valid yml should produce data that can be read by the FindString function to look for pertinent values
func checkValidation(validations []config.ValidateElement) ([]config.ValidateElement, bool) {
	phpValidate := []config.ValidateElement{}

	//Check the validated yml for some php attributes that don't exist in Ruby
	for _, validation := range validations {
		if filepath.Ext(validation.Config.FileName) != ".ini" && filepath.Ext(validation.Config.FileName) != ".cfg" {
			continue
		}
		for _, key := range phpKeys {
			attributes := validation.ParsedResult.FindKey(key)

			if len(attributes) > 0 {
				log.Debug("found ", attributes, "in validated yml. PHP language detected")
				phpValidate = append(phpValidate, validation)
				break
			}
		}
	}
	//Check for one or more ValidateElements
	if len(phpValidate) > 0 {
		log.Debug(len(phpValidate), " php ValidateElements found")
		return phpValidate, true
	}

	log.Debug("no php configuration elements found, setting false")
	return phpValidate, false
}

func checkConfig(configs []config.ConfigElement) ([]config.ConfigElement, bool) {
	var phpConfig []config.ConfigElement

	// compile keys into a single regex and then loop through each file
	regexString := "("
	for i, key := range phpKeys {
		if i == 0 {
			regexString += key //This doesn't add a | for the first string
		} else {
			regexString += "|" + key
		}
	}
	regexString += ")"
	log.Debug("regex to search is ", regexString)

	for _, configItem := range configs {
		file := configItem.FilePath + configItem.FileName
		if filepath.Ext(file) != ".ini" && filepath.Ext(file) != ".cfg" {
			continue
		}

		if tasks.FindStringInFile(regexString, file) {
			log.Debug("Found match in ", file)
			phpConfig = append(phpConfig, configItem)
		}

	}

	if len(phpConfig) > 0 {
		log.Debug("php key found in config file, setting php true")
		return phpConfig, true
	}

	log.Debug("no php configuration elements found, setting false")
	return phpConfig, false

}
