// +build linux darwin

package jvm

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	"github.com/shirou/gopsutil/process"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/java/env"
)

/* PermissionsPayload - struct used to construct the eventual result payload */
type PermissionsPayload struct {
	PID                int32
	AgentJARRead       payloadField
	LogDirCanCreate    payloadField
	LogFileCanCreate   payloadField
	LogFileCanWrite    payloadField
	TempFilesCanCreate payloadField
	ProcessAgentValues agentValues
}

/* struct used to store if permissions are set correctly for a file/dir and if an error was encountered while determining permissions*/
type payloadField struct {
	Success           bool
	PayloadFieldError error
}

/* stores locations of important files and directories */
type agentValues struct {
	AgentJarLocation valueSource
	YamlFile         valueSource
	LogPath          valueSource
	LogFileName      valueSource
	LogFile          valueSource
	TempDir          valueSource
}

/* struct to store value (e.g. log file location, temp directory location) and Source from which this value is obtained (e.g. obtained from config file or from environment variable) */
type valueSource struct {
	Value  string
	Source string
}

var (
	taskName      = "JavaJVMPermissions - "
	envVarSource  = "environment variable"
	sysPropSource = "system property or JVM arg"
	yamlSource    = "configuration file"
	defaultSource = "New Relic Java Agent Default"
)

type JavaJVMPermissions struct {
	name string
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
		"Java/Env/Process",
	}
}

/* Execute - This task checks for processes running new relic java agents and determines if permissions are set correctly to run the Agent */
func (p JavaJVMPermissions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	/* obtain map of environment variables from upstream task */
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug(taskName, "Upstream environment variables check failed.")
	} else {
		log.Debug(taskName, "Successfully gathered Environment Variables from upstream.")
	}
	/* slice to store what will eventually be our result payload */
	var payload []PermissionsPayload
	var result tasks.Result

	result.Status = tasks.None
	result.Summary = "The java permissions task is unable to report any meaningful information."

	/* if there are no running Java Agents */
	if upstream["Java/Env/Process"].Status != tasks.Success {
		return result
	}

	/* if there is at least one running Java Agent */
	/* obtain slice of java processes and command-line args from SysPropCollect task */
	javaProcs, ok := upstream["Java/Env/Process"].Payload.([]env.ProcIdAndArgs)
	if ok {
		log.Debug(taskName, "Java/Env/Process payload correct type")
	}

	for _, process := range javaProcs {
		/* single instance of a structure to be stored in our payload slice */
		var payloadEntity PermissionsPayload
		payloadEntity.PID = process.Proc.Pid
		/* slice of options passed to Java process */
		cmdLineArgs := process.CmdLineArgs
		/* Order of precedence for setting the agentValues struct : environment variable > system property > newrelic.yml > default */
		var values agentValues
		/* Environment variables have highest order of precedence; check first */
		values.LogFile, values.LogPath, values.LogFileName = checkEnvValues(envVars)
		/* System properties have second-highest order of precedence; check second */
		values = checkForSystemProps(cmdLineArgs, values)
		var err error
		/* Java Agent Config File has second-highest order of precedence; check second */
		values.YamlFile, values.LogFileName, values.LogPath, err = checkForYAMLValues(values)
		if err != nil {
			log.Debug("Error reading Java Agent config file for PID ", fmt.Sprint(process.Proc.Pid), ". Error is: ", err)
		}
		/* if any file locations haven't been set by env var, sys prop, or YAML, set to their defaults */
		values = setDefaultValues(values)
		payloadEntity.ProcessAgentValues = values

		/* Can the user read the Agent JAR */
		agentJARReadError := canRunJAR(process.Proc, values.AgentJarLocation.Value)
		payloadEntity.AgentJARRead = payloadField{(agentJARReadError == nil), agentJARReadError}
		/* can the user create the log directoy for the agent */
		logDirCanCreateError := canCreateLogDirectory(process.Proc, values.LogPath.Value)
		payloadEntity.LogDirCanCreate = payloadField{(logDirCanCreateError == nil), logDirCanCreateError}
		/* can the user create the log file */
		logFileCanCreateError := canCreateAgentLog(process.Proc, values.LogFile.Value)
		payloadEntity.LogFileCanCreate = payloadField{(logFileCanCreateError == nil), logFileCanCreateError}
		/* can the user write to the log */
		logFileCanWriteError := canWriteToLog(process.Proc, values.LogFile.Value, payloadEntity.LogFileCanCreate.Success)
		payloadEntity.LogFileCanWrite = payloadField{(logFileCanWriteError == nil), logFileCanWriteError}
		/* can the user create files in the temp directory */
		tempFileCanCreateError := canCreateFilesInTempDir(process.Proc, values.TempDir.Value)
		payloadEntity.TempFilesCanCreate = payloadField{(tempFileCanCreateError == nil), tempFileCanCreateError}
		payload = append(payload, payloadEntity)
	}
	result = determineResult(payload)
	return result
}

