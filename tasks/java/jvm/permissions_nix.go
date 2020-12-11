// +build linux darwin

package jvm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	baseLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/env"
	"github.com/shirou/gopsutil/process"
)

var (
	jarPathSysPropSource    = "-javaagent:"
	logPathEnvVarSource     = "NEW_RELIC_LOG" //The unqualified log file name(not a path) or the string STDOUT which will log to standard out. The latter would inmediately give a "permission denied" error so no much need to troubleshoot for this option
	logPathConfigFileSource = "log_file_path" //the directory must already exist if specified
	logPathSysPropSource    = "-Dnewrelic"
	logPathDefaultLocation  = "/logs/newrelic_agent.log"
	nrTempDirKey            = "-Dnewrelic.tempdir" //can only be set as sys prop not as an env var neither the config file
	javaTmpDirKey           = "-Djava.io.tmpdir"   //On UNIX systems the default value of this property can be "/tmp" or "/var/tmp"; on Windows "c:\temp". A different path value may be given through this system property
	ubuntuTmpDirPath        = "/tmp/"              //On linux Ubuntu
	//expected permissions users must have to get the Java agent configuration to work:
	fileOwnerPermissionsRgx          = "-r.+"
	fileGroupPermissionsRgx          = "[rwx-]{4}r.+"
	filePublicPermissionsRgx         = "[rwx-]{7}r.+"
	dirOwnerPermissionsRgx           = "[drwx-]{2}wx.+"
	dirGroupPermissionsRgx           = "[drwx-]{5}wx.+"
	dirPublicPermissionsRgx          = "[drwx-]{8}wx"
	errPermissionsCannotBeDetermined = errors.New("Permissions could not be determined")
	errTempDirDoesNotExist           = errors.New("Diagnostics CLi was unable to find the temp directory location for this host")
	errTempDirHasNoJars              = errors.New("Diagnostics CLi was unable to find the New Relic tmp jar files")
	tempDirRecommendation            = "If you are seeing a java.io.IOException in the New Relic logs, we recommend to manually create the tmp directory passing -Djava.io.tmpdir or -Dnewrelic.tempdir as a JVM argument at runtime. The Java Agent needs this temp directory to create temp JAR files"
)

/* JavaAgentPermissions - struct used to construct the eventual result payload */
type JavaAgentPermissions struct {
	PID                int32
	AgentJarCanRead    requirementDescription //The owner of the Java process to which the -javaagent option will be passed must have read permissions for the java agent JAR. This is because Java is what is being executed, and Java needs to read the newrelic.jar
	LogCanCreate       requirementDescription //The process owner requires write/execute permissions for the directory in which the log directory will be created and execute permissions for all parent directories of the log directory so the java process can traverse into the directory and create the java agent log file
	TempFilesCanCreate requirementDescription //The process owner must have write/execute access to the temp directory for the Java process. This may be the default directory for temporary Java files (specified system-wide), or it may be one specific to the process,
}

// PermissionsStatus represents what we know about the level of permissions a user has on a directory or file and if they match our requirements
type PermissionsStatus int

const (
	denied = iota
	granted
	undetermined
)

func (ps PermissionsStatus) String() string {
	return [...]string{"denied", "granted", "undetermined"}[ps]
}

/* struct used to store if permissions are set correctly for a file/dir and if an error was encountered while determining permissions*/
type requirementDescription struct {
	SuccessLevel PermissionsStatus
	ErrorMsg     error
	Source       string
	Value        string
}

type JavaJVMPermissions struct {
}

// Identifier - returns the Category (Agent), Subcategory (Java) and Name (Permissions)
func (p JavaJVMPermissions) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/JVM/Permissions")
}

// Explain - Returns the help text for the customer for this task
func (p JavaJVMPermissions) Explain() string {
	return "Check Java process permissions meet New Relic Java agent requirements"
}

/* Dependencies - Depends on the the SysPropCollect task to get PIDs with corresponding command-line args */
func (p JavaJVMPermissions) Dependencies() []string {
	return []string{
		"Base/Env/CollectSysProps",
		"Base/Log/Copy",
		"Java/Env/Process",
	}
}

