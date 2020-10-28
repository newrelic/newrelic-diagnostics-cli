package registration

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestRegisterSingleTask(t *testing.T) {
	//baseConfig.LogLevel = baseConfig.Verbose

	// re-init the results struct
	Work.Results = make(map[string]TaskResult)
	// make a large channel so we don't block
	Work.WorkQueue = make(chan tasks.Task, 100)
	// re-init the queue
	queuedTasks = make(map[tasks.Identifier]bool)

	AddIdentifierToQueue(tasks.IdentifierFromString("Base/Env/CollectEnvVars"))
	CompleteTaskRegistration()

	//	dumpTasks(Work.Tasks)
	//	dumpQueue(Work.WorkQueue)
	if len(Work.WorkQueue) != 1 {
		t.Error("WorkQueue expected to have 1 items after adding Base/Env/CollectEnvVars; has:", len(Work.WorkQueue))
	}
}

func TestRegisterDependentTasks(t *testing.T) {
	// config.LogLevel = config.Verbose

	// re-init the results struct
	Work.Results = make(map[string]TaskResult)
	// make a large channel so we don't block
	Work.WorkQueue = make(chan tasks.Task, 100)
	// re-init the queue
	queuedTasks = make(map[tasks.Identifier]bool)

	AddIdentifierToQueue(tasks.IdentifierFromString("Base/Config/Validate"))
	CompleteTaskRegistration()

	// dumpTasks(Work.Tasks)
	// dumpQueue(Work.WorkQueue)

	if len(Work.WorkQueue) != 3 {
		t.Error("WorkQueue expected to have 3 items after adding Base/Config/Validate; has:", len(Work.WorkQueue))
	}
}

func TestRegisterAllTasks(t *testing.T) {
	//baseConfig.LogLevel = baseConfig.Verbose

	// re-init the results struct
	Work.Results = make(map[string]TaskResult)
	// make a large channel so we don't block
	Work.WorkQueue = make(chan tasks.Task, 200)

	queuedTasks = make(map[tasks.Identifier]bool)

	AddAllToQueue()
	CompleteTaskRegistration()

	// dumpTasks(Work.Tasks)
	// dumpQueue(Work.WorkQueue)

	runnableTasks := 0
	for _, regTask := range registeredTasks {
		if regTask.runByDefault {
			runnableTasks++
		}
	}
	if len(Work.WorkQueue) != runnableTasks {
		t.Error("WorkQueue expected to have same number of items as Tasks: ", len(Work.WorkQueue), " vs. ", runnableTasks, "/", len(registeredTasks))
	}
}

func TestTasksHaveValidExplain(t *testing.T) {
	for _, regTask := range registeredTasks {
		explain := regTask.Task.Explain()
		if (explain == "Explaintory help text displayed for this task" || explain == "This task doesn't go anything") && regTask.Task.Identifier().Category != "Example" {
			t.Error(regTask.Task.Identifier().String(), " still has template Explain() message. Please update with something specific to this task.")
		}
	}
}

// func countPatternInFile(pattern string, filename string) int {
// 	lines, err := ioutil.ReadFile(filename)
// 	if err != nil {
// 		panic(err)
// 	}
// 	r, _ := regexp.Compile(pattern)
// 	matches := r.FindAllStringSubmatch(string(lines), -1)

// 	return len(matches)
// }

// func dumpTasks(ts map[string]registeredTask) {
// 	fmt.Println("### map has ", len(ts), " items")
// 	for _, t := range ts {
// 		if t.runByDefault {
// 			fmt.Println("dumping: ", t.Task.Identifier())
// 		}
// 	}
// }

// func dumpQueue(q chan tasks.Task) {
// 	fmt.Println("### queue has ", len(q), " items")
// 	for t := range q {
// 		fmt.Println("from queue: ", t.Identifier())
// 	}
// }