func determineResult(payloadEntities []PermissionsPayload) (result tasks.Result) {
	result.Payload = payloadEntities

	/* if any field in the payloadEntity contains false (if a check didn't pass)
	check if there is an error associated with the false value
	return warning or failure based on the error; failures are prioritized */
	for _, payloadEntity := range payloadEntities {
		log.Debug(taskName + "The value for the Java Agent JAR location for PID " + fmt.Sprint(payloadEntity.PID) + " is " + payloadEntity.ProcessAgentValues.AgentJarLocation.Value + " and was obtained from " + payloadEntity.ProcessAgentValues.AgentJarLocation.Source)
		if payloadEntity.AgentJARRead.PayloadFieldError != nil {
			if !payloadEntity.AgentJARRead.Success {
				result.URL = "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/error-starting-app-server-java"
				/* if the check didn't pass and we were able to determine permissions then the user has incorrect permissions */
				if strings.Contains(payloadEntity.AgentJARRead.PayloadFieldError.Error(), "does not have permissions") {
					result.Status = tasks.Failure
					result.Summary = "The process owner for PID " + fmt.Sprint(payloadEntity.PID) + " needs read permission for Agent JAR " + payloadEntity.ProcessAgentValues.AgentJarLocation.Value
					return result
				}
				/* the check didn't pass because we weren't able to determine permissions */
				result.Status = tasks.Warning
				result.Summary = "Permission settings cannot be determined for the New Relic Agent JAR at location " + payloadEntity.ProcessAgentValues.AgentJarLocation.Value + "for PID " + fmt.Sprint(payloadEntity.PID) + ". Problem is: " + payloadEntity.AgentJARRead.PayloadFieldError.Error()
				return result
			}
		}
		if payloadEntity.LogDirCanCreate.PayloadFieldError != nil {
			if !payloadEntity.LogDirCanCreate.Success {
				result.URL = "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/no-log-file-java#no-log-file"
				/* if the check didn't pass and we were able to determine permissions then the user has incorrect permissions */
				if strings.Contains(payloadEntity.LogDirCanCreate.PayloadFieldError.Error(), "does not have permissions") {
					result.Status = tasks.Failure
					result.Summary = "The process owner for PID " + fmt.Sprint(payloadEntity.PID) + " needs write permissions for directory " + filepath.Dir(payloadEntity.ProcessAgentValues.LogPath.Value)
					return result
				} else if strings.Contains(payloadEntity.LogDirCanCreate.PayloadFieldError.Error(), "file does not exist") {
					/* the directory containing the java agent log directory does not exist */
					result.Status = tasks.Failure
					result.Summary = "The log directory " + payloadEntity.ProcessAgentValues.LogPath.Value + "does not exist and cannot be created."
					return result
				}
				/* the check didn't pass because we weren't able to determine permissions */
				result.Status = tasks.Warning
				result.Summary = "Permission settings cannot be determined for the New Relic log directory at location " + payloadEntity.ProcessAgentValues.LogPath.Value + "for PID " + fmt.Sprint(payloadEntity.PID) + ". Problem is: " + payloadEntity.LogDirCanCreate.PayloadFieldError.Error()
				return result
			}
		}
		log.Debug(taskName + "The value for the Java Agent log file for PID " + fmt.Sprint(payloadEntity.PID) + " is " + payloadEntity.ProcessAgentValues.LogFile.Value + " and was obtained from " + payloadEntity.ProcessAgentValues.LogFile.Source)
		if payloadEntity.LogFileCanCreate.PayloadFieldError != nil {
			if !payloadEntity.LogFileCanCreate.Success {
				result.URL = "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/no-log-file-java#no-log-file"
				if strings.Contains(payloadEntity.LogFileCanCreate.PayloadFieldError.Error(), "does not have permissions") {
					result.Status = tasks.Failure
					result.Summary = "The process owner for PID " + fmt.Sprint(payloadEntity.PID) + " needs write permissions for directory " + payloadEntity.ProcessAgentValues.LogPath.Value
					return result
				} else if strings.Contains(payloadEntity.LogFileCanCreate.PayloadFieldError.Error(), "file does not exist") {
					result.Status = tasks.Failure
					result.Summary = "The log directory for PID " + fmt.Sprint(payloadEntity.PID) + "does not exist; therefore, the log file cannot be created."
					return result
				}
				/* the check didn't pass because we weren't able to determine permissions */
				result.Status = tasks.Warning
				result.Summary = "Permission settings cannot be determined for the New Relic Agent log file at location " + payloadEntity.ProcessAgentValues.LogFile.Value + "for PID " + fmt.Sprint(payloadEntity.PID) + ". Problem is: " + payloadEntity.LogFileCanWrite.PayloadFieldError.Error()
				return result
			}
		}
		if payloadEntity.LogFileCanWrite.PayloadFieldError != nil {
			if !payloadEntity.LogFileCanWrite.Success {
				result.URL = "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/no-log-file-java#no-log-file"
				if strings.Contains(payloadEntity.LogFileCanWrite.PayloadFieldError.Error(), "does not have permissions") {
					result.Status = tasks.Failure
					result.Summary = "The process owner for PID " + fmt.Sprint(payloadEntity.PID) + " needs write permissions for log file " + payloadEntity.ProcessAgentValues.LogFile.Value
					return result
				} else if strings.Contains(payloadEntity.LogFileCanWrite.PayloadFieldError.Error(), "file does not exist") {
					result.Status = tasks.Failure
					result.Summary = "The log file for PID " + fmt.Sprint(payloadEntity.PID) + " does not exist and cannot be created at " + payloadEntity.ProcessAgentValues.LogFile.Value
					return result
				}
				/* the check didn't pass because we weren't able to determine permissions */
				result.Status = tasks.Warning
				result.Summary = "Permission settings cannot be determined for the New Relic Agent log file at location " + payloadEntity.ProcessAgentValues.LogFile.Value + "for PID " + fmt.Sprint(payloadEntity.PID) + ". Problem is: " + payloadEntity.LogFileCanWrite.PayloadFieldError.Error()
				return result
			}
		}
		log.Debug(taskName + "The value for the Java Agent temp directory for PID " + fmt.Sprint(payloadEntity.PID) + " is " + payloadEntity.ProcessAgentValues.TempDir.Value + " and was obtained from " + payloadEntity.ProcessAgentValues.TempDir.Source)
		if payloadEntity.TempFilesCanCreate.PayloadFieldError != nil {
			if !payloadEntity.TempFilesCanCreate.Success {
				result.URL = "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/error-bootstrapping-new-relic-agent"
				if strings.Contains(payloadEntity.TempFilesCanCreate.PayloadFieldError.Error(), "does not have permissions") {
					result.Status = tasks.Failure
					result.Summary = "The process owner for PID " + fmt.Sprint(payloadEntity.PID) + " needs write permissions for directory " + payloadEntity.ProcessAgentValues.TempDir.Value
					return result
				} else if strings.Contains(payloadEntity.TempFilesCanCreate.PayloadFieldError.Error(), "file does not exist") {
					result.Status = tasks.Failure
					result.Summary = "The temp directory for PID " + fmt.Sprint(payloadEntity.PID) + " does not exist at " + payloadEntity.ProcessAgentValues.TempDir.Value + ". Therefore, temp files cannot be created by the new relic Java Agent. "
					return result
				}
				/* the check didn't pass because we weren't able to determine permissions */
				result.Status = tasks.Warning
				result.Summary = "Permission settings cannot be determined for the New Relic Agent temp directory at location " + payloadEntity.ProcessAgentValues.TempDir.Value + "for PID " + fmt.Sprint(payloadEntity.PID) + ". Problem is: " + payloadEntity.TempFilesCanCreate.PayloadFieldError.Error()
				return result
			}
		}
		log.Debug(taskName + "The value for the Java Agent config file location for PID " + fmt.Sprint(payloadEntity.PID) + " is " + payloadEntity.ProcessAgentValues.YamlFile.Value + " and was obtained from " + payloadEntity.ProcessAgentValues.YamlFile.Source)
	}
	/* if we've exited the for loop and haven't yet returned the result then permissions are set correctly for all PIDs running New Relic Java Agents - success! */
	result.Status = tasks.Success
	result.Summary = "All running Java Agent environments meet minimum permissions requirements."
	return result
}

