package config

import (
	"path/filepath"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// NodeConfigAgent - This struct defined the sample plugin which can be used as a starting point
type NodeConfigAgent struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

var nodeKeys = []string{
	"logging.filepath",
	"app_name",
	"license_key",
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Config/Agent") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p NodeConfigAgent) Explain() string {
	return "Detect New Relic Nodejs agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p NodeConfigAgent) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p NodeConfigAgent) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	if upstream["Base/Config/Validate"].HasPayload() {
		validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}

		nodeValidation, checkValidationTrue := checkValidation(validations)

		if checkValidationTrue {
			log.Debug("Identified Node from validated config file, setting Node to true")
			result.Status = tasks.Success
			result.Summary = "Node agent identified as present on system"
			result.Payload = nodeValidation
			return result
		}
	}

	//If this fails to identify the language, now check the raw file itself

	if upstream["Base/Config/Collect"].Status == tasks.Success {
		configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}

		nodeConfig, checkConfigTrue := checkConfig(configs)

		if checkConfigTrue {
			log.Debug("Identified Node from config file parsing, setting Node to true")
			result.Status = tasks.Success
			result.Summary = "Node agent identified as present on system"
			//Map config elements into ValidationElements so we always return a ValidationElement
			var validationResults []config.ValidateElement

			for _, configItem := range nodeConfig {
				nodeItem := config.ValidateElement{Config: configItem, Status: tasks.None} //This defines the mocked validate element we'll put in the results that is empty expect the config element
				validationResults = append(validationResults, nodeItem)
			}
			result.Payload = validationResults
			return result
		}
	}

	log.Debug("No Node agent found on system")
	result.Status = tasks.None
	result.Summary = "No Node agent found on system"
	return result

}

// This uses the validation output since a valid yml should produce data that can be read by the FindString function to look for pertinent values
func checkValidation(validations []config.ValidateElement) ([]config.ValidateElement, bool) {
	var nodeValidate []config.ValidateElement

	//Check the validated yml for some node attributes that don't exist in Ruby

	for _, validation := range validations {
		if filepath.Ext(validation.Config.FileName) != ".js" {
			continue
		}

		for _, key := range nodeKeys {
			attributes := validation.ParsedResult.FindKey(key)
			if len(attributes) > 0 {
				log.Debug("found ", attributes, "in validated yml. Node language detected")
				nodeValidate = append(nodeValidate, validation)
				break
			}
		}
	}

	//Check for one or more ValidateElements
	if len(nodeValidate) > 0 {
		log.Debug(len(nodeValidate), " node ValidateElements found")
		return nodeValidate, true
	}

	log.Debug("no node configuration elements found, setting false")
	return nodeValidate, false
}

func checkConfig(configs []config.ConfigElement) ([]config.ConfigElement, bool) {
	var nodeConfig []config.ConfigElement

	// compile keys into a single regex and then loop through each file
	regexString := "("
	for i, key := range nodeKeys {
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
		if filepath.Ext(file) != ".js" {
			continue
		}

		if tasks.FindStringInFile(regexString, file) {
			log.Debug("Found match in ", file)
			nodeConfig = append(nodeConfig, configItem)
		}

	}

	if len(nodeConfig) > 0 {
		log.Debug("node key found in config file, setting node true")
		return nodeConfig, true
	}

	log.Debug("no node configuration elements found, setting false")
	return nodeConfig, false

}
