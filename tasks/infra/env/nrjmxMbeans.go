package env

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
	"gopkg.in/yaml.v3"
)

type collectionDefinitionParser struct {
	Collect []struct {
		Domain    string                 `yaml:"domain"`
		EventType string                 `yaml:"event_type"`
		Beans     []beanDefinitionParser `yaml:"beans"`
	}
}
type beanDefinitionParser struct {
	Query      string        `yaml:"query"`
	Exclude    interface{}   `yaml:"exclude_regex"`
	Attributes []interface{} `yaml:"attributes"`
}

// InfraEnvNrjmxMbeans - This struct defines the task
type InfraEnvNrjmxMbeans struct {
	getMBeanQueriesFromJMVMetricsYml func(string) ([]string, error)
	executeNrjmxCmdToFindBeans       func([]string, infraConfig.JmxConfig) ([]string, map[string]string)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraEnvNrjmxMbeans) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Env/NrjmxMbeans")
}

// Explain - Returns the help text for each individual task
func (p InfraEnvNrjmxMbeans) Explain() string {
	return "Validate list of Mbeans against JMX integrations"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraEnvNrjmxMbeans) Dependencies() []string {
	return []string{
		"Infra/Config/ValidateJMX",
	}
}

func (p InfraEnvNrjmxMbeans) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Infra/Config/ValidateJMX"].Status == tasks.None || upstream["Infra/Config/ValidateJMX"].Status == tasks.Failure {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No jmx-config.yml was found or Infra/Config/ValidateJMX did not pass our validation (in which case, the issue will need to be resolved first before this task can be executed.)",
		}
	}

	jmxConfig, ok := upstream["Infra/Config/ValidateJMX"].Payload.(infraConfig.JmxConfig)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	mbeanQueries, parseYmlErr := p.getMBeanQueriesFromJMVMetricsYml(jmxConfig.CollectionFiles)

	if parseYmlErr != nil {
		//This may overkill as we expect that any parsing issues get caught early by the integration itself: https://github.com/newrelic/nri-jmx/blob/master/src/config_parse.go#L62
		return tasks.Result{
			Status:  tasks.Error,
			Summary: parseYmlErr.Error(),
		}
	}

	if len(mbeanQueries) < 1 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No queries have been configured yet through the JMX integration files. This task did not run.",
		}
	}

	mbeansNotFound, mbeansWithErr := p.executeNrjmxCmdToFindBeans(mbeanQueries, jmxConfig)

	summaryIntro := fmt.Sprintf("In order to validate your queries defined in your metrics yml file against our JMX integration, we attempted to parsed them and ran each of them with the command echo {yourquery} | nrjmx -H %s -P %s -v -d -\n", jmxConfig.Host, jmxConfig.Port)

	var summaryFailure string
	if len(mbeansNotFound) > 0 {
		summaryFailure = fmt.Sprintf("These queries returned an empty object({}): %s\nThis can mean that either those mBeans are not available to this JMX server or that the queries targetting them may need to be reformatted in the metrics yml file.\n", strings.Join(mbeansNotFound, ", "))
	}

	var summaryErr string
	if len(summaryErr) > 0 {
		//timeout error: point them out to docs about how to fix timeout error
		summaryErr = "These queries returned an error:\n"
		for query, errMsg := range mbeansWithErr {
			summaryErr = summaryErr + query + " - " + errMsg + "\n"
		}
	}

	if len(summaryErr) > 0 || len(summaryFailure) > 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: summaryIntro + summaryErr + summaryFailure,
			URL:     "https://docs.newrelic.com/docs/integrations/host-integrations/host-integrations-list/jmx-monitoring-integration#troubleshoot",
		}
	}

	summarySuccess := "All queries returned successful metrics! The nrjmx integration is able to connect to the JMX server and query all the Mbeans that you had configured through your collection files."
	return tasks.Result{
		Status:  tasks.Success,
		Summary: summaryIntro + summarySuccess,
	}

}

func getMBeanQueriesFromJMVMetricsYml(collectionFilesStr string) ([]string, error) {

	if collectionFilesStr == "" {
		return []string{}, nil
	}

	paths := strings.Split(collectionFilesStr, ",") //collection_files: "/etc/newrelic-infra/integrations.d/jvm-metrics.yml,/etc/newrelic-infra/integrations.d/tomcat-metrics.yml"

	formattedQueries := []string{}
	for _, path := range paths {
		//I am not doing a check for absolute path because the integration already is checking for that: https://github.com/newrelic/nri-jmx/blob/master/src/jmx.go#L103
		// Parse the yaml file into a raw definition
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			log.Debugf("failed to open %s: %s", path, err)
			return []string{}, fmt.Errorf("we got a parsing error from %s: %s", path, err.Error())
		}
		// Parse the file (Sample file: https://github.com/newrelic/nri-jmx/blob/master/jvm-metrics.yml.sample)
		var c collectionDefinitionParser
		if err := yaml.Unmarshal(yamlFile, &c); err != nil {
			log.Debugf("failed to parse collection: %s", err)
			return []string{}, err
		}

		for _, domain := range c.Collect {
			domainName := domain.Domain
			// For each bean in the domain
			for _, bean := range domain.Beans {
				queryName := bean.Query
				formattedQueries = append(formattedQueries, domainName+":"+queryName)
			}
		}
	} //end of looping through all metrics files found in collection_files

	return formattedQueries, nil
}

func executeNrjmxCmdToFindBeans(mBeanQueries []string, jmxConfig infraConfig.JmxConfig) ([]string, map[string]string) {

	errorCmdOutputs := make(map[string]string)
	emptyCmdOutputs := []string{}

	for _, query := range mBeanQueries {
		var cmd1 tasks.CmdWrapper
		if runtime.GOOS == "windows" {
			cmd1 = tasks.CmdWrapper{
				Cmd:  "cmd.exe",
				Args: []string{"/C", "echo", query},
			}
		} else {
			cmd1 = tasks.CmdWrapper{
				Cmd:  "echo",
				Args: []string{query},
			}
		}

		jmxArgs := []string{"-hostname", jmxConfig.Host, "-port", jmxConfig.Port, "-v", "-d", "-"}
		var nrjmxCmd string
		if runtime.GOOS == "windows" {
			nrjmxCmd = `C:\Program Files\New Relic\nrjmx\nrjmx` //backticks to escape backslashes
		} else {
			nrjmxCmd = "nrjmx"
		}
		cmd2 := tasks.CmdWrapper{
			Cmd:  nrjmxCmd,
			Args: jmxArgs,
		}

		//We perform a cmd that looks like this: echo 'Glassbox:type=OfflineHandler,name=Offline_client_query' | ./nrjmx -H localhost -P 5002 -v -d -
		cmdOutput, err := tasks.MultiCmdExecutor(cmd1, cmd2)
		log.Debug("cmdOutput", string(cmdOutput))

		if err != nil {
			if len(cmdOutput) == 0 {
				errorCmdOutputs[query] = err.Error()
			}
			errorCmdOutputs[query] = err.Error() + ": " + string(cmdOutput)
		}

		if strings.Contains(string(cmdOutput), "{}") {
			//it means data does not exist for that query, they should check query in jconsole and verify existing of Mbean
			emptyCmdOutputs = append(emptyCmdOutputs, query)
		}
	}

	return emptyCmdOutputs, errorCmdOutputs
}