func checkEnvValues(envVars map[string]string) (LogFile, LogPath, LogFileName valueSource) {
	envLogVal := envVars["NEW_RELIC_LOG"]
	/* NEW_RELIC_LOG=path/to/filename.log */
	if strings.Contains(envLogVal, "/") {
		LogFile = valueSource{envLogVal, envVarSource}
		LogPath = valueSource{filepath.Dir(LogFile.Value), envVarSource}
		LogFileName = valueSource{filepath.Base(LogFile.Value), envVarSource}
		log.Debug(taskName + "Setting log location (path+filename) to " + LogFile.Value + " from " + envVarSource)
	} else {
		/* NEW_RELIC_LOG=filename.log */
		LogFileName = valueSource{envLogVal, envVarSource}
		log.Debug(taskName + "Setting log file name to " + LogFileName.Value + " from " + LogFileName.Source)
	}
	return
}


func checkForSystemProps(args []string, values agentValues) (newValues agentValues) {
	newValues = values
	/* System property checks: range through each JVM option */
	for _, arg := range args {
		/* -javaagent argument specifies the agent JAR location and default locations for other files */
		if newValues.AgentJarLocation.Value == "" {
			newValues.AgentJarLocation = valueSource{checkForSingleProp("-javaagent:(.+newrelic\\.jar)", arg), sysPropSource}
			log.Debug(taskName + "Setting agent jar location to " + newValues.LogFile.Value + " from " + newValues.LogFile.Source)
		}
		/* -Dnewrelic.config.file=path/to/newrelic.yml */
		if newValues.YamlFile.Value == "" {
			newValues.YamlFile = valueSource{checkForSingleProp("-Dnewrelic.config.file=(.+)", arg), sysPropSource}
			log.Debug(taskName + "Setting yaml file location to " + newValues.YamlFile.Value + " from " + newValues.YamlFile.Source)
		}
		if newValues.LogPath.Value == "" {
			/* -Dnewrelic.config.log_file_path=/path/to/logdir */
			newValues.LogPath = valueSource{checkForSingleProp("-Dnewrelic.config.log_file_path=(.*)", arg), sysPropSource}
			log.Debug(taskName + "Setting log path to " + newValues.LogPath.Value + " from " + newValues.LogPath.Source)
		}
		if newValues.LogFileName.Value == "" {
			/* -Dnewrelic.log_file_name=filename.log */
			newValues.LogFileName = valueSource{checkForSingleProp("-Dnewrelic.config.log_file_name=(.*)", arg), sysPropSource}
			log.Debug(taskName + "Setting log file name to " + newValues.LogFileName.Value + " from " + newValues.LogFileName.Source)
		}
		if newValues.LogFileName.Value == "" {
			/* -Dnewrelic.LogFile=filename.log
			regardless of other system properties, this will mean the log file is at {$APP_HOME}/filename.log */
			wd, wdErr := os.Getwd()
			if wdErr == nil {
				// downstream logic relies on a non-match to return an empty string, hence the if statement
				if checkForSingleProp("-Dnewrelic.logfile=(.*)", arg) != "" {
					newValues.LogFile = valueSource{wd + "/" + checkForSingleProp("-Dnewrelic.logfile=(.*)", arg), sysPropSource}
					newValues.LogFileName = valueSource{checkForSingleProp("-Dnewrelic.logfile=(.*)", arg), sysPropSource}
					newValues.LogPath = valueSource{wd, sysPropSource}
					log.Debug(taskName + "Setting log file location to " + newValues.LogFile.Value + " from " + newValues.LogFile.Source)
				}
			} else {
				exeDir, getExeErr := os.Executable()
				if getExeErr == nil {
					// downstream logic relies on a non-match to return an empty string, hence the if statement
					if checkForSingleProp("-Dnewrelic.logfile=(.*)", arg) != "" {
						newValues.LogFile = valueSource{exeDir + checkForSingleProp("-Dnewrelic.logfile=(.*)", arg), sysPropSource}
						log.Debug(taskName + "Setting log file location to " + newValues.LogFile.Value + " from " + newValues.LogFile.Source)
					}
				} else {
					log.Debug(taskName + "Unable to determine the cwd or directory containing nrdiag executable. If you are using the system property newrelic.logfile to set the newrelic log file name, you should manually check that the log file is created properly.")
				}
			}
		}
		if newValues.TempDir.Value == "" {
			/* -Dnewrelic.TempDir=/path/to/temp/dir */
			newValues.TempDir = valueSource{checkForSingleProp("-Dnewrelic.tempdir=(.+)", arg), sysPropSource}
			log.Debug(taskName + "Setting temp dir location to " + newValues.TempDir.Value + " from " + newValues.TempDir.Source)
		}
		/* -Djava.io.tmpdir=/Path/to/temp */
		/* ensure this location hasn't already been set with -Dnewrelic.TempDir, which has greater precedence */
		if newValues.TempDir.Value == "" {
			newValues.TempDir = valueSource{checkForSingleProp("-Djava.io.tmpdir=(.+)", arg), sysPropSource}
			log.Debug(taskName + "Setting temp file location to " + newValues.TempDir.Value + " from " + newValues.TempDir.Source)
		}
	}
	return newValues
}

