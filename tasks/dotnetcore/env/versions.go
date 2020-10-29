package env

import (
	"os"
	"errors"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
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
	var result tasks.Result

	// Gather env variables from upstream
	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the map[string]string I know it should return
	if !ok {
		logger.Debug("DotNetCoreVersions - Error gathering Environment Variables from upstream.")
	} else {
		logger.Debug("DotNetCoreVersions - Successfully gathered Environment Variables from upstream.")
	}

	versions, checkVersionErrorCount, checkVersionsErrorDetails := checkVersions(envVars)

	if versions == nil || len(versions) < 1 {
		if checkVersionsErrorDetails != nil || checkVersionErrorCount > 0 {
			logger.Debug("DotNetCoreVersions - Error determining the version .NET Core installed. There were", checkVersionErrorCount, "errors. Please check previous log entries for other errors. The last error seen was: ", checkVersionsErrorDetails.Error())
			result.Status = tasks.Error
			result.Summary = "Error determining the version .NET Core installed."
			return result
		}

		logger.Debug("DotNetCoreVersions - No .NET Core version information found.")
		result.Status = tasks.None
		result.Summary = ".NET Core is not installed."
		return result
	}

	if checkVersionErrorCount > 0 {
		logger.Debug("DotNetCoreVersions - There were", checkVersionErrorCount, "errors, but also found", strconv.Itoa(len(versions)), ".NET Core versions installed. Marking successful, not reporting errors.")
	}

	if checkVersionsErrorDetails != nil {
		logger.Debug("DotNetCoreVersions - There was an error, but also found", strconv.Itoa(len(versions)), ".NET Core versions installed. Marking successful, not reporting errors.")
		logger.Debug("DotNetCoreVersions - Error was:", checkVersionsErrorDetails.Error())
	}

	result.Summary = strings.Join(versions, ", ")
	result.Status = tasks.Info
	result.Payload = versions
	return result
}

func checkVersions(envVars map[string]string) (versions []string, countErrors int, retErr error) {
	countErrors = 0

	// first check if version is in env vars
	versionFromEnvVars := envVars["DOTNET_SDK_VERSION"]
	if versionFromEnvVars != "" {
		logger.Debug("DotNetCoreVersions - found .NET Core version in DOTNET_SDK_VERSION Env Var")
		return append(versions, versionFromEnvVars), countErrors, retErr
	}

	// check dirs
	dirsToRead, retErr := buildDirsToRead(envVars)
	if retErr != nil {
		countErrors++
		return nil, countErrors, retErr
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
			retErr = err  // keep track of the last error
			countErrors++ // keep track of how many errors we encounter
			continue      // go to the next directory
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

	return versions, countErrors, retErr
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
