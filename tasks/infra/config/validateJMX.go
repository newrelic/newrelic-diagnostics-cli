package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// Values used when no host or port are found in the jmx-config.yml
const defaultJMXhost = "localhost"
const defaultJMXport = "9999"

// InfraConfigValidateJMX - This struct defines the task
type InfraConfigValidateJMX struct {
	mCmdExecutor func(cmdWrapper, cmdWrapper) ([]byte, error)
}

//cmdWrapper is used to specify commands & args to be passed to the multi-command executor (mCmdExecutor)
//allowing for: cmd1 args | cmd2 args
type cmdWrapper struct {
	cmd  string
	args []string
}

type JmxConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	CollectionFiles string
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

	return json.Marshal(&struct {
		Host            string `json:"jmx_host"`
		Port            string `json:"jmx_port"`
		User            string `json:"jmx_user"`
		Password        string `json:"jmx_pass"`
		CollectionFiles string `json:"collection_files"`
	}{
		Host:            j.Host,
		Port:            j.Port,
		User:            sanitizedUserString,
		Password:        sanitizedPasswordString,
		CollectionFiles: j.CollectionFiles,
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

	nrjmxErr := p.checkJMXServer(jmxKeys)
	if nrjmxErr != nil {
		log.Debug("nrjmxErr", nrjmxErr)
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "Error connecting to local JMXServer:\n" + nrjmxErr.Error(),
			URL:     "https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#troubleshoot",
			Payload: jmxKeys,
		}
	}

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
	cmd1 := cmdWrapper{
		cmd:  "echo",
		args: []string{"*:type=*,name=*"},
	}

	jmxArgs := buildNrjmxArgs(detectedConfig)
	cmd2 := cmdWrapper{
		cmd:  "nrjmx", // note we're using nrjmx here instead of nr-jmx, nrjmx is the raw connect to JMX command while nr-jmx is the wrapper that queries based on collection files
		args: jmxArgs,
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

// takes multiple commands and pipes the first into the second
func multiCmdExecutor(cmdWrapper1, cmdWrapper2 cmdWrapper) ([]byte, error) {

	cmd1 := exec.Command(cmdWrapper1.cmd, cmdWrapper1.args...)
	cmd2 := exec.Command(cmdWrapper2.cmd, cmdWrapper2.args...)

	// Get the pipe of Stdout from cmd1 and assign it
	// to the Stdin of cmd2.
	pipe, err := cmd1.StdoutPipe()
	if err != nil {
		return []byte{}, err
	}
	cmd2.Stdin = pipe

	// Start() cmd1, so we don't block on it.
	err = cmd1.Start()
	if err != nil {
		return []byte{}, err
	}

	// Run Output() on cmd2 to capture the output.
	return cmd2.CombinedOutput()

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