func checkForSingleProp(regex string, key string) (value string) {
	reg := regexp.MustCompile(regex)
	match := reg.FindStringSubmatch(key)
	if len(match) != 0 {
		return match[1]
	}
	return value
}

func checkForYAMLValues(values agentValues) (newYAMLFile, newLogFileName, newLogPath valueSource, err error) {
	newYAMLFile = values.YamlFile
	newLogFileName = values.LogFileName
	newLogPath = values.LogPath

	/* if YAML file location isn't set by sys prop, it should be in default location */
	if newYAMLFile.Value == "" {
		/* default YAML file location is in dir containing Agent JAR */
		newYAMLFile = valueSource{(filepath.Dir(values.AgentJarLocation.Value) + "/newrelic.yml"), yamlSource}
		log.Debug(taskName + "Setting yaml file location to " + newYAMLFile.Value + " from " + newYAMLFile.Source)
	}
	/* YAML file variable checks */
	if newLogFileName.Value == "" {

		rawFileName, err := tasks.ReturnLastStringSubmatchInFile("log_file_name: (.+)", newYAMLFile.Value)
		for _, v := range rawFileName {
			if !strings.Contains(v, "log_file_name:") {
				newLogFileName.Value = v
			}
		}

		if strings.Contains(fmt.Sprint(err), ("string not found in file")) {
			/* we aren't concerned if the customer does not set this variable in the newrelic.yml; it will later be set to the default value if so */
			newLogFileName = valueSource{"", ""}
			err = nil
		} else {
			newLogFileName.Source = yamlSource
			log.Debug(taskName + "Setting log file name to " + newLogFileName.Value + " from " + newLogFileName.Source)
		}
	}
	if newLogPath.Value == "" {
		rawFilePath, err := tasks.ReturnStringSubmatchInFile("log_file_path: (.+)", newYAMLFile.Value)
		for _, v := range rawFilePath {
			if !strings.Contains(v, "log_file_path:") {
				newLogPath.Value = v
			}
		}
		if strings.Contains(fmt.Sprint(err), ("string not found in file")) {
			/* we aren't concerned if the customer does not set this variable in the newrelic.yml; it will later be set to the default value if so */
			newLogPath = valueSource{"", ""}
			err = nil
		} else {
			newLogPath.Source = yamlSource
			log.Debug(taskName + "Setting log path to " + newLogPath.Value + " from " + newLogPath.Source)
		}
	}
	return newYAMLFile, newLogFileName, newLogPath, err
}

