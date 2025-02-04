package config

import (
	"errors"
	"os"
	"regexp"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

// app.config/web.config files can contain custom paths to New Relic .NET Agent config files, that we need to collect in this task.
// We use these regexs to tease out these files for early parsing (i.e. in this task, before Base/Config/Validate)
// so we can get the specified .NET agent config paths, if they're set.
// See: https://docs.newrelic.com/docs/agents/net-agent/configuration/net-agent-configuration#app-cfg-location
var filesWithNETAgentConfigPaths = []*regexp.Regexp{
	regexp.MustCompile("^(?i)(web|app)[.]config$"),
	regexp.MustCompile("(?i).+[.]exe[.]config$"),
}

// isFileWithAgentConfigPath - determine if file may contain custom path to .NET Agent config file
func isFileWithNETAgentConfigPath(filename string) bool {
	for _, regex := range filesWithNETAgentConfigPaths {
		if regex.MatchString(filename) {
			return true
		}
	}
	return false
}

// getAgentConfigPathFromFile - Parse app.config/web.config files for custom path to
// New Relic .NET Agent config file, if it is set.
// Example: <add key = "NewRelic.ConfigFile" value="C:\Path-to-alternate-config-dir\newrelic.config" />
func getNETAgentConfigPathFromFile(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Debugf("Unable to open '%s': %s\n", filepath, err.Error())
		return "", err
	}
	defer file.Close()

	parsedFile, err := parseXML(file)
	if err != nil {
		log.Debugf("Unable to parse '%s': %s\n", filepath, err.Error())
		return "", err
	}

	var agentConfigPath string

	// Tease out <add /> objects from XML:
	// <add key = "NewRelic.ConfigFile" value="C:\Path-to-alternate-config-dir\newrelic.config" />
	addObjects := parsedFile.FindKey("add")
	for _, object := range addObjects {
		objectProperties := map[string]string{}

		for _, property := range object.Children {
			objectProperties[property.Key] = property.Value()
		}

		if objectProperties["-key"] == "NewRelic.ConfigFile" {
			agentConfigPath = objectProperties["-value"]
			break
		}
	}

	if agentConfigPath == "" {
		return "", nil
	}

	// Check if file exists
	pathInfo, err := os.Stat(agentConfigPath)
	if err != nil {
		return "", err
	}

	// .NET docs state this needs to be an absolute filepath, so we'll honor that instead
	// of implementing additional logic for handling a directory
	if pathInfo.IsDir() {
		return "", errors.New("expected absolute filepath")
	}

	return agentConfigPath, nil
}
