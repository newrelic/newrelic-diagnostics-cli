package env

import (
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type NodeEnvNpmPackage struct {
	Getwd      func() (string, error)
	fileFinder func([]string, []string) []string
}

type PackageJsonElement struct {
	FileName string
	FilePath string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p NodeEnvNpmPackage) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Node/Env/NpmPackage") // This should match the struct name
}

// Explain - Returns the help text for each individual task
func (p NodeEnvNpmPackage) Explain() string {
	return "Collect package.json and package-lock.json if they exist" // This is the customer visible help text that describes what this particular task does
}

// Dependencies - Returns the dependencies for each task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p NodeEnvNpmPackage) Dependencies() []string {
	return []string{
		"Node/Config/Agent",
		"Node/Env/NpmVersion",
	}
}

// Execute - The core work within each task
func (p NodeEnvNpmPackage) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if upstream["Node/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Node agent not detected. This task did not run",
		}
	}

	if upstream["Node/Env/NpmVersion"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "NPM is not installed. This task did not run",
		}
	}

	// Search for package.json and package-lock.json files in working directory and append to filesToCopy

	foundFiles, err := p.findPackageFiles()
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to search for package.json and package-lock.json files. Error:\n%s", err.Error()),
		}
	}

	if len(foundFiles) == 0 {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: "The package.json and package-lock.json files were not found where NR Diagnostics was executed. Please ensure the NR Diagnostics executable is within your application's directory alongside your package.json file",
		}
	}

	// format the output of the result to return the files found and their content
	pje := []PackageJsonElement{}
	filesToCopy := []tasks.FileCopyEnvelope{}

	for _, file := range foundFiles {
		dir, fileName := filepath.Split(file)
		log.Debug("found file", fileName)
		pje = append(pje, PackageJsonElement{fileName, dir})
		filesToCopy = append(filesToCopy, tasks.FileCopyEnvelope{Path: file})
	}

	resultSummary := "We have succesfully retrieved the following file(s):"

	for _, file := range pje {
		resultSummary += "\n" + file.FileName
	}

	return tasks.Result{
		Status:      tasks.Success,
		Summary:     resultSummary,
		Payload:     pje,
		FilesToCopy: filesToCopy,
	}
}

//Recursively find package*.json files, skipping those in node_modules
func (p NodeEnvNpmPackage) findPackageFiles() ([]string, error) {
	var packageFiles []string

	workingDir, err := p.Getwd()
	if err != nil {
		return packageFiles, err
	}

	foundFiles := p.fileFinder([]string{"package.json", "package-lock.json"}, []string{workingDir})

	for _, filePath := range foundFiles {
		if strings.Contains(filePath, "node_modules") {
			continue
		}
		packageFiles = append(packageFiles, filePath)
	}

	return packageFiles, nil
}