/* Execute - This task checks for processes running new relic java agents and determines if permissions are set correctly to run the Agent */
func (p JavaJVMPermissions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	/* if there are no running Java Agents */
	if upstream["Java/Env/Process"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "The New Relic agent has not been added to a running JVM process yet. This task did not run.",
		}
	}

	/* if there is at least one running Java Agent */
	/* obtain slice of java processes from JavaEnvProcess task */
	javaAgentProcs, ok := upstream["Java/Env/Process"].Payload.([]env.ProcIdAndArgs)
	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "We were unable to run this health check due to an internal type assertion error for the task Java/Env/Process",
		}
	}

	failureCount := 0
	failureSummary := ""
	warningCount := 0
	warningSummary := ""
	payloadResult := []*JavaAgentPermissions{}

	for _, process := range javaAgentProcs { //though we expect to find one single process running the new relic agent, it is not un-heard of users running multiple agents in different processes
		var j JavaAgentPermissions
		/* Can the user read the Agent JAR */
		determineJarPermissions(process.Proc, process.JarPath, &j)
		/* Can the user create the log directory/file for the agent */
		determineLogPermissions(process.Proc, process.Cwd, upstream, &j)
		/* Can the user create jar files within the tmp directory */
		determineTmpDirPermissions(process.Proc, upstream, &j)

		payloadResult = append(payloadResult, &j)

		fmt.Println("luces javaAgentPermissions:", j)

		if j.AgentJarCanRead.SuccessLevel == denied || j.LogCanCreate.SuccessLevel == denied || j.TempFilesCanCreate.SuccessLevel == denied {
			failureCount++
			errorsFound := (j.AgentJarCanRead.ErrorMsg).Error() + "\n" + (j.LogCanCreate.ErrorMsg).Error() + "\n" + (j.TempFilesCanCreate.ErrorMsg).Error()
			failureSummary += fmt.Sprintf("The process for the for PID %d did not meet the New Relic Java Agent permissions requirements. Errors found:\n%s", process.Proc.Pid, errorsFound)
		} else if j.AgentJarCanRead.SuccessLevel == undetermined || j.LogCanCreate.SuccessLevel == undetermined || j.TempFilesCanCreate.SuccessLevel == undetermined {
			warningCount++
			if errors.Is(j.TempFilesCanCreate.ErrorMsg, errTempDirHasNoJars) {
				warningSummary += (j.TempFilesCanCreate.ErrorMsg).Error() + "\n" + tempDirRecommendation //this in only undetermined permissions case that is worth giving a warning. The other undetermined cases are due to our program running into error beyond the user's control
			}
		}
	}
	fmt.Println("luces sumaries and counts:", warningCount, warningSummary, failureCount, failureSummary)
	if failureCount > 0 {
		if warningCount > 0 {
			return tasks.Result{
				Status:  tasks.Failure,
				Summary: failureSummary + warningSummary,
				Payload: payloadResult,
				URL:     "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/determine-permissions-requirements-java",
			}
		}
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: failureSummary,
			Payload: payloadResult,
			URL:     "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/determine-permissions-requirements-java",
		}
	}

	if warningCount > 0 {
		if len(warningSummary) > 0 {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: warningSummary,
				Payload: payloadResult,
				URL:     "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/determine-permissions-requirements-java",
			}
		}
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: tasks.ThisProgramFullName + " ran into some unexpected errors and was unable to fully verify that your application meets the permissions requirements for the Java agent. This health check can be ignore unless your are not seeing new relic logs or you are seeing the java.io.IOException error in your logs. More details can be found on the nrdiag-output.json",
			Payload: payloadResult,
			URL:     "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/determine-permissions-requirements-java",
		}
	}
	return tasks.Result{
		Status:  tasks.Success,
		Summary: "All running Java Agent environments meet minimum permissions requirements",
		URL:     "https://docs.newrelic.com/docs/agents/java-agent/troubleshooting/determine-permissions-requirements-java",
	}

}

