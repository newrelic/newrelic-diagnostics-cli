package appserver

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// JavaAppserverJbossEapCheck - This struct defined the Jboss EAP check
type JavaAppserverJbossEapCheck struct {
	runtimeOs             string
	fileExists            tasks.FileExistsFunc
	returnSubstringInFile tasks.ReturnStringInFileFunc
	findFiles             func([]string, []string) []string
	listDir               listDirType
}

// Identifier - This returns the Category, Subcategory and Name of this task
func (p JavaAppserverJbossEapCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Appserver/JbossEapCheck")
}

// Explain - Returns the help text for this task
func (p JavaAppserverJbossEapCheck) Explain() string {
	return "Check JBoss EAP version compatibility with New Relic Java agent"
}

// Dependencies - Returns the dependencies for this task.
func (p JavaAppserverJbossEapCheck) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within this task
func (p JavaAppserverJbossEapCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	result.URL = "https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"

	envVars, _ := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)

	if upstream["Java/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Java config file not detected, this task didn't run"
		return result
	}
	//1. Find version file
	// If env var is set, retrieve from there
	// else look in a bunch of different places
	//2. Read version file
	//3. Verify version

	jBossHome := envVars["JBOSS_HOME"]
	var jbossConfFile string

	//Check for JBOSS version file in environment variable
	if jBossHome != "" {
		jbossConfFile = filepath.Join(jBossHome, "version.txt")
	} else {
		var fileGetterErr error
		if p.runtimeOs == "windows" {
			jbossConfFile, fileGetterErr = p.getJbossFileWindows(envVars["Program_files"])
		} else if p.runtimeOs == "linux" {
			jbossConfFile, fileGetterErr = p.getJbossFileLinux()
		} else {
			result.Status = tasks.None
			result.Summary = "OS not detected as Linux or Windows, not running."
			return result
		}
		log.Debug("Error returned fileGetterErr:", fileGetterErr)

		if fileGetterErr != nil {
			if fileGetterErr.Error() == "Unable to detect JBoss directory" {
				result.Status = tasks.None
				result.Summary = "Can't find version file, didn't exist"
				return result
			}
			result.Status = tasks.Error
			result.Summary = "Error getting jboss version file: " + fileGetterErr.Error()
			return result
		}

	}

	versionString, err := p.readJbossVersionFile(jbossConfFile)
	log.Debug("error getting versionString: ", err)
	if err != nil {
		if err.Error() == "Version file didn't exist" {
			result.Status = tasks.Warning
			result.Summary = "JBossEAP detected but unable to detect version: Version file not found"
			return result
		}
		log.Debugf("Error reading version file: %v", err)
		result.Status = tasks.Error
		result.Summary = "JBossEAP detected but unable to detect version: Error reading version file"
		return result
	}

	result.Status, result.Summary = versionCheckJboss(versionString)

	return result
}

func (p JavaAppserverJbossEapCheck) getJbossFileWindows(programFiles string) (string, error) {
	//For versions above 6.1 this is required per Red Hat docs to set as a service
	if programFiles != "" {

		programFilesDirs := p.listDir(programFiles, ioutil.ReadDir)
		if len(programFilesDirs) == 0 {
			return "", errors.New("unable to list directory")
		}

		for _, f := range programFilesDirs {
			if strings.Contains(f, "EAP-") {
				eapDir := filepath.Join(programFiles, f)
				subDirs := p.listDir(eapDir, ioutil.ReadDir)
				for _, sf := range subDirs {
					if strings.Contains(sf, `jboss-eap-`) {
						pathEapDir := filepath.Join(eapDir, sf, "version.txt")
						return pathEapDir, nil
					}
				}

			}
		}

	}
	return "", errors.New("Unable to detect JBoss directory")
}

type listDirType func(string, readDirFunc) []string
type readDirFunc func(string) ([]os.FileInfo, error)

func listDir(directory string, readDir readDirFunc) []string {
	subDirs, subError := readDir(directory)
	if subError != nil {
		return []string{}
	}
	var directoryList []string
	for _, file := range subDirs {
		directoryList = append(directoryList, file.Name())
	}
	return directoryList
}

