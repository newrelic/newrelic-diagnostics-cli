# Anatomy of a Task

## Naming

The name is a critical part of the nrdiag task and will drive both the file name itself, it's identity and the identifier.

Example of a task name: Node/Env/Version

All of our tasks live in the tasks directory. Individual tasks are broken out by Category and Subcategory where the Category is the broad grouping of similar tasks and the Subcategory is the functional area of the category.

Example category: Node

Example subcategory: Env

Categories are designed as top level groupings to keep all items related to the New Relic product our tasks are meant to troubleshoot for. You should ensure your task doesn't fit into one of the existing categories before proposing a new one. 

If the category or subcategory you want to use does not yet exist, it is possible to add it but the reasoning for making a new one should be included in any PRs. A new task should always come with a unit test file. You'll notice that the task `Node/Env/Version` itself is written in the `version.go` file inside the directory `tasks/node/env`, and that it comes with a `version_test.go file` for its unit tests.

After you are done writing the task, you will also need to add it to `registerTasks.go` (or `registerTasks_windows.go`) and add the relevant package registration file (Copy [template.go](../tasks/example/template/template.go) in the example task and rename it to match your category, subcategory, and package name).

####  Notes on Sub-Categories

Within each category, we typically have the following subcategories:

 * Config - This deals with config file tasks
 * Log - This deals with log errors and other tasks related to log files
 * Agent - This deals with running agent configuration tasks

 Other example sub-categories that aren't as general as the above ones:

 * Collector - This deals with connections to the New Relic collector
 * Profiler - Specific to dotnet profiler
 * W3wp - Specific to dotnet IIS server tasks 
 * JVM - Specific to Java JVM tasks
 * Daemon - Specific to PHP Daemon tasks
 * Minion - Specific to Synthetics private minion tasks

## Locations

The name determines the directory structure:

```
tasks/<category>/<subcategory>/<task>
```

If you were making a task that checked if the Morty agent is installed, the path would look something like...

```
tasks/morty/agent/installed.go
```

... and the package name would be `agent`.


The registration file to add your task to this package would be:

```
tasks/morty/agent/agent.go
```

This is where you register the task(s) for the Morty agent, so that they can be used and run by the Diagnostics CLI core. Just add a...

```
registrationFunc(MortyAgentTaskname{}, true)
```

... line to the `RegisterWith` function, one for each task, and you're good to go!

*Note:* the `true` includes the task by default when running, change this to `false` for something you don't want run by default.

## Main parts of a task
There are 4 required functions for a task.

* `Identifier()`: This returns the internal representation of the task. All you need to do here is update the string to match your task.
* `Explain()`: This returns the help text shown on the help screen. It should describe what the Task does.
* `Dependencies()`: This is a list of the dependencies for the task. Even if there are no dependencies, this should still be present.
* `Execute()`: This is where the Result variable is created and populated.

There is 1 main variable for a task. 
* `Result`: This is where you store the results of the task. This is a struct that looks like this:

```
type Result struct {
	Status      Status      // something like tasks.Success or tasks.Failure
	Summary     string      // verbiage that explains what the task found
	URL         string      // a URL pointing to documention about the findings of the task; "required" on Warning or Failure, desireable on any status, needs to help explain the findings
	FilesToCopy []string    // List of files identified by the task to be included in zip file
	Payload     interface{} // task defined list of returned data. This is what is used by downstream tasks so data format agreements are between tasks
}
```

There are also 2 optional variables:

* `options`: This is where the custom override comes in. It's accessed via `options.Options["overridehere"]`

* `upstream`: This is where you can access the data from the task's dependencies. It's accessed via `upstream.Results` or `upstream.Status` 


## Code Guidelines and best practices

See [Coding-Guidelines.md](./Coding-Guidelines.md)
