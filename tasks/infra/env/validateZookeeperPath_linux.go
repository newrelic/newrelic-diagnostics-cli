package env

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
)

// InfraEnvValidateZookeeperPath - This struct defines the task
type InfraEnvValidateZookeeperPath struct {
	cmdExec tasks.CmdExecFunc
}

const (
	defaultZookeeperPort = "2181"
	defaultZookeeperPath = "/brokers/ids"
	kafkaEnvVar          = "KAFKA_HOME"
	zookeeperEnvVar      = "ZOOKEEPER_HOME"
	pathEnvVar           = "PATH"
)

type ZookeeperConfig struct {
	Port string
	Path string
}
type ZookeeperHost struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraEnvValidateZookeeperPath) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Env/ValidateZookeeperPath")
}

// Explain - Returns the help text for each individual task
func (p InfraEnvValidateZookeeperPath) Explain() string {
	return "validate Kafka integration connection to zookeeper by collecting list of active brokers"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraEnvValidateZookeeperPath) Dependencies() []string {
	return []string{
		"Infra/Config/IntegrationsMatch",
		"Base/Env/CollectEnvVars",
	}
}

func (p InfraEnvValidateZookeeperPath) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Config/IntegrationsMatch"].Status == tasks.None {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No On-host Integration config and definitions files were collected. Task not executed.",
		}
	}
	if upstream["Base/Env/CollectEnvVars"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.ThisProgramFullName + " was unable to collect env vars from current shell. This task did not run",
		}
	}

	integrationFiles, ok := upstream["Infra/Config/IntegrationsMatch"].Payload.(infraConfig.MatchedIntegrationFiles)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	kafkaConfigPair, ok := integrationFiles.IntegrationFilePairs["kafka"]
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No Kafka Integration config or definition files were found. Task not executed.",
		}
	}

	zookeeperConfig, err := getZookeeperConfigFromYml(kafkaConfigPair)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("We were unable to run this health check because we ran into an error:%s", err.Error()),
		}
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	zookeeperShellPath, isOlderShellVersion := getZookeeperShellScriptPath(envVars) //zookeeper-shell.sh is a popular CLI tool to connect to zookeeper. We are going to run it passing some arguments to get a list of kafka brokers IDs. Getting the list is our way of validating that the New Relic Kafka integration can connect to Zookeeper/kafka

	if zookeeperShellPath == "" {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: `This health check cannot be completed because it requires running the zookeeper-shell script and we were unable to locate your Kafka directory through the current $PATH. If you are sure to have it, and happen to be running "sudo" alongside our Diagnostics CLI tool, keep in mind that "sudo" will reset your path. You can avoid this with: sudo -E`,
		}
	}
	/*how we pass the args to the zk shell may differ depending on the version of the shell:
	zkCli.sh localhost:2181 get /brokers/ids/0
	vs
	zookeeper-shell.sh localhost:2181 ls /brokers/ids
	*/
	var getArg, brokersArg string
	if isOlderShellVersion {
		getArg = "get"
		brokersArg = defaultZookeeperPath + "/0"
	} else {
		getArg = "ls"
		brokersArg = defaultZookeeperPath
	}

	hasBrokerList, resultSummary := p.getKafkaBrokersList(zookeeperConfig, zookeeperShellPath, getArg, brokersArg)

	if !hasBrokerList {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: resultSummary,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: resultSummary,
	}
}

