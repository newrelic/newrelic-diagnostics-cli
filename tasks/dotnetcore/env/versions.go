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

// Execute - The core work within each task
func (t DotNetCoreEnvVersions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	// Gather env variables from upstream
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the map[string]string I know it should return
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}
	versions, errorMessage := checkVersions(envVars)

	if len(versions) < 1 {
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

func checkVersions(envVars map[string]string) ([]string, string) {
	errorMessage := "Unable to complete this health check because we ran into some unexpected errors when attempting to collect this application's .NET Core SDK version:\n"
	versions := []string{}
	// first check if version is accesible through the cmdline
	//dotnet --version will Display .NET Core SDK version https://docs.microsoft.com/en-us/dotnet/core/tools/dotnet. Ex: 5.0.101
	version, err := tasks.CmdExecutor("dotnet", "--version")

	if err != nil {
		errorMessage += fmt.Sprint("Unable to run 'dotnet --version':\n%w\n", err)
	} else {
		versions = append(versions, string(version))
		return versions, errorMessage
	}

	// check dirs
	dirsToRead, errRead := buildDirsToRead(envVars)

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

		for _, dir := range subDirs {
			dirName := dir.Name()
			versionMatch, _ := regexp.MatchString(`^\d+[.]\d+[.]\d+$`, dirName) // ensure we just have the version dirs and not nuget's
			//logger.Debug(dirName)
			if versionMatch { // filter out NuGet's dirs
				versions = append(versions, dirName)
			}
		}
	}

	return versions, errorMessage
}

func buildDirsToRead(envVars map[string]string) (dirsToRead []string, retErr error) {
	pathVarSplit := []string{}
	var searchString string
	netCoreLocWin := []string{`C:\Program Files\dotnet\sdk`}
	netCoreLocLinux := []string{`/usr/share/dotnet/sdk`, `/usr/local/share/dotnet/sdk`}

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
		return nil, errors.New("DotNetCoreVersions task encountered unknown OS: " + os)
	}
	if envVars["DOTNET_INSTALL_PATH"] != "" {
		dirsToRead = appendIfUnique(dirsToRead, filepath.Join(envVars["DOTNET_INSTALL_PATH"], "sdk"))
	}

	fromPathVar := checkPathVarForDotnet(pathVarSplit, searchString)
	if fromPathVar != "" {
		dirsToRead = appendIfUnique(dirsToRead, filepath.Join(fromPathVar, "/sdk"))
	}

	return dirsToRead, nil
}

func checkPathVarForDotnet(paths []string, searchString string) string {
	if paths != nil {
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
