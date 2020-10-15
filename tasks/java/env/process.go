// +build !windows

package env

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/shirou/gopsutil/process"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

/* structure to contain a process and its corresponding command line args */
type ProcIdAndArgs struct {
	Proc        process.Process
	CmdLineArgs []string
	Cwd         string
	EnvVars     map[string]string
}

type JavaEnvProcess struct {
	name           string
	findProcByName tasks.FindProcessByNameFunc
	getCmdLineArgs func(process.Process) (string, error)
	getCurrentDir  func(process.Process, string) string
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

	/* range through java processes and add process IDs and a slice containing their corresponding
	command line args to find the ones that include -javaagent argument */
	javaAgentProcsIdArgs := []ProcIdAndArgs{}
	for _, proc := range javaProcs {
		cmdLineArgsStr, err := p.getCmdLineArgs(proc)
		if err != nil {
			log.Debug("Error getting command line options")
		}
		/* check the command line args contain the new relic Java Agent JAR, called newrelic.jar, though few customer may change the name of the file */
		if strings.Contains(cmdLineArgsStr, "newrelic.jar") {
			CmdLineArgsList := strings.Split(cmdLineArgsStr, " ")
			cwd := p.getCurrentDir(proc, cmdLineArgsStr)
			javaAgentProcsIdArgs = append(javaAgentProcsIdArgs, ProcIdAndArgs{Proc: proc, CmdLineArgs: CmdLineArgsList, Cwd: cwd, EnvVars: envVars})
		}
	}

	if len(javaAgentProcsIdArgs) > 0 {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: fmt.Sprintf("We detected %d New Relic Java Agent(s) running on this host.", len(javaAgentProcsIdArgs)),
			Payload: javaAgentProcsIdArgs,
		}
	}

	return tasks.Result{
		Status:  tasks.Failure,
		Summary: "None of the active Java processes included the -javaagent argument. For proper installation of New Relic Java agent, the -javaagent flag must be passed to the same Java process that is running your application",
		URL:     "https://docs.newrelic.com/docs/agents/java-agent/installation/include-java-agent-jvm-argument",
	}
}

//getCmdLineArgs is a wrapper for dependency injecting proc.Cmdline in testing
func getCmdLineArgs(proc process.Process) (string, error) {
	return proc.Cmdline()
}

//getCurrentDir
func getCurrentDir(process process.Process, cmdLineArgsStr string) string {

	cwd, err := process.Cwd()

	// This command does not work properly on darwin and it will return an error. If that's the case, let's get the cwd from this backup:
	if err != nil {
		log.Debug("error: ", err)
		//Example: -javaagent:/Users/shuayhuaca/Desktop/projects/java/luces/newrelic.jar
		regex := regexp.MustCompile("-javaagent:(.+)newrelic.jar")
		result := regex.FindStringSubmatch(cmdLineArgsStr)
		if result != nil {
			cwd = result[1]
		}
	}

	return cwd
}