func setDefaultValues(values agentValues) (newValues agentValues) {
	newValues = values
	if values.LogPath.Value == "" {
		/* default log dir is in dir containing Agent JAR */
		newValues.LogPath = valueSource{(filepath.Dir(values.AgentJarLocation.Value) + "/logs"), defaultSource}
		log.Debug(taskName + "Setting log path to " + newValues.LogPath.Value + " from " + newValues.LogPath.Source)
	}
	if values.LogFileName.Value == "" {
		newValues.LogFileName = valueSource{"newrelic_agent.log", defaultSource}
		log.Debug(taskName + "Setting log file name to " + newValues.LogFileName.Value + " from " + newValues.LogFileName.Source)
	}
	
	if values.LogFile.Value == "" {
		newValues.LogFile = valueSource{(newValues.LogPath.Value + "/" + newValues.LogFileName.Value), defaultSource}
		log.Debug(taskName + "Setting log file to " + newValues.LogFile.Value + " using " + newValues.LogFile.Source)
	}
	if values.TempDir.Value == "" {
		newValues.TempDir = valueSource{os.TempDir(), defaultSource}
		log.Debug(taskName + "Setting temp directory location to " + newValues.TempDir.Value + " from " + newValues.TempDir.Source)
	}
	return newValues
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

/* process owner needs write/execute permissions for dir in which to create log dir */
func canCreateLogDirectory(proc process.Process, LogPath string) (err error) {
	/* logs directory already exists; return success */
	if _, errDirExist := os.Stat(LogPath); errDirExist == nil {
		err = errDirExist
		return
	}
	truncatedLog := filepath.Dir(LogPath)
	/* the logs directory does not already exist and the directory in which the logs directory will be created does not exist */
	if _, errDirExist := os.Stat(truncatedLog); os.IsNotExist(errDirExist) {
		err = errDirExist
		return
	}
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID := getFileAndProcUIDsGIDs(proc, truncatedLog)
	if procOwnerUID == "" && procOwnerGID == "" && fileOwnerUID == "" && fileOwnerGID == "" {
		err = errors.New("The permissions settings to create Agent log directory are indeterminable")
		return
	}
	info, errStatDir := os.Stat(truncatedLog)
	if errStatDir != nil {
		log.Debug(taskName, "Error determining directory permissions. Error is: ", err)
		err = errStatDir
		return
	}
	filePerm := info.Mode().Perm()
	/* the process owner also is the directory owner */
	if procOwnerUID == fileOwnerUID {
		matched, errRegexMatch := regexp.MatchString("[drwx-]{2}wx.+", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the process/dir owner has write permissions */
			return
		}
	} else if procOwnerGID == fileOwnerGID {
		/* the process owner is part of the dir group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{5}wx.+", fmt.Sprint(filePerm))
		if err != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the dir group has write permissions */
			return
		}
	} else {
		/* process owner is neither the dir owner nor part of the dir's specified group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{8}wx", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* all have write permissions */
			return
		}
	}
	err = errors.New("The owner of the process does not have permissions to create the Agent log directory")
	return
}

