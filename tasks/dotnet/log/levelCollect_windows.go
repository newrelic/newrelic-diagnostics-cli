package log

import (
	"strings"

	"github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

// DotNetLogLevelCollect - This struct defines this plugin
type DotNetLogLevelCollect struct {
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (t DotNetLogLevelCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Log/LevelCollect")
}

// Explain - Returns the help text for this task
func (t DotNetLogLevelCollect) Explain() string {
	return "Determine New Relic .NET Core agent logging level"
}

// Dependencies - Returns the dependencies for this task.
func (t DotNetLogLevelCollect) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
		"DotNet/Config/Agent",
	}
}

// Execute - The core work within this task
func (t DotNetLogLevelCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	logger.Debug("DotNet/Log/LevelCollect Start")

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Framework Agent not installed, skipping this task."
		return
	}

	if upstream["DotNet/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Framework Agent newrelic.config files not present, skipping this task."
		return
	}

	configFiles, ok := upstream["DotNet/Config/Agent"].Payload.([]config.ValidateElement)
	if !ok || len(configFiles) == 0 {
		result.Status = tasks.None
		result.Summary = ".NET Framework Agent newrelic.config files not present, skipping this task."
		return
	}

	result = logLevelCollect(configFiles)

	return
}

func logLevelCollect(configs []config.ValidateElement) (result tasks.Result) {
	configsChecked := make(map[string]struct{})
	levelsFound := make(map[string]string) // levelsFound[filepath]levelfound
	numConfigs := 0

	// loop through configs and gather the log levels in a slice
	for _, config := range configs {
		configFullPath := config.Config.FilePath + config.Config.FileName
		_, alreadyChecked := configsChecked[configFullPath] // sometimes the same config file is gathered twice (unsure why)... This will ensure we only check and report on it once.
		if !alreadyChecked && strings.EqualFold(config.Config.FileName, "newrelic.config") {
			numConfigs++
			configsChecked[configFullPath] = struct{}{}
			levelFound := config.ParsedResult.FindKeyByPath("/configuration/log/-level").Value()
			levelsFound[configFullPath] = levelFound
		}
	}

	if numConfigs == 0 || len(levelsFound) == 0 {
		result.Status = tasks.None
		result.Summary = "No newrelic.config files found."
		return
	}

	// build a basic slice for strings.Join
	levelsInfo := []string{}
	for _, lvl := range levelsFound {
		levelsInfo = append(levelsInfo, lvl)
	}

	result.Status = tasks.Info
	result.Summary = "Levels found: " + strings.Join(levelsInfo, ", ")
	result.Payload = levelsFound
	return
}
