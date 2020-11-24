package config

import (
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// Values used when no host or port are found in the jmx-config.yml
const defaultJMXhost = "localhost"
const defaultJMXport = "9999"

// InfraConfigValidateJMX - This struct defines the task
type InfraConfigValidateJMX struct {
	mCmdExecutor func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error)
}

type JmxConfig struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	CollectionFiles       string
	JavaVersion           string
	JmxProcessCmdlineArgs string
}

func (j JmxConfig) MarshalJSON() ([]byte, error) {
	var sanitizedUserString string
	var sanitizedPasswordString string

	if j.User != "" {
		sanitizedUserString = "_REDACTED_"
	}

	if j.Password != "" {
		sanitizedPasswordString = "_REDACTED_"
	}
	//ps -ef | grep jmx
	return json.Marshal(&struct {
		Host                  string `json:"jmx_host"`
		Port                  string `json:"jmx_port"`
		User                  string `json:"jmx_user"`
		Password              string `json:"jmx_pass"`
		CollectionFiles       string `json:"collection_files"`
		JavaVersion           string `json:"java_version"`
		JmxProcessCmdlineArgs string `json:"ps_ef_grep_jmx"`
	}{
		Host:                  j.Host,
		Port:                  j.Port,
		User:                  sanitizedUserString,
		Password:              sanitizedPasswordString,
		CollectionFiles:       j.CollectionFiles,
		JavaVersion:           j.JavaVersion,
		JmxProcessCmdlineArgs: j.JmxProcessCmdlineArgs,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraConfigValidateJMX) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Config/ValidateJMX")
}

// Explain - Returns the help text for each individual task
func (p InfraConfigValidateJMX) Explain() string {
	return "Validate New Relic Infrastructure JMX integration configuration file"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraConfigValidateJMX) Dependencies() []string {
	return []string{
		"Infra/Config/IntegrationsMatch",
		"Java/Env/Version",
	}
}

func (p InfraConfigValidateJMX) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Config/IntegrationsMatch"].Status == tasks.None {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No On-host Integration config and definitions files were collected. Task not executed.",
		}
	}

	matchedIntegrationFiles, ok := upstream["Infra/Config/IntegrationsMatch"].Payload.(MatchedIntegrationFiles)

	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	jmxConfigPair, ok := matchedIntegrationFiles.IntegrationFilePairs["jmx"]
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No JMX Integration config or definition files were found. Task not executed.",
		}
	}

	jmxKeys, err := processJMXFiles(jmxConfigPair)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Failure,
			URL:     "https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection",
			Summary: "Unexpected results for jmx-config.yml: " + err.Error(),
		}
	}

	//This data (JavaVersion and JmxProcessCmdlineArgs) it's relevant for TSE troubleshooting process
	if upstream["Java/Env/Version"].Status != tasks.None {
		javaVersion, ok := upstream["Java/Env/Version"].Payload.(string)
		if ok {
			jmxKeys.JavaVersion = javaVersion
		}
	}
	portCmdlineArgs, err := p.getPortCmdlineArgs() //run ps -ef | grep jmx
	if err != nil {
		jmxKeys.JmxProcessCmdlineArgs = "We ran the command ps -ef | grep jmx and are unable to find a running process with JMX enabled: " + err.Error()
	} else {
		jmxKeys.JmxProcessCmdlineArgs = portCmdlineArgs
	}

	nrjmxErr := p.checkJMXServer(jmxKeys)
	if nrjmxErr != nil {
		log.Debug("nrjmxErr", nrjmxErr)
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "We tested the JMX integration connection to local JMXServer by running the command echo '*:*' | nrjmx -H localhost -P 8080 -v -d - and we found this error:\n" + nrjmxErr.Error(),
			URL:     "https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#troubleshoot",
			Payload: jmxKeys,
		}
	}
	//We are able to run our nrjmx integration directly against their JMX server
	if jmxKeys.Host == "" || jmxKeys.Port == "" {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Successfully connected to JMX server but no hostname or port defined in jmx-config.yml\nWe recommend configuring this instead of relying on default parameters",
			URL:     "https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#host-connection",
			Payload: jmxKeys,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Successfully connected to configured JMX Integration config",
		Payload: jmxKeys,
	}

}