/* need write/execute access to log dir */
func canCreateAgentLog(proc process.Process, LogFile string) (err error) {
	/* logs file already exists; return success */
	if _, fileExistError := os.Stat(LogFile); fileExistError == nil {
		return
	}
	logDir := filepath.Dir(LogFile)
	/* if logs directory doesn't exist, log file cannot be created */
	if _, fileExistError := os.Stat(logDir); fileExistError != nil {
		err = fileExistError
		return
	}
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID := getFileAndProcUIDsGIDs(proc, logDir)
	if procOwnerUID == "" && procOwnerGID == "" && fileOwnerUID == "" && fileOwnerGID == "" {
		err = errors.New("Permission settings to create Agent log file at " + LogFile + " are indeterminable for PID " + fmt.Sprint(proc.Pid))
		return
	}
	info, errStatDir := os.Stat(logDir)
	if errStatDir != nil {
		log.Debug(taskName, "Error determining permissions regarding agent log file creation. Error is: ", err)
		err = errStatDir
		return
	}
	filePerm := info.Mode().Perm()
	/* the process owner also is the directory owner */
	if procOwnerUID == fileOwnerUID {
		matched, errRegexMatch := regexp.MatchString("[drwx-]{2}wx.+", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the process/dir owner has write permissions */
			return
		}
	} else if procOwnerGID == fileOwnerGID {
		/* the process owner is part of the dir group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{5}wx.+", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the dir group has write permissions */
			return
		}
	} else {
		/* process owner is neither the dir owner nor part of the dir's specified group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{8}wx", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* all have write permissions */
			return
		}
	}
	err = errors.New("The owner of the process for PID " + fmt.Sprint(proc.Pid) + "does not have permissions to create the Agent log file located at " + LogFile)
	return
}

