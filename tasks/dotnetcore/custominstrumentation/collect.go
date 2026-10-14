package custominstrumentation

import (
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type DotNetCoreCustomInstrumentationCollect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetCoreCustomInstrumentationCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNetCore/CustomInstrumentation/Collect")
}

// Explain - Returns the help text
func (p DotNetCoreCustomInstrumentationCollect) Explain() string {
	return "Collect New Relic .NET Core agent custom instrumentation file(s)" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task.
func (p DotNetCoreCustomInstrumentationCollect) Dependencies() []string {
	return []string{
		"DotNetCore/Agent/Installed",
		"Base/Env/CollectEnvVars",
	}
}

// CustomInstrumentationElement - holds a reference to the custom instrumentation file name and location
type CustomInstrumentationElement struct {
	FileName string
	FilePath string
}

// Execute - The core work within each task
func (p DotNetCoreCustomInstrumentationCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	results := []CustomInstrumentationElement{}
	filesToCopy := []tasks.FileCopyEnvelope{}

	// check if the agent is installed
	checkInstalled := upstream["DotNetCore/Agent/Installed"].Status

	// abort if it isn't installed
	if checkInstalled != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Core Agent not installed, not checking for custom instrumentation files"
		return result
	}

	// get path of agent
	installPath, ok := upstream["DotNetCore/Agent/Installed"].Payload.(string)
	if !ok {
		result.Status = tasks.None
		result.Summary = ".NET Core Agent not installed, not checking for custom instrumentation files"
		return result
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		logger.Debug("Error gathering Environment Variables from upstream. Resolving the New Relic home directory without them.")
		envVars = map[string]string{}
	}

	// the profiler loads instrumentation from <New Relic home>/extensions and <New Relic home>/extensions/netcore
	extensionsPath := filepath.Join(resolveNewRelicHome(runtime.GOOS, installPath, envVars), "extensions")

	logger.Debug("Searching for custom instrumentation xml files in:", extensionsPath)

	// search for xml files, case insensitive
	searchPattern := []string{"^(?i).+[.]xml"}

	// Gather the files
	allInstrumentationFiles := filterToLoadedDirectories(tasks.FindFiles(searchPattern, []string{extensionsPath}), extensionsPath)

	// filter out built in instrumentation files
	customInstrumentationFiles := getCustomInstrumentationFiles(allInstrumentationFiles)

	// check for custom instrumentation files and set up result
	if customInstrumentationFiles != nil {
		result.Status = tasks.Success
		// format the output of the result to return the files found and their content

		for _, custInstFile := range customInstrumentationFiles {
			file := custInstFile
			dir, fileName := filepath.Split(file)

			c := CustomInstrumentationElement{fileName, dir}
			results = append(results, c)
			filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: file, Identifier: p.Identifier().String()})
		}
		// now add the results into a single json string
		logger.Debug("results", results)
		result.Payload = results
		result.Summary = fmt.Sprintf("There were %d file(s) found", len(results))
		result.FilesToCopy = filesToCopy

	} else {
		result.Status = tasks.None
		result.Summary = "Custom Instrumentation files not found."
	}

	return result
}

// resolveNewRelicHome mirrors how the profiler finds its home directory for CoreCLR: the CORECLR_NEW_RELIC_HOME
// environment variable (or its legacy CORECLR_NEWRELIC_HOME name), then on Windows the NewRelicHome the MSI writes
// to the registry, which is the ProgramData directory. Without either, the agent install path is the best guess.
func resolveNewRelicHome(goos string, installPath string, envVars map[string]string) string {
	for _, name := range []string{"CORECLR_NEW_RELIC_HOME", "CORECLR_NEWRELIC_HOME"} {
		if home := envVars[name]; home != "" {
			return home
		}
	}

	if goos == "windows" {
		programData := envVars["ProgramData"]
		if programData == "" {
			programData = `C:\ProgramData`
		}
		return filepath.Join(programData, "New Relic", ".NET Agent")
	}

	return installPath
}

// filterToLoadedDirectories keeps only the files the profiler loads: those directly in the extensions directory or in
// its netcore subdirectory.
func filterToLoadedDirectories(files []string, extensionsPath string) []string {
	// FindFiles resolves symlinks (and Windows short names) in the search path, so the paths it returns must be
	// compared against the resolved path
	if resolved, err := filepath.EvalSymlinks(extensionsPath); err == nil {
		extensionsPath = resolved
	}
	loadedDirs := []string{filepath.Clean(extensionsPath), filepath.Join(extensionsPath, "netcore")}

	var filtered []string
	for _, file := range files {
		dir := filepath.Dir(file)
		for _, loadedDir := range loadedDirs {
			if strings.EqualFold(dir, loadedDir) {
				filtered = append(filtered, file)
				break
			}
		}
	}
	return filtered
}

func getCustomInstrumentationFiles(allFiles []string) []string {
	var customInstrumentationFiles []string

	filterFiles, _ := regexp.Compile("NewRelic[.]Providers[.]Wrapper.+[.]Instrumentation[.]xml")

	// loop through files add them to the payload
	for _, instrumentationFile := range allFiles {
		// filter out built in instrumentation files
		if !filterFiles.MatchString(instrumentationFile) {
			customInstrumentationFiles = append(customInstrumentationFiles, instrumentationFile)
		}
	}

	return customInstrumentationFiles
}
