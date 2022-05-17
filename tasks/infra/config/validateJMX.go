package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// Values used when no host or port are found in the jmx-config.yml
const defaultJMXhost = "localhost"
const defaultJMXport = "9999"

// InfraConfigValidateJMX - This struct defines the task
type InfraConfigValidateJMX struct {
	mCmdExecutor             func(tasks.CmdWrapper, tasks.CmdWrapper) ([]byte, error)
	getJMXProcessCmdlineArgs func() []string
}

type JmxConfig struct {
	Host                  string
	Port                  string
	User                  string
	Password              string
	CollectionFiles       string
	JavaVersion           string
	JmxProcessCmdlineArgs []string
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
		Host                  string   `json:"jmx_host"`
		Port                  string   `json:"jmx_port"`
		User                  string   `json:"jmx_user"`
		Password              string   `json:"jmx_pass"`
		CollectionFiles       string   `json:"collection_files"`
		JavaVersion           string   `json:"java_version"`
		JmxProcessCmdlineArgs []string `json:"jmx_process_arguments"`
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
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
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
	if upstream["Java/Env/Version"].Status == tasks.Success {
		javaVersion, ok := upstream["Java/Env/Version"].Payload.(string)
		if ok {
			jmxKeys.JavaVersion = javaVersion
		}
	} else { //this implies tasks.Warning and not tasks.None because that task would definitely attempt to run once Infra/Config/Agent was found
		jmxKeys.JavaVersion = "Unable to find a Java path/version after running the command: java -version"
	}
	jmxCmdlineArgs := p.getJMXProcessCmdlineArgs()
	if len(jmxCmdlineArgs) < 1 {
		jmxKeys.JmxProcessCmdlineArgs = []string{"Unable to find a running JVM process that has JMX enabled or configured in its arguments"}
	} else {
		jmxKeys.JmxProcessCmdlineArgs = jmxCmdlineArgs
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

func getJMXProcessCmdlineArgs() []string {
	collectedJmxArgs := []string{}

	javaProcs := tasks.GetJavaProcArgs()
	for _, proc := range javaProcs {
		for _, arg := range proc.Args {
			//We only capture jvm args that are jmx configuration related
			if !(strings.Contains(arg, "jmx")) {
				continue
			}
			if strings.Contains(arg, "password") || strings.Contains(arg, "access") || strings.Contains(arg, "login") {
				/* Remove values from sensitive arguments
				https://docs.oracle.com/javase/8/docs/technotes/guides/management/agent.html
				com.sun.management.jmxremote.login.config
				com.sun.management.jmxremote.access.file
				com.sun.management.jmxremote.password.file
				*/
				keyVal := strings.Split(arg, "=")
				collectedJmxArgs = append(collectedJmxArgs, keyVal[0]+"="+"_REDACTED_")
			} else {
				collectedJmxArgs = append(collectedJmxArgs, arg)
			}
		}
	}

	return collectedJmxArgs
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
			return JmxConfig{}, fmt.Errorf("multiple key %s found", key)
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
		return JmxConfig{}, errors.New("invalid configuration found: collection_files not set")
	}

	return jmxExtractedConfig, nil
}

func (p InfraConfigValidateJMX) checkJMXServer(detectedConfig JmxConfig) error {
	// echo "*:type=*,name=*" | nrjmx -hostname 127.0.0.1 -port 9999 --verbose true // basic command that returns all the things
	// this queries for all beans, and givens back all types and all names
	var cmd1 tasks.CmdWrapper
	if runtime.GOOS == "windows" {
		//The first argument passed to exec.Command is the name of an executable (somefile.exe) and "echo" is not. To use shell commands, call the shell executable, and pass in the built-in command (and parameters)
		cmd1 = tasks.CmdWrapper{
			Cmd:  "cmd.exe",
			Args: []string{"/C", "echo", "*:type=*,name=*"},
		}
	} else {
		cmd1 = tasks.CmdWrapper{
			Cmd:  "echo",
			Args: []string{"*:type=*,name=*"},
		}
	}

	jmxArgs := buildNrjmxArgs(detectedConfig)
	var nrjmxCmd string
	if runtime.GOOS == "windows" {
		nrjmxCmd = `C:\Program Files\New Relic\nrjmx\nrjmx` //backticks to escape backslashes
	} else {
		nrjmxCmd = "nrjmx"
	}
	cmd2 := tasks.CmdWrapper{
		Cmd:  nrjmxCmd, // note we're using nrjmx here instead of nr-jmx, nrjmx is the raw connect to JMX command while nr-jmx is the wrapper that queries based on collection files
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