/* need write access to log file */
func canWriteToLog(proc process.Process, LogFile string, logFileCanCreate bool) (err error) {
	/* if the logs file doesn't exist and can't be created, return false */
	if _, errFileExist := os.Stat(LogFile); os.IsNotExist(err) && !logFileCanCreate {
		err = errFileExist
		return
	} else if _, errFileExist := os.Stat(LogFile); os.IsNotExist(errFileExist) && logFileCanCreate {
		/* if log file doesn't exist but can be created then return true */
		return
	} else {
		procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID := getFileAndProcUIDsGIDs(proc, LogFile)
		if procOwnerUID == "" && procOwnerGID == "" && fileOwnerUID == "" && fileOwnerGID == "" {
			err = errors.New("Permission settings to write to Agent log file are indeterminable")
			return
		}
		info, errLogFileStat := os.Stat(LogFile)
		if errLogFileStat != nil {
			log.Debug(taskName, "Error determining log file permissions. Error is: ", errLogFileStat)
			err = errLogFileStat
			return
		}
		filePerm := info.Mode().Perm()
		/* the process owner also is the directory owner */
		if procOwnerUID == fileOwnerUID {
			matched, errRegexMatch := regexp.MatchString("[r-]{2}w.+", filePerm.String())
			if errRegexMatch != nil {
				log.Debug(taskName, errRegexMatch)
				err = errRegexMatch
				return err
			}
			if matched {
				/* the process/dir owner has write permissions */
				return
			}
		} else if procOwnerGID == fileOwnerGID {
			/* the process owner is part of the file group */
			matched, errRegexMatch := regexp.MatchString("[rwx-]{5}w.+", filePerm.String())
			if errRegexMatch != nil {
				log.Debug(taskName, err)
				err = errRegexMatch
				return err
			}
			if matched {
				/* the file group has write permissions */
				return
			}
		} else {
			/* process owner is neither the file owner nor part of the file's specified group */
			matched, errRegexMatch := regexp.MatchString("[rwx-]{8}w.+", filePerm.String())
			if errRegexMatch != nil {
				log.Debug(taskName, err)
				err = errRegexMatch
				return err
			}
			if matched {
				/* all have write permissions */
				return
			}
		}
	}
	err = errors.New("The owner of the process does not have permissions to write to the Agent log file")
	return
}

