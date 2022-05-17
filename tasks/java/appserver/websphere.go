package appserver

import (
	"errors"
	"strconv"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var (
	// files containing websphere version information
	keyFileNames = []string{"WAS.product", "platform.websphere", "version.txt"}
	// regex strings to match within key files
	stringsToMatch = []string{"<version>([0-9.]*).*", ".*version=\"([0-9.]*).*", "Version:[ \t]*([0-9.]*)"}
	// used for logging
	websphereTaskName = "Java/Appserver/Websphere - "
	unknown           = "Unknown"
)

type JavaAppServerWebSphere struct {
	osGetwd             osFunc
	osGetExecutable     osFunc
	fileFinder          func([]string, []string) []string
	versionIsCompatible tasks.VersionIsCompatibleFunc
	findStringInFile    tasks.FindStringInFileFunc
	returnSubstring     tasks.ReturnStringInFileFunc
}

type osFunc func() (string, error)

type WebspherePayload struct {
	Name    string
	Version string
	Status  string
}

func (p JavaAppServerWebSphere) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/AppServer/WebSphere")
}

func (p JavaAppServerWebSphere) Explain() string {
	return "Check Websphere AS version compatibility with New Relic Java agent"
}

func (p JavaAppServerWebSphere) Dependencies() []string {
	return []string{}
}

// Execute - this function iterates through the working and nrdiag exeutable dirs, searches for key filenames
// of files which should determine version information for websphere else it declares websphere not present
func (p JavaAppServerWebSphere) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	dirs, getDirsErr := p.getDirs()
	if getDirsErr != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Task unable to complete. Errors encountered while determining path to the executable and working dir.",
		}
	}
	// obtain path to file(s) named following the patterns of the keyFileNames
	filesToPeruse := p.fileFinder(keyFileNames, dirs)
	if len(filesToPeruse) == 0 {
		//base case - no websphere files found
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Websphere was not detected in this environment.",
		}
	}

	// at least one file was found; iterate through file(s) and regex match stringsToMatch
	payload := p.searchFilesForVersion(filesToPeruse)
	return p.determineResult(payload)
}

func (p JavaAppServerWebSphere) getDirs() ([]string, error) {
	websphereHomeDir, getWorkDirErr := p.osGetwd()
	var dirs []string
	if getWorkDirErr != nil {
		// problem obtaining the current working directory location
		log.Debug(websphereTaskName + "Error encountered while determining the current working directory. Error is " + getWorkDirErr.Error())
	} else {
		dirs = append(dirs, websphereHomeDir)
	}
	nrdiagExeDir, getExeErr := p.osGetExecutable()
	if getExeErr != nil {
		// problem obtaining the nrdiag executable directory location
		log.Debug(websphereTaskName + "Error encountered while determining the nrdiag executable directory. Error is " + getExeErr.Error())
		if getWorkDirErr != nil {
			// we can get neither directory - therefore we'll have no where to search for websphere files
			return []string{}, errors.New("obtained neither the current working directory nor the executable directory location")
		}
	} else {
		dirs = append(dirs, nrdiagExeDir)
	}
	return dirs, nil
}

func (p JavaAppServerWebSphere) determineResult(payload WebspherePayload) (result tasks.Result) {

	// we found a potential websphere file but did not determine the version - warning
	if payload.Version == unknown {
		payload.Status = unknown
		result.Status = tasks.Warning
		result.Payload = payload
		result.URL = "https://docs.newrelic.com/docs/agents/java-agent/additional-installation/ibm-websphere-application-server"
		result.Summary = "We suspect this is a WebSphere environment but we're unable to determine the version. Supported status is unknown."
		return
	}
	// this will set payload status to supported, unsupported, or unknown

	// supported - WebSphere 7.0 to 9.x as of Java agent 3.47.0
	isSupported, versionSupportError := p.versionIsCompatible(payload.Version, []string{"7.0-9"})

	log.Debugf("isSupported: %v", isSupported)
	log.Debugf("versionSupportError: %v", versionSupportError)

	if versionSupportError != nil {
		result.Status = tasks.Error
		result.Payload = payload
		result.URL = "https://docs.newrelic.com/docs/agents/java-agent/additional-installation/ibm-websphere-application-server"
		result.Summary = "Websphere and its version were found in this environment but we're unable to determine its supported status."
		return
	}
	//result = setStatusAndSummary(result, payload)
	if isSupported {
		result.Status = tasks.Success
		result.Summary = "Supported version of Websphere detected in this environment."
	} else {
		result.Status = tasks.Failure
		result.Summary = "Unsupported version of Websphere detected in this environment."
	}
	result.Payload = payload
	return
}

func (p JavaAppServerWebSphere) searchFilesForVersion(files []string) (payload WebspherePayload) {
	payload.Version = unknown
	log.Debug(websphereTaskName + "Found " + strconv.Itoa(len(files)) + " potential Websphere file(s).")
Loop:
	for _, file := range files {
		for _, stringToMatch := range stringsToMatch {
			// vague file name, double check it's definitely Websphere 7's version.txt
			if strings.Contains(file, "version.txt") {
				// if this version.txt is not Websphere's, don't check for the version
				if !p.findStringInFile("WebSphere", file) {
					log.Debug(websphereTaskName + "Found a version.txt but it does not appear to be WebSphere's.")
					continue
				}
			}
			version := ""
			versionRaw, strFindErr := p.returnSubstring(stringToMatch, file)
			// file was not found or does not contain substring
			if strFindErr != nil {
				if strings.Contains(strFindErr.Error(), "string not found in file") {
					// do nothing - we expect only one string in stringsToMatch will be found, so this could be common
				} else {
					/* this is also logged in ReturnSubstringInFile but we want to associate this error with this task */
					log.Debug(websphereTaskName + "Error finding websphere version in " + file + ". Error is: " + strFindErr.Error())
				}
			} else {
				for _, v := range versionRaw {
					if !strings.Contains(v, "ersion") {
						version = v
						break
					}
				}
				log.Debug(websphereTaskName + "Found Websphere version in " + file + ". Version is: " + version)
				payload.Version = version
				break Loop
			}
		}
	}
	return
}
