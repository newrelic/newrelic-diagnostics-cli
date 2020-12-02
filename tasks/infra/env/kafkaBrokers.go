package env

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
)

// InfraEnvKafkaBrokers - This struct defines the task
type InfraEnvKafkaBrokers struct {
}

const (
	defaultZookeeperPort = "2181"
	defaultZookeeperPath = "/brokers/ids"
)

type ZookeeperConfig struct {
	Port string
	Path string
}
type ZookeeperHost struct {
	Host string `json:"host"`
	Port string `json:"port"`
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraEnvKafkaBrokers) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Env/KafkaBrokers")
}

// Explain - Returns the help text for each individual task
func (p InfraEnvKafkaBrokers) Explain() string {
	return "get list of active kafka brokers ids that are available to the New Relic Kafka integration"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraEnvKafkaBrokers) Dependencies() []string {
	return []string{
		"Infra/Config/ValidateJMX",
	}
}

func (p InfraEnvKafkaBrokers) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Config/IntegrationsMatch"].Status == tasks.None {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No On-host Integration config and definitions files were collected. Task not executed.",
		}
	}

	integrationFiles, ok := upstream["Infra/Config/IntegrationsMatch"].Payload.(infraConfig.MatchedIntegrationFiles)

	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
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
	hasBrokerList, resultSummary := findBrokersList(zookeeperConfig)

	//return an error that tells us that the cli binary is not found and we did not run this task

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

func findBrokersList(zookeeperConfig ZookeeperConfig) (bool, string) {
	var hostPortArg string
	if len(zookeeperConfig.Port) > 0 {
		hostPortArg = "localhost:" + zookeeperConfig.Port
	} else {
		hostPortArg = "localhost:" + defaultZookeeperPort
	}
	//zookeeper-shell is a popular CLI tool to connect to zookeeper. We are going to run it pointing to the zookeeper default path because the user has not set any custom paths through the config file
	if len(zookeeperConfig.Path) < 1 {
		cmdOutput, cmdErr := tasks.CmdExecutor("zookeeper-shell", hostPortArg, "ls", defaultZookeeperPath)
		cmdRan := "zookeeper-shell " + hostPortArg + " ls " + defaultZookeeperPath

		if cmdErr == nil && cmdOutputHasBrokerIds(string(cmdOutput)) {
			//is not our job to validate those brokers IDs, just to make sure they are accessible
			return true, fmt.Sprintf("We ran the command %s and succesfully found a list of brokers: %s", cmdRan, string(cmdOutput))
		}
		return false, fmt.Sprintf("We ran the command %s and were unable to locate a list of brokers: %s\n%s\nThis might be due to the Zookeeper nodes not being network accessible to where the integration is in place, or it could be that the Zookeeper namespace has your broker information kept under a different path. Keep in mind that an alternative Zookeeper path can be set in the kafka-config.yml: https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/kafka-monitoring-integration#arguments", cmdRan, cmdErr, string(cmdOutput))
	}
	//Run the same command using the custom path set by the user. Example cmd:zookeeper-shell localhost:5101 ls /mycustomkafkapath/brokers/ids
	customPathArg := zookeeperConfig.Path + defaultZookeeperPath
	customCmd := "zookeeper-shell " + hostPortArg + " ls " + customPathArg // We'll use this string for our result summary
	customCmdOuput, customCmdErr := tasks.CmdExecutor("zookeeper-shell", hostPortArg, "ls", customPathArg)

	if customCmdErr == nil && cmdOutputHasBrokerIds(string(customCmdOuput)) {
		return true, fmt.Sprintf("We ran the command %s and succesfully found a list of brokers: %s", customCmd, customCmdOuput)
	}
	//Using their custom path either gaves an error or an output that we did not expect; let's attempt to connect to zookeeper using the defaultpath because its likely they passed an invalid custom path
	defaultCmd := "zookeeper-shell " + hostPortArg + " ls " + defaultZookeeperPath
	defaultCmdOutput, defaultCmdErr := tasks.CmdExecutor("zookeeper-shell", hostPortArg, "ls", defaultZookeeperPath)
	if defaultCmdErr == nil && cmdOutputHasBrokerIds(string(defaultCmdOutput)) {
		//though we found broker ids through an alternative path, we still want to set this as false/failure status to let the user know their custom path will no provide the connection to zookeeper they are hoping for
		return false, fmt.Sprintf("First we ran the command %s using the value found for zookeeper_path setting, and this path did not connect us to your list of brokers IDs: %s\nHowever, when we ran the same command using the kafka integration default path for zookeeper (/brokers/ids), we were able to connect to your brokers: %s\n. We recommend commenting out the zookeeper_path setting and letting the integration automatically connect to the default path.", customCmd, customCmdErr, string(defaultCmdOutput))
	}
	return false, fmt.Sprintf("We ran the command %s and were unable to locate a list of brokers: %s\nJust in case, we also attempted to run the same command with the zookeeper default path (%s), but we also got an error:%s\nThis might be due to the Zookeeper nodes not being network accessible to where the integration is in place, or it could be that the Zookeeper namespace has your broker information kept under a different path. An alternative configuration you can try is to the bootstrap discovery mechanism, which will cause the integration to instead reach out to your defined bootstrap broker(s) to collect information on the other brokers in your cluster. The options for configuring this can be found on our knowledge base: https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/kafka-monitoring-integration#arguments", customCmd, customCmdErr, defaultCmd, defaultCmdErr)
}

func cmdOutputHasBrokerIds(cmdOutput string) bool {
	captureGroup := `\[([0-9]|,|\s)+\]`
	re := regexp.MustCompile(captureGroup)
	return re.MatchString(cmdOutput)
}

func getZookeeperConfigFromYml(kafkaConfigPair *infraConfig.IntegrationFilePair) (ZookeeperConfig, error) {
	zookeeperExtractedConfig := ZookeeperConfig{}
	zookeeperPathBlobs := kafkaConfigPair.Configuration.ParsedResult.FindKey("zookeeper_path")

	if len(zookeeperPathBlobs) > 1 {
		return ZookeeperConfig{}, fmt.Errorf("Multiple keys found for zookeeper_path")
	}
	zookeeperExtractedConfig.Path = zookeeperPathBlobs[0].Value()

	/*unlike zookeeper_path that is a simple string, zookeeper_hosts needs to be parsed because the values look like this:
	zookeeper_hosts: '[{"host": "something10.company.com", "port": 5101},{"host": "something11.company.com", "port": 5101},{"host": "something12.company.com", "port": 5101}]'
	*/
	kafkaConfigYmlPath := kafkaConfigPair.Configuration.ParsedResult.PathAndKey()
	yamlFile, err := ioutil.ReadFile(kafkaConfigYmlPath)
	if err != nil {
		return ZookeeperConfig{}, fmt.Errorf("We got a parsing error from %s: %s", kafkaConfigYmlPath, err.Error())
	}

	var zookeeperHosts []*ZookeeperHost
	jsonErr := json.Unmarshal(yamlFile, &zookeeperHosts)
	if jsonErr != nil {
		return ZookeeperConfig{}, fmt.Errorf("failed to parse zookeepers from json: %s", jsonErr)
	}

	for _, zookeeperHost := range zookeeperHosts {
		if zookeeperHost.Port != "" {
			zookeeperExtractedConfig.Port = zookeeperHost.Port
		}
	} //it should be fine if the value for port gets overwritten, we expect the port values to be the same

	return zookeeperExtractedConfig, nil

}
