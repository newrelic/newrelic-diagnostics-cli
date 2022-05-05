package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/env"
	"github.com/shirou/gopsutil/process"
)

// JavaConfigValidate - This struct defined the sample plugin which can be used as a starting point
type JavaConfigValidate struct {
}

type JavaValidatedConfig struct {
	Proc              process.Process
	ParsedResult      tasks.ValidateBlob
	ConfigPath        string
	CurrentWorkingDir string
}

//MarshalJSON - custom JSON marshaling for this task, in this case we ignore the ParsedResult
func (el JavaValidatedConfig) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Proc              process.Process
		ConfigPath        string
		CurrentWorkingDir string
	}{
		Proc:              el.Proc,
		ConfigPath:        el.ConfigPath,
		CurrentWorkingDir: el.CurrentWorkingDir,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t JavaConfigValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Config/Validate")
}

// Explain - Returns the help text for each individual task
func (t JavaConfigValidate) Explain() string {
	return "Pair detected Java processes with corresponding New Relic Java agent config files"
}

// Dependencies - Returns the dependencies for ech task.
func (t JavaConfigValidate) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
		"Base/Config/Validate",
		"Java/Env/Process",
	}
}

// Execute - The core work within each task
func (t JavaConfigValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	// Validate Java Config Agent Success
	if upstream["Java/Config/Agent"].Status != tasks.Success && upstream["Java/Env/Process"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Java agent not detected. This task did not run."
		return result
	}

	// Get Validate and Process Payloads with type assertions
	if !upstream["Base/Config/Validate"].HasPayload() {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Unable to validate a new relic config file. This task did not run.",
		}
	}
	validations, ok := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)
	if !ok {
		result.Status = tasks.Error
		result.Summary = tasks.AssertionErrorSummary
		return result
	}

	if upstream["Java/Env/Process"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Java/Env/Process check did not pass. This task did not run.",
		}
	}

	processes, ok := upstream["Java/Env/Process"].Payload.([]env.ProcIdAndArgs)
	if !ok {
		result.Status = tasks.Error
		result.Summary = tasks.AssertionErrorSummary
		return result
	}

	// Step through processes to identify the active running config
	var javaConfigs []JavaValidatedConfig
	for _, process := range processes {
		var config JavaValidatedConfig
		config.CurrentWorkingDir = process.Cwd
		config.Proc = process.Proc

		config.ConfigPath, config.ParsedResult = matchConfigFile(config.CurrentWorkingDir, validations, process.CmdLineArgs)

		config.ParsedResult = replaceEnv(process, config.ParsedResult)

		// Update config info with SysProps
		config.ParsedResult = updateSysValues(process, config.ParsedResult)

		// Update config info with Environment Variables
		config.ParsedResult = updateEnvValues(process.EnvVars, config.ParsedResult)

		// Pass along the valid running config for each Process

		javaConfigs = append(javaConfigs, config)
	}

	result.Status = tasks.Success
	result.Payload = javaConfigs

	for _, javaConfig := range javaConfigs {
		stream := make(chan string)
		output := fmt.Sprintf("PID: %d\n", javaConfig.Proc.Pid)
		output += "ConfigFile: " + javaConfig.ConfigPath + "\n"
		output += "CurrentWorkingDir: " + javaConfig.CurrentWorkingDir + "\n"
		output += "WorkingConfig: \n" + javaConfig.ParsedResult.String()

		go streamBlob(output, stream)

		result.FilesToCopy = append(result.FilesToCopy, tasks.FileCopyEnvelope{Path: fmt.Sprintf("Config_%d.txt", javaConfig.Proc.Pid), Stream: stream})
	}

	return result
}

func streamBlob(input string, ch chan string) {
	defer close(ch)

	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}

}

func matchConfigFile(processWorkingDir string, validations []config.ValidateElement, processCmdLineArgs []string) (configPath string, parsedResult tasks.ValidateBlob) {
	//check for config file specified on command line - newrelic.config.file

	for _, cmdLineArg := range processCmdLineArgs {
		if strings.Contains(cmdLineArg, "newrelic.config.file=") {
			_, configPath = splitSystemProp(cmdLineArg)
		}
	}

	if configPath == "" {
		// match validation with process
		configPath, parsedResult = matchConfig(processWorkingDir, validations)
	} else {
		//Check for a validation that matches the ConfigPath and use that validation blob
		for _, validation := range validations {
			if validation.Config.FilePath+validation.Config.FileName == configPath {
				parsedResult = validation.ParsedResult
			} 
		}
	}
	// last ditch effort to find the config file
	if configPath == "" {
		log.Debug("No config file found, finding jar and looking for newrelic.yml")
		jarPath := ""
		for _, cmdLineArg := range processCmdLineArgs {
			if strings.Contains(cmdLineArg, "-javaagent") {
				_, jarPath = splitjavaAgent(cmdLineArg)
			}
		}
		jarPath = filepath.Dir(filepath.Clean(jarPath))

		log.Debug("jarlocation is", jarPath)
		if tasks.FileExists(jarPath + "/newrelic.yml") {
			configPath = jarPath + "/newrelic.yml"
		} else {
			log.Debug("Config file not found in path:", jarPath)
		}

	}

	// now check for a parsed result and parse the file if we don't have one
	if configPath != "" && parsedResult.IsLeaf() {
		//Now we need to create a new validated blob
		log.Debug("Configpath is", configPath)
		file, err := os.Open(configPath)
		if err != nil {
			log.Debug("error reading file", err)
		}
		parsedResult, err = config.ParseYaml(file)
		if err != nil {
			log.Debug("error reading file", err)
		}

	}

	return
}

