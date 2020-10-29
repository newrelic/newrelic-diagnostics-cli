package custominstrumentation

import (
	"fmt"
	"path/filepath"
	"regexp"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotNetCustomInstrumentationCollect - This struct defined the sample plugin which can be used as a starting point
type DotNetCustomInstrumentationCollect struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetCustomInstrumentationCollect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/CustomInstrumentation/Collect") // This should be updated to match the struct name
}

// Explain - Returns the help text for each individual task
func (p DotNetCustomInstrumentationCollect) Explain() string {
	return "Collect New Relic .NET agent custom instrumentation file(s)" //This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p DotNetCustomInstrumentationCollect) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed", //This identifies this task as dependent on "DotNet/Agent/Installed" and so the results from that task will be passed to this task. See the execute method to see how to interact with the results.
		"Base/Env/CollectEnvVars",
	}
}

// CustomInstrumentationElement - holds a reference to the custom instrumentation file name and location
type CustomInstrumentationElement struct {
	FileName string
	FilePath string
}

// Execute - The core work within each task
func (p DotNetCustomInstrumentationCollect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	var result tasks.Result //This is what we will use to pass the output from this task back to the core and report to the UI

	results := []CustomInstrumentationElement{}
	filesToCopy := []tasks.FileCopyEnvelope{}
	sysProgramData := ""

	// check if the agent is installed
	checkInstalled := upstream["DotNet/Agent/Installed"].Status

	// abort if it isn't installed
	if checkInstalled != tasks.Success {
		result.Status = tasks.None
		result.Summary = ".NET Agent not installed, not checking for custom instrumentation files"
		return result
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the map[string]string I know it should return
	if !ok {
		log.Debug(`Error gathering Environment Variables from upstream. Will default to C:\ProgramData`)
		sysProgramData = `C:\ProgramData`
	} else {
		log.Debug("Successfully gathered Environment Variables from upstream.")
		sysProgramData = envVars["ProgramData"]
	}

	// set up path to custom instrumentation xml files
	extensionsPath := filepath.Join(sysProgramData, "New Relic", ".NET Agent", "Extensions")
	custInstXmlPath := []string{extensionsPath}

	log.Debug("Searching for custom instrumentation xml files in:", extensionsPath)

	// search for xml files, case insensitive
	searchPattern := []string{"^(?i).+[.]xml"}

	// Gather the files
	allInstrumentationFiles := tasks.FindFiles(searchPattern, custInstXmlPath)

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
		log.Debug("results", results)
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