func determineTmpDirPermissions(proc process.Process, upstream map[string]tasks.Result, javaAgentPermissions *JavaAgentPermissions) {
	var tempDir, tempDirSource string
	//Find location of tempDir in System Properties. New Relic sys prop should take precedence over standard java tmp files directory sys prop
	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		sysPropsProcesses, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if !ok {
			log.Debug("Failed type assertion for Base/Env/CollectSysProps in JavaJVMPermissions task")
		}
		for _, process := range sysPropsProcesses {
			if process.ProcID == proc.Pid {
				nrTempDirVal, isTempDirPresent := process.SysPropsKeyToVal[nrTempDirKey]
				if isTempDirPresent {
					tempDir = nrTempDirVal
					tempDirSource = nrTempDirKey
				} else {
					javaTmpDirVal, isTmpDirPresent := process.SysPropsKeyToVal[javaTmpDirKey]
					if isTmpDirPresent {
						tempDir = javaTmpDirVal
						tempDirSource = javaTmpDirKey
					}
				}
			}
		}
	}
	//if none of those system properties are set, check the operating system's default tmp dir
	if len(tempDir) == 0 {
		tempDir = os.TempDir() //may return a directory that does not exist but soon we'll find out as we check for this directory's permissions
		tempDirSource = "the default tmp directory for the Operation System"
	}
	fmt.Println("luces tempdir:", tempDir)
	//start assigning javaAgentPermissions values to temp directory
	javaAgentPermissions.TempFilesCanCreate.Source = tempDirSource
	javaAgentPermissions.TempFilesCanCreate.Value = tempDir

	err := canCreateFilesInTempDir(proc, tempDir)

	if err != nil {
		javaAgentPermissions.TempFilesCanCreate.ErrorMsg = err
		if errors.Is(err, errPermissionsCannotBeDetermined) {
			javaAgentPermissions.TempFilesCanCreate.SuccessLevel = undetermined
		} else {
			javaAgentPermissions.TempFilesCanCreate.SuccessLevel = denied
		}
	} else {
		//now before we can declare victory, let's peak in that tmp directory. We expect to see at least one temporary jar file name that includes the word "newrelic" in it
		files, err := ioutil.ReadDir(tempDir)
		if err != nil {
			javaAgentPermissions.TempFilesCanCreate.SuccessLevel = undetermined
			javaAgentPermissions.TempFilesCanCreate.ErrorMsg = err
		}
		var foundTmpJar string
		for _, f := range files {
			if strings.Contains(f.Name(), "newrelic") && strings.Contains(f.Name(), ".jar") {
				foundTmpJar = f.Name()
				break
			}
		}
		if len(foundTmpJar) > 0 {
			javaAgentPermissions.TempFilesCanCreate.SuccessLevel = granted
		} else {
			javaAgentPermissions.TempFilesCanCreate.SuccessLevel = undetermined
			javaAgentPermissions.TempFilesCanCreate.ErrorMsg = errTempDirHasNoJars
		}

	}
}

/* need write/execute access to temp dir */
func canCreateFilesInTempDir(proc process.Process, tempDir string) (err error) {
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err := getUIDsGIDs(proc, tempDir)
	fmt.Println("luces cancreateFilesIntempDir:", procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err)
	if err != nil {
		return err
	}
	info, errStatTempDir := os.Stat(tempDir)

	if os.IsNotExist(errStatTempDir) {
		return errTempDirDoesNotExist
	}
	// if errStatTempDir != nil {
	// 	return errStatTempDir
	// }
	filePerm := info.Mode().Perm()
	var matchedPermissions bool
	var errRegexMatch error
	/* the process owner also is the directory owner */
	if procOwnerUID == fileOwnerUID {
		matchedPermissions, errRegexMatch = regexp.MatchString(dirOwnerPermissionsRgx, fmt.Sprint(filePerm))
	} else if procOwnerGID == fileOwnerGID {
		/* the process owner is part of the dir group */
		matchedPermissions, errRegexMatch = regexp.MatchString(dirGroupPermissionsRgx, fmt.Sprint(filePerm))
	} else {
		/* process owner is neither the dir owner nor part of the dir's specified group */
		matchedPermissions, errRegexMatch = regexp.MatchString(dirPublicPermissionsRgx, fmt.Sprint(filePerm))
	}

	if errRegexMatch != nil {
		return fmt.Errorf("Permissions could not be determined for tmp directory at %s. Error is: %w", tempDir, errRegexMatch)
	}
	if !matchedPermissions {
		/* the process/file owner has no write/execute permissions */
		return fmt.Errorf("The owner of the process for PID %d does not have permissions to create the necessary temporary files in %s: %s", proc.Pid, tempDir, filePerm.String())
	}
	/* the process/dir owner has write/execute permissions */
	return nil
}

