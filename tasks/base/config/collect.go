package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var pathsToIgnore = []string{"node_modules"}

var configSysProp = "-Dnewrelic.config.file"

var configEnvVarKeys = []string{
	"NEW_RELIC_HOME",        // Node, Java
	"NEW_RELIC_CONFIG_FILE", // Python
	"NEW_RELIC_CONFIG_PATH", // Ruby
	"NRIA_CONFIG_FILE",      // Infra
	"NEWRELIC_INSTALL_PATH", // .NET, .NET Core (Windows)
	"CORECLR_NEWRELIC_HOME", // .NET Core
	// PHP agent does not support config env vars
}

var noConfigEnvVar = "NEW_RELIC_NO_CONFIG_FILE"

var patterns = []string{
	"newrelic[.]yml",
	"newrelic[.]xml",
	"^(?i)newrelic[.]config$",
	"newrelic[.]js",
	"newrelic[.]cfg",
	"newrelic[.]ini",
	"Podfile",
	// "Podfile[.]lock",
	"proguard-rules[.]pro",
	"proguard[.]multidex[.]config",
	"dexguard-release[.]pro",
	"newrelic[.]properties",
	"NewRelicConfig[.]java",
	"gradle-wrapper[.]properties",
	"private-location-settings[.]json",
	"newrelic-infra[.]yml",
	"NewRelic[.]h",
}

//This list will prompt the end user asking for permission to include each file
var secureFilePatterns = []string{
	"AppDelegate[.]m",
	"AppDelegate[.]swift",
	"AndroidManifest[.]xml",
	"gradle[.]properties",
	"build[.]gradle",
	"project[.]pbxproj",
	"^(?i)(web|app)[.]config$",
	"(?i).+[.]exe[.]config$", //  app.config files are almost always app-me.exe.config. filter NewRelicStatusMonitor.exe.config later
	"(.+).csproj$",           //project file use to configure NET app
	"(?i)appSettings[.]json$",
}

var warningSummaryFmt = "The " + tasks.ThisProgramFullName + " cannot collect New Relic config files from the provided path (%s):\n%s\nIf you are working with a support ticket, manually provide your New Relic config file for further troubleshooting\n"

// BaseConfigCollect - Primary task to search for and find config file. Will optionally take command line input as source
type BaseConfigCollect struct {
}

// ConfigElement - holds a reference to the config file name and location
type ConfigElement struct {
	FileName string
	FilePath string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseConfigCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/Collect")
}

// Explain - Returns the help text for each individual task
func (p BaseConfigCollect) Explain() string {
	return "Collect New Relic configuration files"
}