// matchConfig - This takes the current working directory and attempts to reconstruct the config file path to point to a validated config blob
func matchConfig(processWorkingDir string, validations []config.ValidateElement) (configPath string, validation tasks.ValidateBlob) {
	//processWorkingDir = "/Users/sdelight/git/tomcat/apache-tomcat-8.0.45/" // local override so mac dev work can work since Cwd isn't supported on mac

	for _, validationElement := range validations {
		log.Debug("processWorkingDir is", filepath.Dir(processWorkingDir), "end")
		log.Debug("validation path is", validationElement.Config.FilePath)
		split := strings.Split(filepath.Dir(validationElement.Config.FilePath), "/")
		log.Debug("length is ", len(split))

		if split[len(split)-1] == "newrelic" {
			log.Debug("split is ", split)
			split = split[:len(split)-1] // remove last element of slice if it's newrelic
		}

		log.Debug("joined path is", "/"+filepath.Join(split...), "end")
		if filepath.Dir(processWorkingDir) == "/"+filepath.Join(split...) {
			log.Debug("matched processWorkingDir and validation path")
			return validationElement.Config.FilePath + validationElement.Config.FileName, validationElement.ParsedResult
		}
	}

	log.Debug("Failed to find matching config file in detected config file")

	return
}

func replaceEnv(config env.ProcIdAndArgs, parsedValues tasks.ValidateBlob) tasks.ValidateBlob {
	environ := determineEnvironment(config.CmdLineArgs)
	return parsedValues.FindKeyByPath("/" + environ)
}

type javaMapping struct {
	mappingVar    string
	configSetting string
}

func updateSysValues(config env.ProcIdAndArgs, parsedValues tasks.ValidateBlob) tasks.ValidateBlob {
	// first look for explicit config file setting via Sysprop

	sysPropMappings := []javaMapping{
		{mappingVar: "-Dnewrelic.debug", configSetting: "-Dnewrelic.config.debug"},
		{mappingVar: "-Dnewrelic.logfile", configSetting: "-Dnewrelic.config.log_file_name"},
	}

	for _, arg := range config.CmdLineArgs {

		if strings.Contains(strings.ToLower(arg), "newrelic") {

			//map out these keys to override the values in the blob
			key, value := splitSystemProp(arg)
			if key == "-Dnewrelic.config.file" || key == "-Dnewrelic.home" {
				log.Debug("Key matched is", key, "skipping")
				continue //Skip this
			}
			for _, mapping := range sysPropMappings {

				if key == mapping.mappingVar {
					log.Debug("Found matching special case sysprop", arg, "updating to", mapping.configSetting)
					key = mapping.configSetting
				}
			}
			log.Debug("Got override System Property", key, ":", value)
			key = strings.Replace(strings.Replace(key, "-Dnewrelic.config.", "", 1), ".", "/", -1)
			log.Debug("Searching for", key)

			//now the generic replacement for anything with newrelic.config

			parsedValues = parsedValues.UpdateOrInsertKey(key, value)
		}

	}
	return parsedValues

}

func determineEnvironment(cmdLineArgs []string) (environ string) {
	for _, arg := range cmdLineArgs {
		if strings.Contains(strings.ToLower(arg), "newrelic.environment=") {
			log.Debug("Found newrelic environment", arg)
			_, environ := splitSystemProp(arg)
			log.Debug("environment is", environ)
		}
	}
	if environ == "" {
		environ = "production" //failsafe setting to the default value of production if not set elsewhere
	}
	return
}

// splitSystemPro - Splits a java system property into key/value pairing
func splitSystemProp(input string) (key, value string) {
	split := strings.SplitN(input, "=", 2)
	if len(split) != 2 {
		log.Debug("didn't get equal key/value pair", input)
		return "", ""
	}
	return split[0], split[1]
}

// splitSystemPro - Splits a java system property into key/value pairing
func splitjavaAgent(input string) (key, value string) {
	split := strings.SplitN(input, ":", 2)
	if len(split) != 2 {
		log.Debug("didn't get equal key/value pair", input)
		return "", ""
	}
	return split[0], split[1]
}

func updateEnvValues(envVars map[string]string, parsedResult tasks.ValidateBlob) tasks.ValidateBlob {
	// do struct mapping of environment value to config file setting to override the parsedResult
	envVarMappings := []javaMapping{
		{mappingVar: "NEW_RELIC_APP_NAME", configSetting: "app_name"},
		{mappingVar: "NEW_RELIC_PROCESS_HOST_DISPLAY_NAME", configSetting: "process_host/display_name"},
		{mappingVar: "NEW_RELIC_LOG", configSetting: "log_file_name"},
		{mappingVar: "NEW_RELIC_LICENSE_KEY", configSetting: "license_key"},
	}

	// loop through mappings to see if any exist, and if so, replace them

	for _, mapping := range envVarMappings {
		if value, ok := envVars[mapping.mappingVar]; ok {
			log.Debug("Env value", mapping.mappingVar, "exists")
			log.Debug("Setting new value to", value)
			parsedResult = parsedResult.UpdateOrInsertKey(mapping.configSetting, value)
		}
	}

	return parsedResult
}