func (p JavaAppserverJbossEapCheck) readJbossVersionFile(versionFilePath string) (string, error) {

	if p.fileExists(versionFilePath) {

		returnedString, err := p.returnSubstringInFile("Enterprise Application Platform - Version ([0-9.]+.*)", versionFilePath)
		if err != nil {
			log.Debugf("error reading version file. Error: %v", err)
			return "", err
		} else if len(returnedString) > 1 {
			return returnedString[1], nil
		}

	}
	return "", errors.New("Version file didn't exist")
}

func (p JavaAppserverJbossEapCheck) getJbossFileLinux() (string, error) {

	//Default to other methods to see if JBoss files are present and point to home path

	jbossConfigFilePath := `/etc/default/jboss-eap.conf`

	//Check if the file exists by trying to get stats on it, and looking at returned error

	if p.fileExists(jbossConfigFilePath) {
		jbossHomeFull, err := p.returnSubstringInFile("JBOSS_HOME=.*", jbossConfigFilePath)
		log.Debug("JbossHomeFull: ", jbossHomeFull)
		if err != nil {
			return "", errors.New("file doesn't contain JBOSS_HOME")
		}
		if len(jbossHomeFull) > 0 {
			jbossHomePathSplit := strings.Split(jbossHomeFull[0], "\"")
			if len(jbossHomePathSplit) < 2 {
				return "", errors.New("Error parsing JBOSS_HOME path in jboss-eap.conf")
			}
			jbossHomePath := jbossHomePathSplit[1]
			return filepath.Join(jbossHomePath, "version.txt"), nil
		}
	}

	//Check default rpm install for version 7+
	default7InstallVersionLoc := `/opt/rh/eap7/root/usr/share/wildfly/version.txt`

	if p.fileExists(default7InstallVersionLoc) {
		return default7InstallVersionLoc, nil
	}

	//Check a few more places where we could find a file with the JBoss Home path, per Red Hat doc
	path := []string{`/etc/init.d`, `etc/jboss-as`, `/etc/rc.d/init.d`, `/etc/rc.d/init.d/jboss`, `/etc/jbossas`, `/etc/sysconfig`, `/etc/sysconfig`, `/opt/rh/eap7/root/usr/share/wildfly`, `/etc/default/jboss-eap`}

	files := []string{`jboss-as-standalone.sh`, `jboss-as.conf`, `jboss_init_redhat.sh`, `jboss_init_suse.sh`, `jboss_init_.*`, `jbossas.conf`, `jbossas`, `jbossas-domain`, `jboss-eap-rhel.sh`, `jboss-eap.conf`}

	fileResults := p.findFiles(files, path)

	if len(fileResults) < 1 {
		return "", errors.New("Unable to detect JBoss directory")
	}

	for _, v := range fileResults {
		jbossHomeFull, err := p.returnSubstringInFile("JBOSS_HOME=.*", v)
		if err == nil && len(jbossHomeFull) > 0 {
			jbossHomePathSplit := strings.Split(jbossHomeFull[0], "\"")
			if len(jbossHomePathSplit) < 2 {
				return "", errors.New("error parsing JBOSS_HOME path in jboss-eap.conf")
			}
			jbossHomePath := jbossHomePathSplit[1]
			return filepath.Join(jbossHomePath, "version.txt"), nil
		}
	}

	return "", errors.New("Unable to detect JBoss directory")

}

func versionCheckJboss(version string) (status tasks.Status, summary string) {
	compatibilityRequirements := []string{"6-6.*"}
	versionSplit := strings.Split(version, ".")
	var err error
	isCompatible := false
	if len(versionSplit) > 0 {
		//String coming in looks something like `6.1.2.GA` but we only care about the major version
		isCompatible, err = tasks.VersionIsCompatible(versionSplit[0], compatibilityRequirements)

	}

	if err != nil {
		status = tasks.Error
		summary = "Error converting version string to int. String: " + version + " Error: " + err.Error()
		return
	}

	if isCompatible {
		status = tasks.Success
		summary = "Jboss EAP version is compatible. Version is " + version
		return
	}
	status = tasks.Failure
	summary = "Jboss EAP detected as incompatible. Detected version was " + version

	return

}
