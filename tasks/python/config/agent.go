package config

import (
	"path/filepath"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

// PythonConfigAgent - This struct defines the Python agent config is present.
type PythonConfigAgent struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

var pythonKeys = []string{
	"transaction_tracer.function_trace",
	"thread_profiler.enabled",
	"monitor_mode",
}

// Identifier - This returns the Category, Subcategory and Name of this task.
func (p PythonConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Python/Config/Agent") // This should be updated to match the struct name
}

// Explain - Returns the help text for the PythonConfigAgent task.
func (p PythonConfigAgent) Explain() string {
	return "Detect New Relic Python agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for this task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p PythonConfigAgent) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within this task.
func (p PythonConfigAgent) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if ok {
		log.Debug("Base/Config/Validate payload correct type")
		//		log.Debug(configs) //This may be useful when debugging to log the entire results to the screen
	}

	pythonValidation, checkValidationTrue := checkValidation(validations)

	if checkValidationTrue {
		log.Debug("Identified Python from validated config file, setting Python to true")
		result.Status = tasks.Success
		result.Summary = "Python agent identified as present on system"
		result.Payload = pythonValidation
		return result
	}
	//If this fails to identify the language, now check the raw file itself

	configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if ok {
		log.Debug("Base/Config/Collect payload correct type")
		//		log.Debug(configs) //This may be useful when debugging to log the entire results to the screen
	}

	pythonConfig, checkConfigTrue := checkConfig(configs)

	if checkConfigTrue {
		log.Debug("Identified Python from config file parsing, setting Python to true")
		result.Status = tasks.Success
		result.Summary = "Python agent identified as present on system"
		//Map config elements into ValidationElements so we always return a ValidationElement
		var validationResults []config.ValidateElement

		for _, configItem := range pythonConfig {
			pythonItem := config.ValidateElement{Config: configItem, Status: tasks.None} //This defines the mocked validate element we'll put in the results that is empty expect the config element
			validationResults = append(validationResults, pythonItem)
		}
		return result
	}

	log.Debug("No Python agent found on system")
	result.Status = tasks.None
	result.Summary = "No Python agent found on system"
	return result

}

// This uses the validation output since a valid yml should produce data that can be read by the FindString function to look for pertinent values
func checkValidation(validations []config.ValidateElement) ([]config.ValidateElement, bool) {
	var pythonValidate []config.ValidateElement

	//Check the validated yml for some python attributes that don't exist in Ruby

	for _, key := range pythonKeys {
		for _, validation := range validations {
			if filepath.Ext(validation.Config.FileName) != ".ini" {
				continue
			}

			attributes := validation.ParsedResult.FindKey(key)
			if len(attributes) > 0 {
				log.Debug("found ", attributes, "in validated config. Python language detected")
				pythonValidate = append(pythonValidate, validation)
			}
		}
	}

	//Check for one or more ValidateElements
	if len(pythonValidate) > 0 {
		log.Debug(len(pythonValidate), " python ValidateElements found")
		return pythonValidate, true
	}

	log.Debug("no python configuration elements found, setting false")
	return pythonValidate, false
}

func checkConfig(configs []config.ConfigElement) ([]config.ConfigElement, bool) {
	var pythonConfig []config.ConfigElement

	// compile keys into a single regex and then loop through each file
	regexString := "("
	for i, key := range pythonKeys {
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
		if filepath.Ext(file) != ".ini" {
			continue
		}

		if tasks.FindStringInFile(regexString, file) {
			log.Debug("Found match in ", file)
			pythonConfig = append(pythonConfig, configItem)
		}

	}

	if len(pythonConfig) > 0 {
		log.Debug("python key found in config file, setting python true")
		return pythonConfig, true
	}

	log.Debug("no python configuration elements found, setting false")
	return pythonConfig, false

}
