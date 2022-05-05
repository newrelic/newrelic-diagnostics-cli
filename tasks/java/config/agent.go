package config

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// JavaConfigAgent - This struct defined the sample plugin which can be used as a starting point
type JavaConfigAgent struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
}

// JavaConfig - defines the payload returned by this task
type JavaConfig struct {
	config.ValidateElement
}

var javaKeys = []string{
	"enable_auto_app_naming",
	"enable_auto_transaction_naming",
	"max_stack_trace_lines",
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p JavaConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Config/Agent") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p JavaConfigAgent) Explain() string {
	return "Detect New Relic Java agent" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p JavaConfigAgent) Dependencies() []string {
	return []string{
		"Base/Config/Collect",
		"Base/Config/Validate", //This identifies this task as dependent on "Base/Config/Validate" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
	}
}

// Execute - The core work within each task
func (p JavaConfigAgent) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task

	if upstream["Base/Config/Validate"].HasPayload() {
		validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)
		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}
		javaValidation, checkValidationTrue := checkValidation(validations)

		if checkValidationTrue {
			log.Debug("Identified Java from validated config file, setting Java to true")
			return tasks.Result{
				Status:  tasks.Success,
				Summary: "Java agent identified as present on system",
				Payload: javaValidation,
			}
		}
	}
	// If checking with the parsed Config failed, now check the file itself line by line to detect java agent for invalid config files

	if upstream["Base/Config/Collect"].Status == tasks.Success {
		configs, ok := upstream["Base/Config/Collect"].Payload.([]config.ConfigElement)

		if !ok {
			return tasks.Result{
				Status:  tasks.Error,
				Summary: tasks.AssertionErrorSummary,
			}
		}
		javaConfig, checkConfigTrue := checkConfig(configs)

		if checkConfigTrue {
			log.Debug("Identified Java from config file parsing, setting Java to true")

			//Map config elements into ValidationElements so we always return a ValidationElement
			var validationResults []config.ValidateElement

			for _, configItem := range javaConfig {
				javaItem := config.ValidateElement{Config: configItem, Status: tasks.None} //This defines the mocked validate element we'll put in the results that is empty expect the config element
				validationResults = append(validationResults, javaItem)
			}
			return tasks.Result{
				Status:  tasks.Success,
				Summary: "Java agent identified as present on system",
				Payload: validationResults,
			}
		}
	}

	//Last check for the existence of the newrelic.jar as a last ditch effort
	if checkForJar() {
		log.Debug("Identified Java from Jar, setting Java to true")

		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Java agent identified as present on system because we found a New Relic JAR file",
			Payload: []config.ValidateElement{},
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: tasks.NoAgentDetectedSummary,
	}
}

// This uses the validation output since a valid yml should produce data that can be read by the FindString function to look for pertinent values
func checkValidation(validations []config.ValidateElement) ([]config.ValidateElement, bool) {

	var javaValidate []config.ValidateElement
	//Check the validated yml for some java attributes that don't exist in Ruby

	for _, validation := range validations {
		if filepath.Ext(validation.Config.FileName) != ".yml" {
			continue
		}

		var attributes []tasks.ValidateBlob
		for _, key := range javaKeys {
			attributes = append(attributes, validation.ParsedResult.FindKey(key)...)
		}

		if len(attributes) > 0 {
			log.Debug("found ", attributes, "in validated yml. Java language detected")
			javaValidate = append(javaValidate, validation)
		}
	}

	//Check for one or more ValidateElements
	if len(javaValidate) > 0 {
		log.Debug(len(javaValidate), " java ValidateElements found")
		return javaValidate, true
	}

	log.Debug("no java configuration elements found, setting false")
	return javaValidate, false
}

// This uses the config elements to manually check to see if a Java agent by stepping through the files
func checkConfig(configs []config.ConfigElement) ([]config.ConfigElement, bool) {
	var javaConfig []config.ConfigElement

	// compile keys into a single regex and then loop through each file
	regexString := "("
	for i, key := range javaKeys {
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

		if filepath.Ext(file) != ".yml" {
			continue
		}

		if tasks.FindStringInFile(regexString, file) {
			log.Debug("Found match in ", file)
			javaConfig = append(javaConfig, configItem)
		}

	}

	if len(javaConfig) > 0 {
		log.Debug("java key found in config file, setting java true")
		return javaConfig, true
	}

	log.Debug("no java configuration elements found, setting false")
	return javaConfig, false

}

// This check looks for the existence of the newrelic.jar in the file system as a final attempt at identifying this as a java app present
func checkForJar() bool {
	//check for existence of this specific file name in the user's system. We do not look for a regex/pattern name because it would be a never ending search
	if tasks.FileExists("newrelic.jar") {
		log.Debug("Jar file found, setting true")
		return true
	}

	// Now we can attempt to find a newrelic filename pattern, but only in the current directory
	jarRgx := regexp.MustCompile(`(?i)(newrelic)([\S]+)?\.jar`)
	dir, errDir := os.Getwd()

	if errDir != nil {
		log.Debug(errDir)
		return false
	}

	foundJar := false
	currentDirDepth := strings.Count(dir, string(filepath.Separator))
	maxDirDepth := currentDirDepth + 3

	// Walk current directory to max depth of 3 to look for jar file. Agent jar placement can be arbitrary
	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			log.Debug(err)
			return err
		}

		if info.IsDir() && strings.Count(path, string(filepath.Separator)) > maxDirDepth {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		if jarRgx.MatchString(info.Name()) {
			foundJar = true
			return errors.New("found jar")
		}

		return nil
	})

	log.Debug(walkErr)

	return foundJar

}