func determineLogPermissions(proc process.Process, jarDir string, upstream map[string]tasks.Result, javaAgentPermissions *JavaAgentPermissions) {
	logElements, ok := upstream["Base/Log/Copy"].Payload.([]baseLog.LogElement)

	if !ok {
		log.Debug("We ran into an type assertion error for Base/Log/Copy payload in JavaJVMPermissions task")
	}

	var logFilePath, logFilePathSource string
	var isLogStdout bool
	for _, logElement := range logElements {
		configMap := logElement.Source.KeyVals
		for configKey, configVal := range configMap {
			if configKey == logPathEnvVarSource || strings.Contains(configKey, logPathSysPropSource) || configKey == logPathConfigFileSource {
				if configVal == "stdout" {
					isLogStdout = true
					javaAgentPermissions.LogCanCreate.SuccessLevel = undetermined
				}
				logFilePath = logElement.Source.FullPath
				logFilePathSource = logElement.Source.FoundBy
			} else {
				logFilePath = filepath.Join(jarDir, logPathDefaultLocation)
				logFilePathSource = "Default location (directory where newrelic.jar is) in which the New Relic java agent will create its logs directory."
			}
		}
	}
	//assign javaAgentPermissions values for new relic log file
	javaAgentPermissions.LogCanCreate.Source = logFilePathSource
	javaAgentPermissions.LogCanCreate.Value = logFilePath
	javaAgentPermissions.PID = proc.Pid

	if !isLogStdout {
		errLogWriteExec := canCreateAgentLog(proc, logFilePath)
		if errLogWriteExec != nil {
			javaAgentPermissions.LogCanCreate.SuccessLevel = denied
			javaAgentPermissions.LogCanCreate.ErrorMsg = errLogWriteExec
		}
		javaAgentPermissions.LogCanCreate.SuccessLevel = granted
	}
}

/* need write/execute access to log dir */
func canCreateAgentLog(proc process.Process, logFilePath string) (err error) {
	/* logs file already exists; return success */
	if _, errFileExist := os.Stat(logFilePath); errFileExist == nil {
		return nil
	}

	logDir := filepath.Dir(logFilePath)
	/* if logs directory doesn't exist, log file cannot be created */
	if _, errDirExist := os.Stat(logDir); errDirExist != nil {
		return errDirExist
	}

	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err := getUIDsGIDs(proc, logDir)
	fmt.Println("luces cancreateAgentLog:", procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err)
	if err != nil {
		return fmt.Errorf("Permissions could not be determined for Agent log file creation at %s for PID %d: %w", logFilePath, proc.Pid, err)
	}

	info, errStatDir := os.Stat(logDir)
	if errStatDir != nil {
		return fmt.Errorf("Permissions could not be determined for Agent log file at %s for PID %d: %w", logDir, proc.Pid, errStatDir)
	}
	filePerm := info.Mode().Perm()
	var matchedPermissions bool
	var errRegexMatch error
	/* the process owner also is the directory owner */
	if procOwnerUID == fileOwnerUID {
		matchedPermissions, errRegexMatch = regexp.MatchString(dirOwnerPermissionsRgx, fmt.Sprint(filePerm))
	} else if procOwnerGID == fileOwnerGID {
		/* the process owner is part of the dir group */
		matchedPermissions, errRegexMatch = regexp.MatchString(dirGroupPermissionsRgx, fmt.Sprint(filePerm))
	} else {
		/* process owner is neither the dir owner nor part of the dir's specified group */
		matchedPermissions, errRegexMatch = regexp.MatchString(dirPublicPermissionsRgx, fmt.Sprint(filePerm))
	}

	if errRegexMatch != nil {
		return fmt.Errorf("Permissions could not be determined for writing the Agent log file at %s. Error is: %w", logFilePath, errRegexMatch)
	}
	if !matchedPermissions {
		return fmt.Errorf("The owner of the process for PID %d does not have permissions to create the Agent log file located at %s: %s", proc.Pid, logFilePath, fmt.Sprint(filePerm))
	}
	return nil
}

