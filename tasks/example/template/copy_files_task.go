package template

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// ExampleTemplateCopyFilesTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateCopyFilesTask struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateCopyFilesTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/CopyFilesTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateCopyFilesTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateCopyFilesTask) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p ExampleTemplateCopyFilesTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	//look for any go testing files
	patterns := []string{".*_test\\.go"}
	paths := []string{"."}

	//call our helpers function for finding files, but you can do this other ways
	files := tasks.FindFiles(patterns, paths)

	if len(files) == 0 {
		result.Status = tasks.Warning
		result.Summary = "Didn't find any Go files."
	} else {
		result.Status = tasks.Success
		result.Summary = "Found at least one Go file."
		//make a note of the files that need to be copied
		result.FilesToCopy = tasks.StringsToFileCopyEnvelopes(files)
	}

	//if the FilesToCopy was filled, simply returning this result will kick off the file copy process
	return result
}
