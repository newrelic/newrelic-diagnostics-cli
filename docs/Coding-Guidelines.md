# Guidelines and practices when contributing

## Coding Conventions

### General principles

* Code readability and maintainability is top priority.
  * Don’t be clever.
  * Don’t sacrifice readability to DRY up code.
  * Consistent naming is important, if you want to include the name of this tool in a tasks output use const `tasks.ThisProgramFullName`
* Your code will be running on the customer’s machine.
  * Be sensitive to your code’s memory footprint (e.g. don’t load large files into memory).
  * Don’t perform blocking operations for long periods of time.
  * Don’t run shell commands unless necessary.
  * Don’t make network calls unless necessary.
* Respect end users data! Gather only data required for check.
  * If you’re curious about whether you should gather potentially sensitive data, work with a maintainer of the code
  * Tests do not need to be DRY

### Coding style

Execute should return `tasks.Result{}` instances immediately (do not mutate them before return)

```go
//Good
if configFileFound {
    return tasks.Result{
        Status: tasks.Success,
        Summary: "Config file found!"
    }
}

//Bad
var result tasks.Result

if configFileFound {
    result.Status = tasks.Success
    result.Summary = "Config file found!"
    return result
}
```

Always opt for early return rather than nesting code in positive case if-blocks

```go
//Good
if upstream["Infra/Config/IntegrationsValidate"].Status != tasks.Success {
    return tasks.Result{
        Status:  tasks.None,
        Summary: "Task did not meet requirements necessary to run: no validated integrations",
    }
}
//Happy path logic ...


//Bad
if upstream["Infra/Config/IntegrationsValidate"].Status == tasks.Success {
    //Happy path logic ...
} else {
    return tasks.Result{
        Status:  tasks.None,
        Summary: "Task did not meet requirements necessary to run: no validated integrations",
    }
}
```

Early in a tasks `Execute()` method, tasks which rely on upstream dependencies should perform a type assertion the payloads of dependencies before proceeding. 

```go
// example type assertion
logs, ok := upstream["Base/Log/Collect"].Payload.([]log.LogElement)

if !ok {
	return tasks.Result{
		Status:  tasks.None,
		Summary: "Task did not meet requirements necessary to run: type assertion failure",
	}
}
```

* If possible, only Execute() should return `tasks.Result` type
* Most logic should be outside of Execute-- execute is assembly line. Not an absolute rule: use common sense.
* Use [dependency injection](./Dependecy-injection.md) for external APIs (including running commands on host) when reasonable
* Don’t use implicit returns (increases cognitive load)
* Don’t use pointer receiver unless necessary
* When dealing with values that are subject to change (e.g. support agent versions), isolate these variables outside of code logic (e.g. as a dependency when the task is instantiated).
* Single letter variables are fine when close to their usage (e.g. for _, k := range keys { log(k) }), otherwise use variables descriptive of their contents (not their type)

## End user facing verbiage

Each task has an `Explain()` method which tells the user its purpose. Try to make these as succinct as possible and in the imperative case:

* **Good:** "Determine New Relic Java agent version"
* **Bad:** "This task determines the New Relic Java agent version"

Scope to task outputs, not implementation details

* **Good:**  "Determine New Relic Java agent version"
* **Bad:** "Determine New Relic Java agent version from configuration files"

### Task explain verbs

* **"Detect"** - truthy/falsy situations (presence of something)
  * "Detect if running in ______ environment"
  * "Detect New Relic _____ agent"
* **"Determine"** - single detail (version of something)
  * "Determine New Relic _____ agent version"
  * "Determine ______ version"
  * "Determine New Relic account id and application id"
* **"Collect"** - multiple details (app dependencies)
  * "Collect ______ application dependencies"
* **"Identify"** - active state stuff (running processes)
  * “Identify running ________ processes”
* **"Check"** - single assertion of something (network connection to endpoint)
  * "Check _______ compatibility with New Relic ______ agent"
* **"Validate"** - many assertions (validate values in configuration file)
  * "Validate _______ configuration files"

## Testing

We recommend test-driven development when working on tasks. See our [doc on unit testing](./Unit-testing.md) for more information.

You should add additional functions to the task to do any heavy lifting required. The `Execute()` function should call your additional functions and be kept clean and lightweight on any task logic. In general writing functions that get called to do the heavy lifting both makes it easier to write tests against your business logic and keeps the actual execution path of the task easier to read and understand. This also makes it easier to write unit tests against each one of the functions used within your task

When building a task for Windows, be sure to put `_windows` on all the files. This tells Go only to include that task when building for Windows.

`Result.URL` must always be set when a task fails. It should point to public documentation to help resolve the failure. If there is no documentation that is relevant, attempt to draft some or file a GitHub Issue.