func (p InfraConfigValidateJMX) getPortCmdlineArgs() (string, error) {
	cmd1 := tasks.CmdWrapper{
		Cmd:  "ps",
		Args: []string{"-ef"},
	}
	cmd2 := tasks.CmdWrapper{
		Cmd:  "grep",
		Args: []string{"jmx"},
	}
	output, err := p.mCmdExecutor(cmd1, cmd2)
	log.Debug("output", string(output))

	if err != nil {
		if len(output) == 0 {
			return "", err
		}
		return "", errors.New(string(output))
	}

	return string(output), nil
}

func processJMXFiles(jmxConfigPair *IntegrationFilePair) (JmxConfig, error) {
	jmxKeys := []string{
		"jmx_host",
		"jmx_port",
		"jmx_user",
		"jmx_pass",
		"collection_files",
	}
	jmxExtractedConfig := JmxConfig{}
	for _, key := range jmxKeys {
		foundkey := jmxConfigPair.Configuration.ParsedResult.FindKey(key)
		if len(foundkey) > 1 {
			return JmxConfig{}, fmt.Errorf("Multiple key %s found", key)
		}

		for _, fieldValue := range foundkey {
			switch key {
			case "jmx_host":
				jmxExtractedConfig.Host = fieldValue.Value()
			case "jmx_port":
				jmxExtractedConfig.Port = fieldValue.Value()
			case "jmx_user":
				jmxExtractedConfig.User = fieldValue.Value()
			case "jmx_pass":
				jmxExtractedConfig.Password = fieldValue.Value()
			case "collection_files":
				jmxExtractedConfig.CollectionFiles = fieldValue.Value()
			}
		}
	}

	//collection_files is minimum required key for jmx-config.yml
	if jmxExtractedConfig.CollectionFiles == "" {
		return JmxConfig{}, errors.New("Invalid configuration found: collection_files not set")
	}

	return jmxExtractedConfig, nil
}

func (p InfraConfigValidateJMX) checkJMXServer(detectedConfig JmxConfig) error {

	// echo "*:type=*,name=*" | nrjmx -hostname 127.0.0.1 -port 9999 --verbose true // basic command that returns all the things
	// this queries for all beans, and givens back all types and all names
	cmd1 := tasks.CmdWrapper{
		Cmd:  "echo",
		Args: []string{"*:type=*,name=*"},
	}

	jmxArgs := buildNrjmxArgs(detectedConfig)
	cmd2 := tasks.CmdWrapper{
		Cmd:  "nrjmx", // note we're using nrjmx here instead of nr-jmx, nrjmx is the raw connect to JMX command while nr-jmx is the wrapper that queries based on collection files
		Args: jmxArgs,
	}

	output, err := p.mCmdExecutor(cmd1, cmd2)
	log.Debug("output", string(output))
	//nrjmx returns an error exit code (in err) and the meaningful error in output if there is a failure connecting
	//if nrjmx is not installed, output will be empty and the meaninful msg will be in err
	if err != nil {
		if len(output) == 0 {
			return err
		}
		return errors.New(string(output))
	}

	return nil
}

func buildNrjmxArgs(detectedConfig JmxConfig) []string {

	var nrjmxArgs []string

	if detectedConfig.Host == "" {
		detectedConfig.Host = defaultJMXhost
		log.Debug("Adding default hostname value to JMX configuration")
	}

	if detectedConfig.Port == "" {
		detectedConfig.Port = defaultJMXport
		log.Debug("Adding default port value to JMX configuration")
	}

	nrjmxArgs = append(nrjmxArgs, "-hostname")
	nrjmxArgs = append(nrjmxArgs, detectedConfig.Host)

	nrjmxArgs = append(nrjmxArgs, "-port")
	nrjmxArgs = append(nrjmxArgs, detectedConfig.Port)

	if detectedConfig.User != "" {
		nrjmxArgs = append(nrjmxArgs, "-username")
		nrjmxArgs = append(nrjmxArgs, detectedConfig.User)
	}

	if detectedConfig.Password != "" {
		nrjmxArgs = append(nrjmxArgs, "-password")
		nrjmxArgs = append(nrjmxArgs, detectedConfig.Password)
	}

	return nrjmxArgs
}
