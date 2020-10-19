
# How To Build A (minimal) Task (and get it running)
## Assumptions

For this document it is assumed that:
 * You have a working `go` environment
 * You can compile and run the `nrdiag` project
 * You are working off a fork of NrDiag
 

**This** doc only goes over the minimal mechanics to get a task up and running.

As you read through, you are welcome to substitute the names/locations with ones that are more meaningful to the task you are writing. However, if you run into issues it is strongly encouraged that you follow the steps more verbatim to make sure something didn't get missed; you can rename things again after finishing out the steps here.

## Steps

### Decide what you want your task to do...

For this HOW-TO we'll create a task that checks on how our coffee at work tastes. Looking at the folder structure already in the project, it seems the best place to put our new task would be in the `java` folder. Since our coffee maker can produce different brews the correct sub-folder might be `jvm`; which obviously stands for "Java Varieties Maker".

The name of the task is often a verb or adjective... think about what your task is *doing* when it runs or what it describes about the system. In this case, our task will be tasting coffee, so we'll just use "taste" as the name of the task. (We could call it "coffee taster" for the name, but since we're already in the Java category, that seems redundant.)

However, someone else might want to have a task to check the flavor of hot chocolate, so for clarity we're going to name the task in code after the folder structure we're storing it: `JavaJVMTaste` seems short and descriptive, that's good!

### Making copies

Make a copy of `./tasks/example/template/minimal_task.go` into `./tasks/java/jvm/taste.go` (note how the task name we choose lines up with the folder structure and filename).

### Search and Replace

We've got a file, but it still has all the original example names

1. Open up your new `taste.go` file.
1. Look at the first line, where the `package` is defined, and replace `template` with `jvm` (the package name will always match the folder we're in).
1. Do a search for `ExampleTemplateMinimalTask` and replace it with `JavaJVMTaste` - there should be 5 of these in code and 1 in a comment. (Be sure to replace the comment as well.)
1. Find the `Identifier()` function and replace `"Example/Template/MinimalTask"` with `"Java/JVM/Taste"`
1. Find the `Explain()` method and replace the string being returned with `"Samples all the coffee varieties to make sure they taste good."`

### Register the new task

Now we have a task that doesn't do much, but we should at least be able to compile and run it. However, NR Diag doesn't know that it's *supposed* to run this task yet. So, we have to "register" the new task (note: this registration step is unique to this project; it's not a normal `go` thing).

Find the "main" package file for the package/folder your task is in. In our case, this is going to be the file `./tasks/java/jvm/jvm.go`

In that file, add another line at the end of the `RegisterWith` method:
`registrationFunc(JavaJVMTaste{}, true)`

Note: if your task is not one that should be run by default, change the `true` to `false` - the rest of these steps should work fine either way.

### Compile and Test

From the main directory of the project, type:

`go build && ./nrdiag -t Java/JVM/Taste`

You should see a result that looks something like:

```
Check Results
-------------------------------------------------


1 results not shown: 1 None

No Issues Found
```

Whew! Done with the "paperwork"... we can finally get started on the **actual** task.

#### I didn't see that!

If you are getting a compile error, it's likely there is a minor typo somewhere in `taste.go` - you'll have to read the compiler complaints and see if you can work out what's wrong.

If it compiled and your result looked more like this:

```
Check Results
-------------------------------------------------

No valid tasks found! (If you used a '*' with the -t option, be sure to quote or escape the string.)

No Issues Found
```

This probably means that either:

1. Your new task isn't registered correctly in the `jvm.go` file
1. The name specified on the command line doesn't match what is returned by the `Identifier()` method

You might get a hint of what went wrong by looking at the help screen that lists all registered tasks.

`./nrdiag -h tasks`

Along with all the other output, you should see:

```
|- Java
|     |- Config
|     |     \- Agent                - Detects if the Java agent is installed.
|     |
|     \- JVM
|           |- SysPropCollect       - Collects the command line options for each JVM with a running NR Java Agent.
|           \- Taste                - Samples all the coffee varieties to make sure they taste good.
```

### Add your logic

Go find the `Execute()` method. There is a lot of stuff passed in, but we're not going to look at that now. The only thing we care about is that we have to return a `Result`.

You should see this:

```
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}
```

Keep that; it just declares a variable that we'll fill with data and return later.  For example, the next line:

```
	return result
```

Let's do something a little more interesting... let's figure out where we're running from. After all, your perspective can influence your taste for coffee. Try adding this code in between the variable declaration and the return statement:

```
	if hostname, _ := os.Hostname(); hostname == "teapot" {
		result.Status = tasks.Failure
		result.Summary = "I'm a teapot."
		result.URL = "https://httpstatuses.com/418"
		result.Payload = 418
	} else {
		result.Status = tasks.Success
		result.Summary = "Everyone likes coffee!"
		result.Payload = []string{
			"Caffeinated",
			"With milk and cream",
			"Decaf",
		}
	}
```

Note: If you are using `gofmt` (or an editor extension that calls it) then you might notice that `"os"` shows up in the `import` statement at the top of the file. Gofmt recognizes that we referred to the `os` package in this new block of code and automatically adds the necessary import.

If you are not running `gofmt`... you should. (No really, we want consistency!) But, in the mean time, you can add the import manually to get it to compile.

### Compile and Test

Run the same commands as above, you should see:

```

Check Results
-------------------------------------------------

Java/JVM/Taste: Success

No Issues Found
```

However, if your hostname happens to be "teapot" you'd see:

```
Check Results
-------------------------------------------------

Java/JVM/Taste: Failure

Issues Found
-------------------------------------------------
Failure
I'm a teapot.
See https://httpstatuses.com/418 for more information.
```

Feel free to change the hostname specified in the `if` statement to see if you can get both results.

#### But wait, we never removed the None status!

True... we never removed the default result even though logically it can never be returned. Generally it's a good idea to initialize your variables as a safety measure (yes, go will do that for you, but in our case it doesn't set a reasonable default). Especially while testing you should leave that in - although change the Summary message!. This will help you catch completeness errors for different conditions.

In fact, let's change that message now from `"I succeeded in doing nothing."` to `"There was no coffee to taste."`. This is a reasonable response if something edge case in our code allows it to complete without setting a new status or summary.

### Testing
 
Don't forget to add tests to your Task. For more details about tests, see [TESTING.md](./TESTING.md)

### Profit!

We're done with this example!

### Other things to try...

There are a lot of additional things you can try. Check out our other examples for some ideas and more sample code.

#### Specify dependencies on other tasks to use their results.
`dependent_task.go` - only looks at the status of the dependency to make a decision

`dependent_payload_task.go` - uses a type assertion to read the payload of the dependency

#### Use the Error result if a check can't be run when we expect it to.
See the type assertion logic in `dependent_payload_task.go`

#### Using a named struct for a more complex payload.
`custom_payload_task.go`

#### Adding custom JSON marshaling
`custom_payload_json_task.go`

#### Creating helper functions in your file.

#### Copying files into the output zip archive
`copy_files_task.go`
