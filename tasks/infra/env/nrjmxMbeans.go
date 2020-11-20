package env

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
)

// InfraEnvnrjmxMbeans - This struct defines the task
type InfraEnvnrjmxMbeans struct {
	mCmdExecutor               func(cmdWrapper, cmdWrapper) ([]byte, error)
	executeNrjmxCmdToFindBeans func(string) (string, error)
}

//cmdWrapper is used to specify commands & args to be passed to the multi-command executor (mCmdExecutor)
//allowing for: cmd1 args | cmd2 args
type cmdWrapper struct {
	cmd  string
	args []string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraEnvnrjmxMbeans) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Env/nrjmxMbeans")
}

// Explain - Returns the help text for each individual task
func (p InfraEnvnrjmxMbeans) Explain() string {
	return "Validate list of Mbeans against JMX integrations"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraEnvnrjmxMbeans) Dependencies() []string {
	return []string{
		"Infra/Config/ValidateJMX",
	}
}

func (p InfraEnvnrjmxMbeans) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Infra/Config/ValidateJMX"].Status != tasks.Warning || upstream["Infra/Config/ValidateJMX"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Infra/Config/ValidateJMX did not pass our validation. That issue will need to be resolved first before this task can be executed.",
		}
	}

	jmxConfig, ok := upstream["Infra/Config/ValidateJMX"].Payload.(config.JmxConfig)

	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: type assertion failure",
		}
	}

	beanQueries := getBeanQueriesFromJMVMetricsYml(jmxConfig.CollectionFiles)

	success, emptyResult, errResult := executeNrjmxCmdToFindBeans(beanQueries, jmxConfig)

	var failureResult string
	if len(emptyResult) > 0 {
		failureResult = fmt.Sprintf("We are able to run our nrjmx integration directly against your JMX server. However, the following query/queries found in jvm-metrics.yml may not exist or will need to be redefined in your yml because we are unable to find it among the MBeans listed on your server: %s", strings.Join(emptyResult, ", "))
	}

	// if cmdOutput == {} {
	// 	return tasks.Result{
	// 		Status: tasks.Failure,
	// 		Summary: "We are able to run our nrjmx integration directly against your JMX server. However, the query found in jvm-metrics.yml may not exist or will need to be redefined in your yml because we are unable to find it among the MBeans listed on your server.",
	// 	}
	// }

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Successfully connected to configured JMX Integration config",
	}

}

func getBeanQueriesFromJMVMetricsYml(pathToJvmMetricsYml string) []string {

}

func executeNrjmxCmdToFindBeans(beanQueries []string, jmxConfig config.JmxConfig) ([]string, []string, map[string]string) {

	//echo 'Glassbox:type=OfflineHandler,name=Offline_client_clingine' | ./nrjmx -H localhost -P 5002 -v -d
	//{}
	successCmdOutputs := []string{}
	errorCmdOutputs := make(map[string]string)
	emptyCmdOutputs := []string{}
	for _, query := range beanQueries {
		cmd1 := cmdWrapper{
			cmd:  "echo",
			args: []string{query},
		}
		jmxArgs := []string{"-hostname", jmxConfig.Host, "-port", jmxConfig.Port, "-v", "-d", "-"}
		cmd2 := cmdWrapper{
			cmd:  "./nrjmx", // note we're using nrjmx here instead of nr-jmx, nrjmx is the raw connect to JMX command while nr-jmx is the wrapper that queries based on collection files
			args: jmxArgs,
		}

		output, err := multiCmdExecutor(cmd1, cmd2)
		log.Debug("output", string(output))
		//nrjmx returns an error exit code (in err) and the meaningful error in output if there is a failure connecting
		//if nrjmx is not installed, output will be empty and the meaninful msg will be in err
		if err != nil {
			if len(output) == 0 {
				errorCmdOutputs[query] = err.Error()
			}
			errorCmdOutputs[query] = err.Error() + ": " + string(output)
		}
		if strings.TrimSpace(string(output)) == "{}" {
			emptyCmdOutputs = append(emptyCmdOutputs, query)
		} else {
			successCmdOutputs = append(successCmdOutputs, query)
		}
	}
	return successCmdOutputs, emptyCmdOutputs, errorCmdOutputs
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
