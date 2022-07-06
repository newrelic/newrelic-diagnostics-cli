package env

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetCoreEnvVersions - This struct defined the sample plugin which can be used as a starting point
type DotNetCoreEnvVersions struct {
	cmdExec tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t DotNetCoreEnvVersions) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/Env/Versions")
}

// Explain - Returns the help text for each individual task
func (t DotNetCoreEnvVersions) Explain() string {
	return "Determine .NET Core version"
}

// Dependencies - Returns the dependencies for ech task.
func (t DotNetCoreEnvVersions) Dependencies() []string {
	return []string{
		"Base/Env/CollectEnvVars",
	}
}

type dotnetCoreVersionSource int

const (
	dotnetInfo dotnetCoreVersionSource = iota
	dotnetVersion
	dotnetDir
)

// Execute - The core work within each task
func (t DotNetCoreEnvVersions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Base/Env/CollectEnvVars"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No environment variables found. This task did not run",
		}
	}

	// Gather env variables from upstream
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	versions, err := t.checkVersions(envVars)

	if len(versions) < 1 {
		errorMessage := "Unable to complete this health check because we ran into some unexpected errors when attempting to collect this application's .NET Core version:\n"
		errorMessage += err
		return tasks.Result{
			Status:  tasks.Error,
			Summary: errorMessage,
		}
	}
	return tasks.Result{
		Status:  tasks.Info,
		Summary: strings.Join(versions, ", "),
		Payload: versions,
	}
}

func (t DotNetCoreEnvVersions) checkVersions(envVars map[string]string) ([]string, string) {
	source := 0
	errorMessages := ""
	versions, errMessage, keepGoing := t.checkVersionWithSource(envVars, source)
	for keepGoing {
		source++
		errorMessages += errMessage
		versions, errMessage, keepGoing = t.checkVersionWithSource(envVars, source)
	}
	errorMessages += errMessage

	return versions, errorMessages
}

func (t DotNetCoreEnvVersions) checkVersionWithSource(envVars map[string]string, source int) ([]string, string, bool) {
	switch dotnetCoreVersionSource(source) {
	case dotnetInfo:
		infoOutput, err := t.cmdExec("dotnet", "--info")
		if err != nil {
			return []string{}, fmt.Sprint("Unable to run 'dotnet --info':\n%w\n", err), true
		}

		versions, parseErr := t.parseDotnetInfoOutput(string(infoOutput))

		return versions, parseErr, len(versions) < 1

	case dotnetVersion:
		version, err := t.cmdExec("dotnet", "--version")
		if err != nil {
			return []string{}, fmt.Sprint("Unable to run 'dotnet --version':\n%w\n", err), true
		}

		stringVersion := string(version)

		if strings.TrimSpace(stringVersion) == "" {
			return []string{}, "Unable to determine version with `dotnet --version`\n", true
		}

		return []string{stringVersion}, "", false

	case dotnetDir:
		versions, err := t.getVerFromDir(envVars)
		if err != "" {
			return []string{}, err, true
		}

		return versions, "", false

	default:
		return []string{}, "Unable to determine .NET core version using any method\n", false
	}
}

func (t DotNetCoreEnvVersions) parseDotnetInfoOutput(output string) ([]string, string) {
	versions := []string{}
	uniqueChecker := make(map[string]struct{})
	r, _ := regexp.Compile(`\d+\.\d+\.\d+\.?\d?`)
	sections := []string{".NET SDKs installed:", ".NET runtimes installed:"}
	outputSplit := strings.Split(output, "\n")
	if len(outputSplit) < 1 {
		return versions, "Unable to determine installed .NET core versions using `dotnet --info` output"
	}
	inSection := false

	for n := 0; n < len(sections); n++ {
		section := sections[n]
		for i := 0; i < len(outputSplit); i++ {
			line := outputSplit[i]
			if strings.TrimSpace(line) == "" {
				inSection = false
			}
			if inSection {
				version := strings.TrimSpace(r.FindString(line))
				if _, dupe := uniqueChecker[version]; !dupe {
					versions = append(versions, version)
					uniqueChecker[version] = struct{}{}
				}
			}
			if strings.EqualFold(strings.TrimSpace(line), strings.TrimSpace(section)) {
				inSection = true
			}
		}
	}

	return versions, ""
}

