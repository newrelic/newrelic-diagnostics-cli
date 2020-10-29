package custominstrumentation

import (
	"runtime"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type DotNetCoreCustomInstrumentationCollect struct { 
	name string
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
	extensionsPath := ""

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

	// set up path to custom instrumentation xml files
	if runtime.GOOS == "windows" {
		extensionsPath = filepath.Join(installPath, "Extensions")
	} else {
		extensionsPath = filepath.Join(installPath, "extensions")
	}
	custInstXMLPath := []string{extensionsPath}

	logger.Debug("Searching for custom instrumentation xml files in:", extensionsPath)

	// search for xml files, case insensitive
	searchPattern := []string{"^(?i).+[.]xml"}

	// Gather the files
	allInstrumentationFiles := tasks.FindFiles(searchPattern, custInstXMLPath)

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
