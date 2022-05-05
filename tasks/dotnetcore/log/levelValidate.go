package log

import (
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetCoreLogLevelValidate - This struct defines this plugin
type DotNetCoreLogLevelValidate struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t DotNetCoreLogLevelValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Log/LevelValidate")
}

// Explain - Returns the help text for this task
func (t DotNetCoreLogLevelValidate) Explain() string {
	return "Validate New Relic .NET Core agent logging level"
}

// Dependencies - Returns the dependencies for this task.
func (t DotNetCoreLogLevelValidate) Dependencies() []string {
	return []string{
		"DotNetCore/Log/LevelCollect",
		"DotNetCore/Agent/Installed",
	}
}

// Execute - The core work within this task
func (t DotNetCoreLogLevelValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	logger.Debug("DotNetCore/Log/LevelValidate Start")

	if upstream["DotNetCore/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Core Agent not installed, skipping this task."
		return
	}

	if upstream["DotNetCore/Log/LevelCollect"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Log levels were not collected successfully, skipping this task."
		return
	}

	logLevels, ok := upstream["DotNetCore/Log/LevelCollect"].Payload.(map[string]string)
	if !ok || len(logLevels) == 0 {
		result.Status = tasks.None
		result.Summary = "Log levels were not collected, skipping this task."
		return
	}

	result = logLevelValidate(logLevels)
	return
}

func logLevelValidate(logLevels map[string]string) (result tasks.Result) {
	var possibleLevels = make(map[string]struct{}) // this should be faster than looping over a string slice
	possibleLevels["off"] = struct{}{}
	possibleLevels["emergency"] = struct{}{}
	possibleLevels["fatal"] = struct{}{}
	possibleLevels["alert"] = struct{}{}
	possibleLevels["critical"] = struct{}{}
	possibleLevels["severe"] = struct{}{}
	possibleLevels["error"] = struct{}{}
	possibleLevels["warn"] = struct{}{}
	possibleLevels["notice"] = struct{}{}
	possibleLevels["info"] = struct{}{}
	possibleLevels["debug"] = struct{}{}
	possibleLevels["fine"] = struct{}{}
	possibleLevels["trace"] = struct{}{}
	possibleLevels["finer"] = struct{}{}
	possibleLevels["verbose"] = struct{}{}
	possibleLevels["finest"] = struct{}{}
	possibleLevels["all"] = struct{}{}

	// initialize variables
	numErrors := 0
	numConfigs := 0
	validLevelsFound := make(map[string]string) // validLevelsFound[filepath]levelfound
	invalidLevelsFound := make(map[string]string)

	// loop through configs and see if there is a valid level set in each
	for configFullPath, levelFound := range logLevels {
		numConfigs++
		if _, ok := possibleLevels[levelFound]; ok { // this should be faster than looping through all the levels
			validLevelsFound[configFullPath] = levelFound
			logger.Debug("The log level is", levelFound, " in", configFullPath)
			continue
		}
		numErrors++
		invalidLevelsFound[configFullPath] = levelFound
		logger.Debug(`Found invalid log level "`, levelFound, `" in `, configFullPath)
	}

	if numErrors > 0 {
		result.URL = "https://docs.newrelic.com/docs/agents/net-agent/configuration/net-agent-configuration#log"
		result.Summary = "Invalid log level found in " + strconv.Itoa(numErrors) + " of " + strconv.Itoa(numConfigs) + " newrelic.config files:\n"
		for config, level := range invalidLevelsFound {
			result.Summary += " - " + config + `: "` + level + `"` + "\n"
		}
		if numConfigs == numErrors {
			// all log levels in all files were invalid
			result.Status = tasks.Failure
		}
		if numConfigs > numErrors {
			// some log levels were invalid, some were valid
			result.Status = tasks.Warning
			result.Payload = validLevelsFound // put the valid level in the payload
		}

		return
	}

	// if it gets here, numErrors was 0
	result.Status = tasks.Success
	result.Summary = "All log levels valid in all " + strconv.Itoa(numConfigs) + " newrelic.config files:"
	for config, level := range validLevelsFound {
		result.Summary += " - " + config + `: "` + level + `"` + "\n"
	}
	result.Payload = validLevelsFound

	return
}