/* need read permissions only; "java" is being executed and it needs to read the JAR */
func canReadAgentJar(proc process.Process, jarLoc string) (err error) {
	/* check if the Java Agent JAR exists */
	if _, errJarNotExist := os.Stat(jarLoc); os.IsNotExist(errJarNotExist) {
		return fmt.Errorf(`Agent JAR does not exist for PID %d. This location is at %s: %w`, proc.Pid, jarLoc, errJarNotExist)
	}
	procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err := getUIDsGIDs(proc, jarLoc)
	fmt.Println("luces 0:", procOwnerUID, procOwnerGID, fileOwnerUID, fileOwnerGID, err)
	if err != nil {
		return err
	}

	info, errGettingJarPerms := os.Stat(jarLoc)
	if errGettingJarPerms != nil {
		return fmt.Errorf("Permissions could not be determined for JAR at %s. Error is: %w", jarLoc, errGettingJarPerms)
	}
	filePerm := info.Mode().Perm()
	fmt.Println("luces 1:", jarLoc, filePerm)

	var matchedPermissions bool
	var errRegexMatch error
	/* the process owner also is the file owner for the Agent JAR */
	if procOwnerUID == fileOwnerUID {
		matchedPermissions, errRegexMatch = regexp.MatchString(fileOwnerPermissionsRgx, filePerm.String())
		/* the process owner is part of the group to which the Agent JAR belongs */
	} else if procOwnerGID == fileOwnerGID {
		matchedPermissions, errRegexMatch = regexp.MatchString(fileGroupPermissionsRgx, filePerm.String())
	} else {
		/* process owner is neither the Agent JAR file owner nor part of the JAR's specified group */
		matchedPermissions, errRegexMatch = regexp.MatchString(filePublicPermissionsRgx, filePerm.String())
	}

	if errRegexMatch != nil {
		return fmt.Errorf("Permissions could not be determined for JAR at %s. Error is: %w", jarLoc, errRegexMatch)
	}
	if !matchedPermissions {
		/* the process/file owner has no read permissions */
		return fmt.Errorf("The owner of the process for PID %d does not have permissions to execute the New Relic Agent Jar located at %s: %s", proc.Pid, jarLoc, filePerm.String())
	}
	return nil
}

func determineJarPermissions(proc process.Process, jarLoc string, javaAgentPermissions *JavaAgentPermissions) {
	err := canReadAgentJar(proc, jarLoc)

	//assign javaAgentPermissions values to Jar

	if err != nil {
		if errors.Is(err, errPermissionsCannotBeDetermined) {
			javaAgentPermissions.AgentJarCanRead.SuccessLevel = undetermined
		} else {
			fmt.Println("luces fatal 1")
			javaAgentPermissions.AgentJarCanRead.SuccessLevel = denied
			fmt.Println("luces fatal 2")

		}
		javaAgentPermissions.AgentJarCanRead.ErrorMsg = err
	} else {
		javaAgentPermissions.AgentJarCanRead.SuccessLevel = granted
	}

	javaAgentPermissions.PID = proc.Pid
	javaAgentPermissions.AgentJarCanRead.Source = jarPathSysPropSource
	javaAgentPermissions.AgentJarCanRead.Value = jarLoc
}

func getUIDsGIDs(proc process.Process, fileOrDirPath string) (string, string, string, string, error) {
	procOwner, _ := proc.Username()
	procOwnerUser, err := user.Lookup(procOwner)
	if err != nil {
		return "", "", "", "", fmt.Errorf("Permissions could not be determined because we ran into a lookup error. Error is: %w", err)
	}
	procOwnerUID := procOwnerUser.Uid
	procOwnerGID := procOwnerUser.Gid
	info, err := os.Stat(fileOrDirPath)
	if err != nil {
		return "", "", "", "", fmt.Errorf("Permissions could not be determined because we ran into a os.Stat error. Error is: %w", err)
	}
	modeInfo := info.Sys()
	fileDirOwner := modeInfo.(*syscall.Stat_t)
	fileDirOwnerUID := fmt.Sprint(fileDirOwner.Uid)
	fileDirOwnerGID := fmt.Sprint(fileDirOwner.Gid)
	return procOwnerUID, procOwnerGID, fileDirOwnerUID, fileDirOwnerGID, nil
}
