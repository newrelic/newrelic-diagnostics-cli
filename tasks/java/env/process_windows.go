package env

import (
	"strings"

	"github.com/shirou/gopsutil/process"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

/* structure to contain a process and its corresponding command line args */
type ProcIdAndArgs struct {
	Proc        process.Process
	CmdLineArgs []string
	Cwd         string
	EnvVars     map[string]string
}

type JavaEnvProcess struct {
	name string
}

// Identifier - returns the Category (Agent), Subcategory (Java) and Name (SysPropCollect)
func (p JavaEnvProcess) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Env/Process") // This should be updated to match the struct name
}

// Explain - Returns the help text for the customer for each individual task
func (p JavaEnvProcess) Explain() string {
	return "Collect Java process JVM command line options"
}

func (p JavaEnvProcess) Dependencies() []string {
	return []string{
		"Base/Env/CollectEnvVars",
	}
}

// This task checks for processes running new relic java agents and returns those processes' command line arguments */
func (p JavaEnvProcess) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	var result tasks.Result

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("No Env Vars detected")
	}
	// Base case will be returned if running NR Java Agent is not detected
	result.Status = tasks.None
	result.Summary = "Java Agent was not detected on this host."

	log.Debug("Checking for the existence of JVM processes with attached New Relic Java Agents.")
	procIdAndArgSlice := []ProcIdAndArgs{}
	/* search for all java processes */
	procObjs, err := tasks.FindProcessByName("java")
	if err != nil {
		log.Debug("Error getting process", err)
	}
	processResults, err := tasks.GetProcInfoByName("java%") 
	log.Debug("SysPropCollect processes ", len(processResults))
	if err != nil {
		log.Debug("Error getting process info", err)
	}

	/* range through java processes and add process IDs and a slice containing their corresponding
	command line args */
	for _, value := range processResults {
		cmdLineArgs := value.CommandLine

		//if err != nil {
		//	log.Debug("Error getting command line options")
		//}
		/* check to make sure the command line args contain the new relic Java Agent JAR
		change the result status and summary if one or more processes with newrelic.jar are found
		this check assumed the java agent jar is called newrelic.jar, but this is not a requirement with more recent agent versions */

		if strings.Contains(value.CommandLine, "newrelic.jar") {
			procObj := procObjs[0]
			for _, proc := range procObjs {
				if proc.Pid == int32(value.ProcessId) {
					procObj = proc
				}
			}

			log.Debug("Found processes containing newrelic.jar ", procObj)

			result.Status = tasks.Success
			result.Summary = "At least one running Java Agent detected on this host."
			sliceOfCmdLineArgs := strings.Split(cmdLineArgs, " ")
			cwd := value.ExecutablePath                           //Throwing away error here because we won't get this anyway in some cases
			procIdAndArgSlice = append(procIdAndArgSlice, ProcIdAndArgs{Proc: procObj, CmdLineArgs: sliceOfCmdLineArgs, Cwd: cwd, EnvVars: envVars})
		}
	}

	result.Payload = procIdAndArgSlice
	return result
}