func getZookeeperShellScriptPath(envVars map[string]string) (string, bool) {
	var path string

	kafkaPath, isKafkaEnvVarPresent := envVars[kafkaEnvVar]
	zookeeperPath, isZookeeperEnvVarPresent := envVars[zookeeperEnvVar]
	bashPath, isBashPathPresent := envVars[pathEnvVar]

	if isKafkaEnvVarPresent {
		//this is our preferred option for finding the path to zookeeper script because is the most popular way to bundle configuration for both kafka and zookeeper
		path = kafkaPath //Example: KAFKA_HOME=/opt/kafka_2.13-2.6.0
	} else if isZookeeperEnvVarPresent {
		path = zookeeperPath //Example: ZOOKEEPER_HOME=C:\Tools\zookeeper-3.4.9
	} else if isBashPathPresent && strings.Contains(bashPath, "kafka") {
		/*Examples of paths
		/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/kafka2.13/bin:/fknvfk
		/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin:/opt/something/kafka:/someotherpaths
		*/
		regex := regexp.MustCompile(`([^:]+kafka([^:]+|))`)
		matches := regex.FindStringSubmatch(bashPath)
		if len(matches) > 0 {
			path = matches[0]
		}
	}

	if len(path) < 1 {
		return "", false
	}
	//It's possible that the path does not include /bin/ (the directory that has zookeeper-shell.sh), in which case we need to add it. Example of full path to the zookeeper shell: /opt/kafka_2.13-2.6.0/bin/zookeeper-shell.sh
	var binPath string
	if strings.Contains(path, "bin") {
		binPath = path
	} else {
		binPath = path + "/bin"
	}
	//Now verify existing of zookeeper-shell
	zkFileStatus := tasks.ValidatePath(binPath + "/zookeeper-shell.sh")
	if zkFileStatus.IsValid {
		return binPath + "/zookeeper-shell.sh", false
	}
	//maybe they have the other less popular version of zookeeper-shell called zkCli.sh
	olderZkFileStatus := tasks.ValidatePath(binPath + "/zkCli.sh")
	if olderZkFileStatus.IsValid {
		return binPath + "/zkCli.sh", true
	}
	return "", false
}

func (p InfraEnvValidateZookeeperPath) getKafkaBrokersList(zookeeperConfig ZookeeperConfig, zookeeperShellPath string, getArg string, brokersArg string) (bool, string) {
	var hostPortArg string
	if len(zookeeperConfig.Port) > 0 {
		hostPortArg = "localhost:" + zookeeperConfig.Port
	} else {
		hostPortArg = "localhost:" + defaultZookeeperPort
	}

	if len(zookeeperConfig.Path) < 1 {
		/*When the zookeeper_path has not been set, our cmd will look like this:
		$ /path-to/zookeeper-shell.sh localhost:2181 ls /brokers/ids
		Instead of:
		$ /path-to/zookeeper-shell.sh localhost:2181 ls {my-zookeeper_path-value}/brokers/ids
		*/
		cmd := zookeeperShellPath + " " + hostPortArg + " " + getArg + " " + brokersArg
		cmdOutput, cmdErr := p.cmdExec("/bin/bash", "-c", cmd)
		if cmdErr != nil {
			return false, fmt.Sprintf("We ran the command - %s - and were unable to locate a list of brokers:\n%s\n%s\nThis might be due to the Zookeeper nodes not being network accessible to where the integration is in place, or Zookeeper is not running, or it could be that the Zookeeper namespace has your broker information kept under a different path other than the default. Keep in mind that an alternative Zookeeper path can be set in the kafka-config.yml: https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/kafka-monitoring-integration#arguments", cmd, cmdErr, string(cmdOutput))
		}
		//is not our job to validate those brokers IDs, just to make sure they are accessible
		return true, fmt.Sprintf("We ran the command %s and succesfully connected to a broker or list of brokers:\n%s", cmd, string(cmdOutput))
	}
	//We found a path set in config file (Example, zookeeper_path: "/kafka-root"), so we'll append it to the brokers argument: zookeeper-shell.sh localhost:2181 ls /kafka-root/brokers/ids
	customBrokersArg := zookeeperConfig.Path + brokersArg
	customCmd := zookeeperShellPath + " " + hostPortArg + " " + getArg + " " + customBrokersArg
	customCmdOutput, customCmdErr := p.cmdExec("/bin/bash", "-c", customCmd)

	if customCmdErr != nil {
		//The path set in config file is likely an invalid path; let's attempt to connect to zookeeper using the default path: /brokers/ids
		defaultCmd := zookeeperShellPath + " " + hostPortArg + " " + getArg + " " + brokersArg
		defaultCmdOutput, defaultCmdErr := p.cmdExec("/bin/bash", "-c", defaultCmd)
		if defaultCmdErr != nil {
			return false, fmt.Sprintf("We ran the command %s and were unable to locate a list of brokers:\n%s\n%s\nJust in case, we also attempted to run the same command with the zookeeper default path (%s), but we also got an error:%s\nThis might be due to the Zookeeper nodes not being network accessible to where the integration is in place, or Zookeeper is not running, or it could be that the Zookeeper namespace has your broker information kept under a different path. An alternative configuration you can try is to the bootstrap discovery mechanism, which will cause the integration to instead reach out to your defined bootstrap broker(s) to collect information on the other brokers in your cluster\nhttps://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/kafka-monitoring-integration#arguments", customCmd, customCmdErr, string(customCmdOutput), defaultCmd, defaultCmdErr)
		}
		//though we were able to access brokers through an alternative path, we still want to set this as a failure status to let the user know the path set in the config file will no provide the connection to zookeeper they are expecting
		return false, fmt.Sprintf("First we ran the command - %s - using the value found for zookeeper_path in the kafka config file, and this path did not connect us to your list of brokers IDs:\n%s\n%s\nHowever, when we ran the same command using the kafka integration default path for zookeeper - %s -, we were able to connect to your brokers:\n%s\nWe recommend commenting out the zookeeper_path setting and letting the integration automatically connect to the default path.", customCmd, customCmdErr, string(customCmdOutput), defaultCmd, string(defaultCmdOutput))
	}
	//Yay, we did not get an error when running the command using the zookeeper_path set in the config file
	return true, fmt.Sprintf("We ran the command - %s - and succesfully connected to a list of brokers:\n%s", customCmd, customCmdOutput)
}