// Dependencies - No dependencies since this is generally one of the first tasks to run
func (p BaseConfigCollect) Dependencies() []string {
	// no dependencies!
	return []string{
		"Base/Env/CollectEnvVars",
		"Base/Env/CollectSysProps",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseConfigCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Could not get envVars from upstream")
	}
	//Get config file from filepath passed through command line argument:
	if options.Options["configFile"] != "" {
		log.Debug("Config file specified on command line " + options.Options["configFile"])
		configOverride, err := filepath.Abs(string(options.Options["configFile"]))
		if err != nil {
			log.Debug("Error reading config file path")
		}

		fileStatuses := tasks.ValidatePaths([]string{configOverride})
		//no need to iterate because configOverride is a single string/file
		if !(fileStatuses[0].IsValid) {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: fmt.Sprintf("The path provided to the config file is not valid:%s", (fileStatuses[0].ErrorMsg.Error())),
			}
		}

		envelope := tasks.FileCopyEnvelope{Path: configOverride, Identifier: p.Identifier().String()}

		return tasks.Result{
			Status:      tasks.Success,
			Summary:     "1 file found",
			Payload:     []ConfigElement{ConfigElement{filepath.Base(configOverride), filepath.Dir(configOverride) + "/"}},
			FilesToCopy: []tasks.FileCopyEnvelope{envelope},
		}
	}

	// Search for config file in standard/default expected locations

	var paths []string

	localPath, err := os.Getwd()

	if err != nil {
		log.Debug("Error reading local working directory")
	}

	paths = append(paths, localPath)

	if runtime.GOOS == "windows" {
		sysProgramFiles := envVars["ProgramFiles"]
		sysProgramData := envVars["ProgramData"]
		paths = append(paths, sysProgramFiles+`\New Relic`)
		paths = append(paths, sysProgramData+`\New Relic\`)
	} else {
		paths = append(paths, "/etc/")
		paths = append(paths, "/opt/newrelic/synthetics/.newrelic/synthetics/minion/")
		paths = append(paths, "/usr/local/newrelic-netcore20-agent/")
		paths = append(paths, "/usr/local/newrelic-dotnet-agent/") // https://github.com/newrelic/newrelic-diagnostics-cli/issues/114
	}

	//Find insecure paths
	foundConfigs := tasks.FindFiles(patterns, paths)

	// These are files to skip for the secure files prompt
	var skippedSecureConfigs = make(map[string]struct{})
	skippedSecureConfigs["NewRelic.ServerMonitor.Config.exe.config"] = struct{}{}
	skippedSecureConfigs["NewRelic.ServerMonitor.exe.config"] = struct{}{}
	skippedSecureConfigs["NewRelicStatusMonitor.exe.config"] = struct{}{}

	//Find insecure paths
	foundSecureConfigs := tasks.FindFiles(secureFilePatterns, paths)

	var invalidConfigFiles, cannotCollectConfigFiles []string //will represent the secure files that the user reject nrdiag to collect at the prompt
	var warningSummaryOnInvalidFiles string

	for _, secureConfig := range foundSecureConfigs {

		filename := filepath.Base(secureConfig)

		// Check web.config & app.config for custom .NET agent paths
		if isFileWithNETAgentConfigPath(secureConfig) {
			configPath, err := getNETAgentConfigPathFromFile(secureConfig)
			if err != nil {
				invalidConfigFiles = append(invalidConfigFiles, configPath)
				warningSummaryOnInvalidFiles += fmt.Sprintf(warningSummaryFmt, configPath, err.Error())
			} else if configPath != "" {
				foundConfigs = append(foundConfigs, configPath)
			}
		}

		// This checks to see if our file is a file we should be skipping as secure
		if _, ok := skippedSecureConfigs[filename]; !ok {
			question := fmt.Sprintf("We've found a file that may contain secure information: %s\n", secureConfig) +
				"Include this file in nrdiag-output.zip?"
			if tasks.PromptUser(question, options) {
				if !config.Flags.Quiet {
					log.Info("Adding file to Diagnostics CLI zip file: ", secureConfig)
				}
				foundConfigs = append(foundConfigs, secureConfig)
			} else {
				cannotCollectConfigFiles = append(cannotCollectConfigFiles, secureConfig)
			}
		} else {
			foundConfigs = append(foundConfigs, secureConfig)
		}
	}
	warningSummaryCannotCollect := ""
	if len(cannotCollectConfigFiles) > 0 {
		warningSummaryCannotCollect += "\nThe following files were not collected because the user opted out from including them in the nrdiag-output.zip: " + strings.Join(cannotCollectConfigFiles, ", ")
	}

	//search for config file in New Relic System Property
	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		processes, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			for _, process := range processes {
				configPath, isPresent := process.SysPropsKeyToVal[configSysProp]
				if isPresent {
					//Example path: -Dnewrelic.config.file=/usr/local/newrelic/newrelic.yml
					invalidConfigFiles, foundConfigs = appendToInvalidOrFoundConfigs(configPath, &warningSummaryOnInvalidFiles, invalidConfigFiles, foundConfigs) //add to list of foundConfigs or to list of invalidConfigFiles depending if the file can be collected
				}
			}
		}
	}

	// Search for config file in New Relic environment variables
	for _, envVarKey := range configEnvVarKeys {
		configPath, envVarKeyIsPresent := envVars[envVarKey]
		if !envVarKeyIsPresent {
			continue
		}
		invalidConfigFiles, foundConfigs = appendToInvalidOrFoundConfigs(configPath, &warningSummaryOnInvalidFiles, invalidConfigFiles, foundConfigs)
	}

	if len(foundConfigs) == 0 {

		if len(invalidConfigFiles) > 0 {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: warningSummaryOnInvalidFiles + warningSummaryCannotCollect,
			}
		}
		noConfigFileVal, envVarIsPresent := envVars[noConfigEnvVar]
		if envVarIsPresent {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: tasks.ThisProgramFullName + " was unable to collect a New Relic config file because the " + noConfigEnvVar + " env var was set to " + noConfigFileVal + "." + warningSummaryCannotCollect,
			}
		}
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "New Relic configuration files not found where the " + tasks.ThisProgramFullName + " was executed. Please ensure the " + tasks.ThisProgramFullName + " executable is within your application's directory alongside your New Relic agent configuration file(s). If you cannot set New Relic configuration files in your application's directory, move the " + tasks.ThisProgramFullName + " to that directory or use the -c <file_path> to specify the New Relic configuration file location." + warningSummaryCannotCollect,
		}
	}

	filesToCopy := []tasks.FileCopyEnvelope{}
	configFilesInfo := []ConfigElement{}

	for _, configFile := range foundConfigs {
		dir, fileName := filepath.Split(configFile)
		//because foundConfigs is using the external helper function tasks.FindFiles which goes recursively into subdirectories, is possible that at this point we have already appended paths we do not care about. Let's not copy/collect those files:
		if isConfigFileinPathToIgnore(dir) {
			continue
		}

		c := ConfigElement{fileName, dir}
		configFilesInfo = append(configFilesInfo, c)
		filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: configFile})
	}

	var finalSummary string
	finalSummary = fmt.Sprintf("There were %d config file(s) found", len(configFilesInfo))

	if len(invalidConfigFiles) > 0 {
		finalSummary += "\n" + warningSummaryOnInvalidFiles
	}

	return tasks.Result{
		Status:      tasks.Success,
		Payload:     configFilesInfo,
		Summary:     finalSummary + warningSummaryCannotCollect,
		FilesToCopy: filesToCopy,
	}
}

func isConfigFileinPathToIgnore(dir string) bool {
	for _, path := range pathsToIgnore {
		if strings.Contains(dir, path) {
			return true
		}
	}
	return false
}

func appendToInvalidOrFoundConfigs(configPath string, warningSummaryOnInvalidFiles *string, invalidConfigFiles, foundConfigs []string) ([]string, []string) {

	pathInfo, err := os.Stat(configPath)
	if err != nil {
		invalidConfigFiles = append(invalidConfigFiles, configPath)
		*warningSummaryOnInvalidFiles += fmt.Sprintf(warningSummaryFmt, configPath, err.Error())
		return invalidConfigFiles, foundConfigs
	}
	//this conditional will probably get skipped for a path set by system property as it requires a full path that includes the yml file, so it is not a directory
	if pathInfo.IsDir() {
		foundFiles := tasks.FindFiles(patterns, []string{configPath})
		for _, file := range foundFiles {
			//make sure we are not adding a config file we already found by looking at standard places
			if tasks.PosString(foundConfigs, file) == -1 {
				foundConfigs = append(foundConfigs, file)
			}
		}
		return invalidConfigFiles, foundConfigs
	}
	//first make sure we are not adding a config file we already found by looking at standard places
	if tasks.PosString(foundConfigs, configPath) == -1 {
		//validate that the path provided takes us to a valid file that can be collected
		fileStatus := tasks.ValidatePath(configPath)
		if !(fileStatus.IsValid) {
			invalidConfigFiles = append(invalidConfigFiles, fileStatus.Path)
			*warningSummaryOnInvalidFiles += fmt.Sprintf(warningSummaryFmt, fileStatus.Path, fileStatus.ErrorMsg.Error())
			return invalidConfigFiles, foundConfigs
		}
		foundConfigs = append(foundConfigs, configPath)
	}
	return invalidConfigFiles, foundConfigs
}
