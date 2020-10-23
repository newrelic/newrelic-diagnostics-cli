package config

import (
	"fmt"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var appNameConfigKeys = []string{
	"app_name",         // Java, Node, Python, Ruby
	"newrelic.appname", // PHP
	"AppName",          // GoLang
	"NewRelic.AppName", // .Net
	"name",             // .Net
}

var defaultAppNames = []string{
	"PHP Application",                  // PHP
	"Python Application",               // Python
	"Python Application (Development)", // Python
	"Python Application (Staging)",     // Python
	"My Application",                   // Ruby, .Net, Node, Java
	"My Application (Development)",     // Ruby, .Net, Node, Java
	"My Application (Test)",            // Ruby, .Net, Node, Java
	"My Application (Staging)",         // Ruby, .Net, Node, Java
}

// BaseConfigAppName - Struct for task definition
type BaseConfigAppName struct {
}

// AppNameInfo - Struct to store relevant AppName info
type AppNameInfo struct {
	Name     string
	FilePath string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseConfigAppName) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/AppName")
}

// Explain - Returns the help text for each individual task
func (t BaseConfigAppName) Explain() string {
	return "Check for default application names in New Relic agent configuration."
}

// Dependencies - Returns the dependencies for each task.
func (t BaseConfigAppName) Dependencies() []string {
	return []string{
		"Base/Config/Validate",
	}
}

// Execute - The core work within each task
func (t BaseConfigAppName) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	// check to see if upstream was successful, exit if not
	if upstream["Base/Config/Validate"].Status != tasks.Success && (upstream["Base/Config/Validate"].Status != tasks.Warning) {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: no validated config files to check",
		}
	}

	configElements, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}
	
	defaultNameMatches := ""

	appNameInfoFromConfig := getAppNamesFromConfig(configElements, appNameConfigKeys)
	if len(appNameInfoFromConfig) == 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "No New Relic app names were found. Please ensure an app name is set in your New Relic agent configuration.",
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
		}
	}

	for _, nameInfo := range appNameInfoFromConfig {
		for _, defaultName := range defaultAppNames {

			if nameInfo.Name == defaultName {
				defaultNameMatches += fmt.Sprintf("\n\t\"%s\" as specified in %s", nameInfo.Name, nameInfo.FilePath)
			}

		}
	}

	var defaultWarning = "\nMultiple applications with the same default appname will all report to the same source. " +
		"You may want to consider changing to a unique appname. Note that this will cause the application to report to " +
		"a new heading in the New Relic user interface, with a total discontinuity of data. If you are overriding the " +
		"default appname with environment variables, you can ignore this warning.\n--"
	if len(defaultNameMatches) > 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("One or more of your applications is using a default appname: %s %s", defaultNameMatches, defaultWarning),
			URL:     "https://docs.newrelic.com/docs/agents/manage-apm-agents/app-naming/name-your-application",
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: fmt.Sprintf("%s unique application name(s) found.", strconv.Itoa(len(appNameInfoFromConfig))),
		Payload: appNameInfoFromConfig,
	}
}

func getAppNamesFromConfig(configElements []ValidateElement, configNames []string) []AppNameInfo {
	result := []AppNameInfo{}

	for _, nameKey := range configNames {
		for _, configFile := range configElements {

			foundKeys := configFile.ParsedResult.FindKey(nameKey)
			configFilePath := configFile.Config.FilePath
			configFileName := configFile.Config.FileName

			for _, key := range foundKeys {

				if !key.IsLeaf() {
					for _, child := range key.Children {
						appName := child.Value()
						result = append(result, AppNameInfo{
							Name:     appName, // should we sanitize this?
							FilePath: fmt.Sprintf("%s%s", configFilePath, configFileName),
						})
					}
				} else {
					appName := key.Value()

					if len(appName) > 0 {
						result = append(result, AppNameInfo{
							Name:     appName, // should we sanitize this?
							FilePath: fmt.Sprintf("%s%s", configFilePath, configFileName),
						})
					}
				}
			}
		}
	}
	return result
}