func getZookeeperConfigFromYml(kafkaConfigPair *infraConfig.IntegrationFilePair) (ZookeeperConfig, error) {
	zookeeperExtractedConfig := ZookeeperConfig{}
	zookeeperPathBlobs := kafkaConfigPair.Configuration.ParsedResult.FindKey("zookeeper_path")

	if len(zookeeperPathBlobs) > 1 {
		return ZookeeperConfig{}, fmt.Errorf("multiple keys found for zookeeper_path")
	} else if len(zookeeperPathBlobs) == 1 {
		zookeeperExtractedConfig.Path = zookeeperPathBlobs[0].Value()
	}
	/*unlike zookeeper_path value that is a simple string, zookeeper_hosts needs to be parsed because the value looks like this:
	zookeeper_hosts: '[{"host": "something10.company.com", "port": 5101},{"host": "something11.company.com", "port": 5101},{"host": "something12.company.com", "port": 5101}]'
	*/
	zookeeperHostsBlobs := kafkaConfigPair.Configuration.ParsedResult.FindKey("zookeeper_hosts")
	var zookeeperHosts []*ZookeeperHost
	jsonErr := json.Unmarshal([]byte(zookeeperHostsBlobs[0].Value()), &zookeeperHosts)
	if jsonErr != nil {
		return ZookeeperConfig{}, fmt.Errorf("failed to parse zookeepers from json: %s", jsonErr)
	}

	for _, zookeeperHost := range zookeeperHosts {
		if zookeeperHost.Port != 0 {
			zookeeperExtractedConfig.Port = strconv.Itoa(zookeeperHost.Port)
		}
	} //it should be fine if the value for port gets overwritten, we expect the port values to be the same
	return zookeeperExtractedConfig, nil
}

//not been used currently, though it may come handy later if we encounter bugs:
//nolint
func cmdOutputHasBrokerIds(cmdOutput string) bool {
	captureGroup := `\[([0-9]|,|\s)+\]`
	re := regexp.MustCompile(captureGroup)
	return re.MatchString(cmdOutput)
}