/* need write/execute access to temp dir */
func canCreateFilesInTempDir(proc process.Process, TempDir string) (err error) {
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID := getFileAndProcUIDsGIDs(proc, TempDir)
	if procOwnerUID == "" && procOwnerGID == "" && fileOwnerUID == "" && fileOwnerGID == "" {
		err = errors.New("Permission settings to create temporary files in " + TempDir + " are indeterminable for PID " + fmt.Sprint(proc.Pid))
		return
	}
	info, errStatTempDir := os.Stat(TempDir)
	if errStatTempDir != nil {
		log.Debug(taskName, "Error creating temp files. Error is: ", errStatTempDir)
		err = errStatTempDir
		return
	}
	filePerm := info.Mode().Perm()
	/* the process owner also is the directory owner */
	if procOwnerUID == fileOwnerUID {
		matched, errRegexMatch := regexp.MatchString("[drwx-]{2}wx.+", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, errRegexMatch)
			err = errRegexMatch
			return
		}
		if matched {
			/* the process/dir owner has write permissions */
			return
		}
	} else if procOwnerGID == fileOwnerGID {
		/* the process owner is part of the dir group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{5}wx.+", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, err)
			err = errRegexMatch
			return
		}
		if matched {
			/* the dir group has write permissions */
			return
		}
	} else {
		/* process owner is neither the dir owner nor part of the dir's specified group */
		matched, errRegexMatch := regexp.MatchString("[drwx-]{8}wx", fmt.Sprint(filePerm))
		if errRegexMatch != nil {
			log.Debug(taskName, err)
			err = errRegexMatch
			return
		}
		if matched {
			/* all have write permissions */
			return
		}
	}
	err = errors.New("The owner of the process for PID " + fmt.Sprint(proc.Pid) + "does not have permissions to create the necessary temporary files in " + TempDir)
	return
}

func getFileAndProcUIDsGIDs(proc process.Process, file string) (procOwnerUID string, procOwnerGID string, fileOwnerUID string, fileOwnerGID string) {
	procOwner, _ := proc.Username()
	procOwnerUser, err := user.Lookup(procOwner)
	if err != nil {
		log.Debug(taskName, "Problem with Lookup. Error is: ", err)
		return
	}
	procOwnerUID = procOwnerUser.Uid
	procOwnerGID = procOwnerUser.Gid
	info, err := os.Stat(file)
	if err != nil {
		log.Debug(taskName, "Problem with os.Stat while determining process/file permissions for PID "+fmt.Sprint(proc.Pid)+". Error is: ", err)
		return
	}
	fileModeInfo := info.Sys()
	fileOwner := fileModeInfo.(*syscall.Stat_t)
	fileOwnerUID = fmt.Sprint(fileOwner.Uid)
	fileOwnerGID = fmt.Sprint(fileOwner.Gid)
	return
}
