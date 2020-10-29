package registration

import (
	"encoding/json"
	"regexp"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type registeredTask struct {
	Task         tasks.Task
	runByDefault bool
}

// TaskResult is a holding tank for a task and it's result after execution
type TaskResult struct {
	Task        tasks.Task
	Result      tasks.Result
	WasOverride bool
}

//MarshalJSON - custom JSON marshaling for this task, we'll strip out the passphrase to keep it only in memory, not on disk
func (tr TaskResult) MarshalJSON() ([]byte, error) {
	//note: this technique can be used to return anything you want, including modified values or nothing at all.
	//anything that gets returned here ends up in the output json file
	return json.Marshal(&struct {
		Identifier tasks.Identifier
		Override   bool
		Result     tasks.Result
	}{
		Identifier: tr.Task.Identifier(),
		Override:   tr.WasOverride,
		Result:     tr.Result,
	})
}

// Workload - the work that has to be done for tasks
type Workload struct {
	WorkQueue      chan tasks.Task
	Results        map[string]TaskResult
	ResultsChannel chan TaskResult
	FilesChannel   chan TaskResult
}

// Work is out "main" object that stores what we need to do
var Work = Workload{
	Results:        make(map[string]TaskResult),
	ResultsChannel: make(chan TaskResult, 2),
	FilesChannel:   make(chan TaskResult, 2),
}

var registeredTasks = make(map[string]registeredTask)
var queuedTasks = make(map[tasks.Identifier]bool)

// Register - allows registration of tasks, probably only used as a callback
// Passing false as the second option prevents the task from running by default.
func Register(t tasks.Task, runByDefault bool) {
	log.Debug("  - " + t.Identifier().String())
	registeredTasks[strings.ToLower(t.Identifier().String())] = registeredTask{Task: t, runByDefault: runByDefault}
}

//TasksForIdentifierString - this returns the registered task(s) for a given identifier, it can have wildcards
func TasksForIdentifierString(ident string) []tasks.Task {
	var tasks []tasks.Task

	if strings.Contains(ident, "*") {
		converter, _ := regexp.Compile("\\*")
		matchString := converter.ReplaceAllString(strings.ToLower(ident), ".*")
		matcher, err := regexp.Compile(matchString)
		if err != nil {
			log.Info("Failed to compile identifier regex from '", matchString, "'")
		}
		for id, regTask := range registeredTasks {
			if regTask.runByDefault && matcher.MatchString(id) {
				tasks = append(tasks, regTask.Task)
			}
		}
	} else {
		if regTask, ok := registeredTasks[strings.ToLower(ident)]; ok {
			tasks = append(tasks, regTask.Task)
		}
	}

	return tasks
}

// AddAllToQueue - adds in all tasks that have been registered
func AddAllToQueue() {
	log.Debugf("Adding %d tasks to queue\n", len(registeredTasks))
	for _, regTask := range registeredTasks {
		if regTask.runByDefault {
			AddTaskToQueue(regTask.Task)
		}
	}
	log.Debugf("Added %d tasks to queue\n", len(Work.WorkQueue))
}

//AddTasksByIdentifiers - will use an identifier string to add tasks, can have wildcards
func AddTasksByIdentifier(ident string) {
	log.Debugf("asked to load %s by string\n", ident)
	tasks := TasksForIdentifierString(ident)
	if len(tasks) == 0 {
		log.Info("No valid tasks found! (If you used a '*' with the -t option, be sure to quote or escape the string.)")
	} else {
		for _, task := range tasks {
			AddTaskToQueue(task)
		}
	}
}

//AddTasksByIdentifiers - takes slice of tasks identifier strings and adds all matching tasks to work queue
func AddTasksByIdentifiers(idents []string) {
	for _, ident := range idents {
		AddTasksByIdentifier(ident)
	}
}

// AddIdentifierToQueue - adds a single Task (identified by name) to the work queue
func AddIdentifierToQueue(ident tasks.Identifier) {
	log.Debugf("asked to load %s by Identifier\n", ident.String())
	regTask := registeredTasks[strings.ToLower(ident.String())]
	if regTask.Task == nil {
		log.Debug(" * Could not find task!")
	} else {
		AddTaskToQueue(regTask.Task)
	}
}

// AddTaskToQueue - adds in a new task and resolves it's dependencies, could be prone to dependency loops
func AddTaskToQueue(p tasks.Task) {
	//QueuedTasks := make(map[tasks.Identifier]string)
	//dent := p.Identifier().String()
	// add all the dependencies for this
	for _, depIdent := range p.Dependencies() {
		log.Debugf("\tfound dependency %s\n", depIdent)
		AddTasksByIdentifier(depIdent)
	}

	// somewhere in here may be a good place to detect dependency loops...
	// if this task has dependencies *and* it's already in the  task list we are in an invalid state
	// since it would be impossible to resolve the dependencies beore it runs

	// if we have already created a key for the results then we aren't in the queue yet
	log.Debug("Checking queue for ", p.Identifier(), ": ", queuedTasks[p.Identifier()])
	if _, ok := queuedTasks[p.Identifier()]; !ok {
		log.Debugf("Couldn't find %s in queue set\n", p.Identifier())
		queuedTasks[p.Identifier()] = true
		log.Debug("adding to queue")
		Work.WorkQueue <- p
	} else {
		log.Debug("already had in queue: ", p.Identifier(), " or ", p.Identifier())
	}
	log.Debug("done with add task to queue")
}

// CompleteTaskRegistration - does some clean up after the setup process
func CompleteTaskRegistration() {
	log.Debug("Closing task registration.")
	close(Work.WorkQueue)
}
