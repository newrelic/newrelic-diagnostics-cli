package log

import (
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetLogLevelValidate - This struct defines this plugin
type DotNetLogLevelValidate struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t DotNetLogLevelValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Log/LevelValidate")
}

// Explain - Returns the help text for this task
func (t DotNetLogLevelValidate) Explain() string {
	return "Validate New Relic .NET Core agent logging level"
}

// Dependencies - Returns the dependencies for this task.
func (t DotNetLogLevelValidate) Dependencies() []string {
	return []string{
		"DotNet/Log/LevelCollect",
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within this task
func (t DotNetLogLevelValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	logger.Debug("DotNet/Log/LevelValidate Start")

	// abort if it isn't installed
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}

	if upstream["DotNet/Log/LevelCollect"].Status != tasks.Info {
		result.Status = tasks.None
		result.Summary = "Log levels were not collected successfully, skipping this task."
		return
	}

	logLevels, ok := upstream["DotNet/Log/LevelCollect"].Payload.(map[string]string)
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
