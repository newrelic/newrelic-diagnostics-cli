// +build linux darwin

package jvm

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/env"
)

/* PermissionsPayload - struct used to construct the eventual result payload */
type PermissionsPayload struct {
	PID                int32
	AgentJarCanRead    payloadField //The owner of the Java process to which the -javaagent option will be passed must have read permissions for the java agent JAR
	LogDirCanCreate    payloadField //The process owner requires write/execute permissions for the directory in which the log directory will be created and execute permissions for all parent directories of the log directory so the java process can traverse into the directory and create the java agent log file
	LogFileCanCreate   payloadField
	LogFileCanWrite    payloadField
	TempFilesCanCreate payloadField //The process owner must have write/execute access to the temp directory for the Java process. This may be the default directory for temporary Java files (specified system-wide), or it may be one specific to the process,
}

/* struct used to store if permissions are set correctly for a file/dir and if an error was encountered while determining permissions*/
type payloadField struct {
	Success  bool
	ErrorMsg error
	Source   string
	Value    string
}

var logFile = map[string]string{
	"envVar": "NEW_RELIC_LOG",      //The unqualified log file name(not a path) or the string STDOUT which will log to standard out. The latter would inmediately give a "permission denied" error so no much need to troubleshoot for this option
	"sysProp": "-Dnewrelic.logfile", //EX: Dnewrelic.logfile=/opt/newrelic/java/logs/newrelic/somenewnameformylogs.log
	"configFile": "log_file_path",      ////Java, ruby: "/Users/shuayhuaca/Desktop/
	//if log_file_path is specified, the directory must already exist. If the default value is used, the agent will attempt to create the directory.
}

var tempDir = []string{
	"-Dnewrelic.tempdir", //can only be set as sys prop not as an env var neither the config file
	"-Djava.io.tmpdir",   //On UNIX systems the default value of this property can be "/tmp" or "/var/tmp"; on Windows "c:\temp". A different path value may be given through this system property
}

type JavaJVMPermissions struct {
}

// Identifier - returns the Category (Agent), Subcategory (Java) and Name (Permissions)
func (p JavaJVMPermissions) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/JVM/Permissions") // This should be updated to match the struct name
}

// Explain - Returns the help text for the customer for this task
func (p JavaJVMPermissions) Explain() string {
	return "Check Java process permissions meet New Relic Java agent requirements"
}

/* Dependencies - Depends on the the SysPropCollect task to get PIDs with corresponding command-line args */
func (p JavaJVMPermissions) Dependencies() []string {
	return []string{
		"Base/Env/CollectEnvVars",
		"Base/Env/CollectSysProps",
		"Java/Env/Process",
	}
}

/* Execute - This task checks for processes running new relic java agents and determines if permissions are set correctly to run the Agent */
func (p JavaJVMPermissions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	/* if there are no running Java Agents */
	if upstream["Java/Env/Process"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "The New Relic agent has yet need to be added to a running JVM process. This task did not run.",
		}
	}


	/* if there is at least one running Java Agent */
	/* obtain slice of java processes and command-line args from SysPropCollect task */
	javaProcs, ok := upstream["Java/Env/Process"].Payload.([]env.ProcIdAndArgs)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "We were unable to run this health check due to an internal type assertion error for the task Java/Env/Process",
		}
	}

	for _, process := range javaProcs {
		/* Can the user read the Agent JAR */
		agentJARReadError := canRunJAR(process.Proc, values.AgentJarLocation.Value)
	}
	//pass the path to the jar + "/logs" + envvarfilename
	logDir := getLogDirectory(upstream)

}

func getLogDirectory(upstream map[string]tasks.Result) string {
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Upstream environment variables check failed.")
	}
}

/* need read permissions only; "java" is being executed and it needs to read the JAR */
func canRunJAR(proc process.Process, jarLoc string) (err error) {
	/* check if the Java Agent JAR exists */
	if _, errJarNotExist := os.Stat(jarLoc); os.IsNotExist(errJarNotExist) {
		log.Debug(taskName + "Agent JAR does not exist for PID " + fmt.Sprint(proc.Pid) + ". This location is at " + jarLoc)
		err = errJarNotExist
		return
	}
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID := getFileAndProcUIDsGIDs(proc, jarLoc)
	if procOwnerUID == "" && procOwnerGID == "" && fileOwnerUID == "" && fileOwnerGID == "" {
		err = errors.New("The permissions settings for the Agent JAR are indeterminable")
		return
	}
	info, errGettingJarPerms := os.Stat(jarLoc)
	if errGettingJarPerms != nil {
		log.Debug("There was an error obtaining JAR permissions for JAR at " + jarLoc + ". Error is: " + errGettingJarPerms.Error())
	}
	filePerm := info.Mode().Perm()
	/* the process owner also is the file owner for the Agent JAR */
	if procOwnerUID == fileOwnerUID {
		matched, errRegexMatch := regexp.MatchString("-r.+", filePerm.String())
		if errRegexMatch != nil {
			log.Debug(taskName, err)
			err = errRegexMatch
			return
		}
		if matched {
			/* the process/file owner has read permissions */
			return
		}
		/* the process owner is part of the group to which the Agent JAR belongs */
	} else if procOwnerGID == fileOwnerGID {
		matched, errRegexMatch := regexp.MatchString("[rwx-]{4}r.+", filePerm.String())
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the file group has read permissions */
			return
		}
	} else {
		/* process owner is neither the Agent JAR file owner nor part of the JAR's specified group */
		matched, errRegexMatch := regexp.MatchString("[rwx-]{7}r.+", filePerm.String())
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* all have read permissions */
			return
		}
	}
	err = errors.New("The owner of the process for PID " + fmt.Sprint(proc.Pid) + "does not have permissions to execute the New Relic Agent JAR located at " + jarLoc)
	return
}