func (t DotNetCoreEnvVersions) getVerFromDir(envVars map[string]string) ([]string, string) {
	errorMessage := ""
	versions := []string{}
	// check dirs
	dirsToRead, errRead := t.buildDirsToRead(envVars)

	if errRead != nil {
		errorMessage += fmt.Sprint("Unable to find version in dotnet sdk path:\n%w\n", errRead)
		return []string{}, errorMessage
	}

	logger.Debug("DotNetCoreVersions - dirs to read:", dirsToRead)

	for _, directory := range dirsToRead {
		logger.Debug("DotNetCoreVersions - Checking ", directory)
		subDirs, err := ioutil.ReadDir(directory)

		if err != nil {
			if os.IsNotExist(err) {
				continue // don't care about this error
			}
			logger.Debug("DotNetCoreVersions - Error reading '", directory, "'. Error: ", err.Error())
			errorMessage += fmt.Sprint("Unable to read from dotnet sdk path:\n%w\n", err)
			continue // go to the next directory
		}

		r, _ := regexp.Compile(`^\d+[.]\d+[.]\d+$`)
		for _, dir := range subDirs {
			dirName := dir.Name()
			versionMatch := r.MatchString(dirName) // ensure we just have the version dirs and not nuget's
			//logger.Debug(dirName)
			if versionMatch { // filter out NuGet's dirs
				versions = append(versions, dirName)
			}
		}
	}

	return versions, errorMessage
}

func (t DotNetCoreEnvVersions) buildDirsToRead(envVars map[string]string) (dirsToRead []string, retErr error) {
	var (
		searchString    string
		pathVarSplit    []string
		netCoreLocWin   = []string{`C:\Program Files\dotnet\sdk`}
		netCoreLocLinux = []string{`/usr/share/dotnet/sdk`, `/usr/local/share/dotnet/sdk`}
	)

	switch os := runtime.GOOS; os {
	case "windows":
		dirsToRead = netCoreLocWin
		dirsToRead = appendIfUnique(dirsToRead, filepath.Join(envVars["LOCALAPPDATA"], "Microsoft", "dotnet", "sdk"))
		pathVarSplit = strings.Split(envVars["PATH"], ";")
		searchString = `\dotnet`
	case "darwin", "linux":
		dirsToRead = netCoreLocLinux
		if envVars["HOME"] != "" {
			dirsToRead = appendIfUnique(dirsToRead, filepath.Join(envVars["HOME"], ".dotnet", "sdk"))
		}
		pathVarSplit = strings.Split(envVars["PATH"], ":")
		searchString = `/dotnet`
	default:
		logger.Debug("DotNetCoreVersions error, Unknown OS: ", os)
		return nil, errors.New("task DotNetCoreVersions encountered unknown OS: " + os)
	}
	if envVars["DOTNET_INSTALL_PATH"] != "" {
		dirsToRead = appendIfUnique(dirsToRead, filepath.Join(envVars["DOTNET_INSTALL_PATH"], "sdk"))
	}

	fromPathVar := t.checkPathVarForDotnet(pathVarSplit, searchString)
	if fromPathVar != "" {
		dirsToRead = appendIfUnique(dirsToRead, filepath.Join(fromPathVar, "/sdk"))
	}

	return dirsToRead, nil
}

func (t DotNetCoreEnvVersions) checkPathVarForDotnet(paths []string, searchString string) string {
	if len(paths) > 0 {
		for _, path := range paths {
			if strings.Contains(path, searchString) {
				logger.Debug("DotNetCoreVersions - Found dotnet core dir in path var: ", path)
				return path
			}
		}
	}
	logger.Debug("DotNetCoreVersions - Did not find dotnet core dir in path var.")
	return ""
}

func appendIfUnique(dirsToRead []string, path string) []string {
	for _, directory := range dirsToRead {
		if directory == path {
			return dirsToRead
		}
	}
	return append(dirsToRead, path)
}
