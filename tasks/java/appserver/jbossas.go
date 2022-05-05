package appserver

import (
	"errors"
	"regexp"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/shirou/gopsutil/process"
)

// JavaAppserverJBossAsCheck - This struct defines the JBoss AS version check
type JavaAppserverJBossAsCheck struct {
	getCmdline            getCmdlineFromProcessFunc
	findFiles             func([]string, []string) []string
	findProcessByName     tasks.FindProcessByNameFunc
	returnSubstringInFile tasks.ReturnStringInFileFunc
}
type getCmdlineFromProcessFunc func(process.Process) string

// Identifier - This returns the Category, Subcategory and Name of this task
func (p JavaAppserverJBossAsCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Appserver/JBossAsCheck")
}

// Explain - Returns the help text for this task
func (p JavaAppserverJBossAsCheck) Explain() string {
	return "Check JBoss AS version compatibility with New Relic Java agent"
}

// Dependencies - Returns the dependencies for this task.
func (p JavaAppserverJBossAsCheck) Dependencies() []string {
	return []string{"Base/Env/CollectEnvVars", "Java/Env/Process"}
}

// Execute - The core work within this task
func (p JavaAppserverJBossAsCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	if upstream["Java/Env/Process"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Java/Env/Process did not pass our validation. This task did not run.", //This is a major java task that users should make it succeeds prior to worrying about any other issues.
		}
	}

	if upstream["Base/Env/CollectEnvVars"].Status == tasks.Info {
		envVars, _ := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)

		jBossAsHome := envVars["JBOSS_HOME"]

		if jBossAsHome != "" {
			log.Debug("JBOSS_HOME env variable set")
			versionString, err := p.getAndParseJBossAsReadMe(jBossAsHome)
			if err != nil {
				result.Summary = tasks.ThisProgramFullName + " was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading jboss readme: " + err.Error() + "\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"
				result.Status = tasks.Error
				return result
			}
			result.Summary, result.Status = p.checkJBossAsVersion(versionString)
			return result
		}
	}

	//JBOSS env var is not present, we'll attempt to find Jboss as a java process argument
	processes, err := p.findProcessByName("java")

	if err != nil {
		log.Debug("Error reading processes. Error: ", err.Error())
		result.Summary = tasks.ThisProgramFullName + " was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading from your java process: " + err.Error() + "\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"
		result.Status = tasks.Error
		return result
	}

	for _, n := range processes {

		cmdLine := p.getCmdline(n)
		log.Debugf("cmdLine: %v", cmdLine)
		if strings.Contains(cmdLine, "jboss.home.dir=") {
			homeDir := p.getHomeDirFromCmdline(cmdLine)
			versionString, err := p.getAndParseJBossAsReadMe(homeDir)

			if err != nil {
				result.Summary = tasks.ThisProgramFullName + " was unable to validate if your JBoss AS version is compatible with New Relic Java agent because it ran into an error when reading processes from homedir: " + err.Error() + "\nYou can take look at this documentation to verify if your version of JBoss is compatible: https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#app-web-servers"
				result.Status = tasks.Error
				return result
			}

			result.Summary, result.Status = p.checkJBossAsVersion(versionString)
			return result
		}
	}

	//fallthrough case
	result.Summary = "Could not find JBoss AS Home Path. Assuming JBoss AS is not installed"
	result.Status = tasks.None
	return result
}

func getCmdlineFromProcess(proc process.Process) string {
	cmdline, _ := proc.Cmdline()
	return cmdline
}

func (p JavaAppserverJBossAsCheck) getHomeDirFromCmdline(cmdLine string) string {
	pat := regexp.MustCompile(`jboss[.]home[.]dir=[\]\s[a-zA-Z0-9\S]+`)
	homeDirRaw := pat.FindString(cmdLine)
	log.Debugf("homeDirRaw: %v\n", homeDirRaw)
	homeDirSplit := strings.SplitAfter(homeDirRaw, "=")
	log.Debugf("homeDirSplit: %v \n", homeDirSplit)

	if len(homeDirSplit) > 1 {
		homeDir := homeDirSplit[1]
		log.Debugf("homeDir: %v\n", homeDir)
		return homeDir
	}
	return ""
}

func (p JavaAppserverJBossAsCheck) getAndParseJBossAsReadMe(homePath string) (string, error) {

	versionPat := regexp.MustCompile("[0-9]+[.][0-9]+[.][0-9]+")
	versionInHomePath := versionPat.FindString(homePath)
	if len(versionInHomePath) > 0 {
		return versionInHomePath, nil
	}
	filePatterns := []string{"README.txt", "readme.html"}
	paths := []string{homePath}

	readmes := p.findFiles(filePatterns, paths)
	if len(readmes) < 1 {
		log.Debug("Error finding JBoss readme at ", homePath)
		return "", errors.New("error finding JBoss version")
	}

	versionSearchString := "JBoss[ Application Server ]+[0-9]+[.][0-9]+[.][0-9]+"
	//there should only be one read me in the Jboss folder, but it can be .txt or .html depending on version
	versionStringRaw, err := p.returnSubstringInFile(versionSearchString, readmes[0])
	if len(versionStringRaw) < 1 || err != nil {
		log.Debug("Could not find version string in readme.txt file")
		return "", errors.New("error finding version string")
	}

	versionString := ""

	for _, v := range versionStringRaw {
		versionString = versionPat.FindString(v)
		if versionString != "" {
			break
		}
	}

	if strings.Count(versionString, ".") < 2 {
		return "", errors.New("error finding version string")
	}

	return versionString, nil

}

// check Version and construct result summary/status
func (p JavaAppserverJBossAsCheck) checkJBossAsVersion(versionString string) (string, tasks.Status) {
	versionRequirements := []string{
		"4.0.5-7.*", // big upper boundary for minor version should catch all 7.x versions
	}
	jBossCompatible, err := tasks.VersionIsCompatible(versionString, versionRequirements)

	if err != nil {
		return "Error parsing detected version string: " + versionString, tasks.Error
	}
	if jBossCompatible {
		return "JBoss version supported. Version is " + versionString, tasks.Success
	}

	return "Unsupported version of JBoss AS detected. Supported versions are 4.0.5 to AS 7.x. Detected version is " + versionString, tasks.Failure

}
