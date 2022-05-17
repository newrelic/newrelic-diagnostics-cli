package env

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/shirou/gopsutil/v3/process"
)

/* structure to contain a process and its corresponding command line args */
type ProcIdAndArgs struct {
	Proc        process.Process
	CmdLineArgs []string
	Cwd         string
	JarPath     string
	EnvVars     map[string]string
}

type JavaEnvProcess struct {
	findProcByName tasks.FindProcessByNameFunc
	getCmdLineArgs func(process.Process) (string, error)
	getCwd         func(process.Process) (string, error)
}

// Identifier - returns the Category (Agent), Subcategory (Java) and Name (SysPropCollect)
func (p JavaEnvProcess) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Env/Process") // This should be updated to match the struct name
}

// Explain - Returns the help text for the customer for each individual task
func (p JavaEnvProcess) Explain() string {
	return "Identify running Java processes that include the New Relic Java agent argument"
}

/* Task should depend on the presence of a config file. It means customer is intending to install Java agent*/
func (p JavaEnvProcess) Dependencies() []string {
	return []string{
		"Base/Env/CollectEnvVars",
		"Java/Config/Agent",
	}
}

// This task checks for processes running new relic java agents and returns those processes' command line arguments */
func (p JavaEnvProcess) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Java/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Java agent config file was not detected on this host. This task did not run",
		}
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("No Env Vars detected")
	}

	/* search for the existence of all JVM/java processes */
	javaProcs, err := p.findProcByName("java")
	log.Debug("SysPropCollect processes", javaProcs)
	if err != nil {
		return tasks.Result{
			Summary: "We encountered an error while detecting all running Java processes: " + err.Error(),
			Status:  tasks.Error,
		}
	}

	if len(javaProcs) == 0 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: tasks.ThisProgramFullName + " is unable to validate the presence of the New Relic -javaagent flag because you have no java processes running at this time. Please re-run " + tasks.ThisProgramFullName + " after starting your Java Agent application.",
		}
	}

	/* range through java processes and add process IDs and a slice containing their corresponding
	command line args to find the ones that include -javaagent argument */
	javaAgentProcsIdArgs := []ProcIdAndArgs{}
	for _, proc := range javaProcs {
		cmdLineArgsStr, err := p.getCmdLineArgs(proc)
		if err != nil {
			log.Debug("Error getting command line options")
		}

		if strings.Contains(cmdLineArgsStr, "-javaagent") {

			jarPath, jarFilename, newRelicAgentNotFoundErr := getJarInfoFromCmdLineArgs(cmdLineArgsStr)
			if newRelicAgentNotFoundErr != nil {
				return tasks.Result{
					Status:  tasks.Failure,
					Summary: "Failed to find the New Relic Java Agent Jar in the following JVM argument: " + newRelicAgentNotFoundErr.Error() + "\nIf this is another Java Agent, keep in mind that New Relic is not compatible with other additional agents.",
				}
			}

			cwd, cwdErr := p.getCwd(proc)
			// getCwd() does not work properly on darwin and will return an error. If that's the case, we can use the jar path as fallback
			if cwdErr != nil {
				cwd = jarPath
			}

			CmdLineArgsList := strings.Split(cmdLineArgsStr, " ")
			javaAgentProcsIdArgs = append(javaAgentProcsIdArgs, ProcIdAndArgs{Proc: proc, CmdLineArgs: CmdLineArgsList, Cwd: cwd, JarPath: filepath.Join(jarPath, jarFilename), EnvVars: envVars})
		}
	}

	if len(javaAgentProcsIdArgs) > 0 {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: fmt.Sprintf("We detected %d New Relic Java Agent(s) running on this host.", len(javaAgentProcsIdArgs)),
			Payload: javaAgentProcsIdArgs,
		}
	}
	//Java’s built-in argument called “-javaagent”
	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "None of the current active Java processes includes the '-javaagent' argument. For proper installation of the New Relic Java agent, the -javaagent argument must be passed to the same process that is running your application. Examples on how to include this argument can be found in the documentation listed below.",
		URL:     "https://docs.newrelic.com/docs/agents/java-agent/installation/include-java-agent-jvm-argument\nhttps://discuss.newrelic.com/t/relic-solution-what-you-need-to-know-about-new-relic-when-deploying-with-docker/52492\n",
	}
}

//getCmdLineArgs is a wrapper for dependency injecting proc.Cmdline in testing
func getCmdLineArgs(proc process.Process) (string, error) {
	return proc.Cmdline()
}

func getJarInfoFromCmdLineArgs(cmdLineArgsString string) (string, string, error) {
	var err error

	javaAgentCmd := "javaagent:"

	var javaAgentArg string
	for _, arg := range strings.Split(cmdLineArgsString, " ") {
		if strings.Contains(arg, javaAgentCmd) {
			javaAgentArg = arg
		}
	}

	firstIndexOfPath := strings.Index(javaAgentArg, javaAgentCmd) + len(javaAgentCmd)
	lastIndexOfPath := strings.LastIndex(javaAgentArg, string(filepath.Separator))

	// if there is no path in the javaAgentArg
	if lastIndexOfPath == -1 {
		lastIndexOfPath = firstIndexOfPath - 1
	}

	fileName := javaAgentArg[lastIndexOfPath+1:]

	/* check that the command line args contain the new relic Java Agent JAR, called newrelic.jar, though few customer may change the name of the file. The filename must include 'newrelic' in it.
	Example: -javaagent:/Users/shuayhuaca/Desktop/projects/java/luces/newrelic.jar */
	isFileNameCorrect := strings.Contains(fileName, "newrelic") && strings.Contains(fileName, ".jar")

	if !isFileNameCorrect {
		err = errors.New(javaAgentArg)
	}

	path := javaAgentArg[firstIndexOfPath : lastIndexOfPath+1]
	if !strings.Contains(path, string(filepath.Separator)) {
		path = "." + string(filepath.Separator)
	}
	return path, fileName, err
}

func getCwd(proc process.Process) (string, error) {
	return proc.Cwd()
}
