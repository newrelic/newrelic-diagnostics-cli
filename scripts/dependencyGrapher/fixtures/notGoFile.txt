package fixtures

import "github.com/newrelic/NrDiag/tasks"

type GoodTaskFileTwoDependencies struct {
}

func (p GoodTaskFileTwoDependencies) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Good/TaskFile/TwoDependencies")
}

func (p GoodTaskFileTwoDependencies) Explain() string {
	return "This task doesn't do anything."
}

func (p GoodTaskFileTwoDependencies) Dependencies() []string {
	return []string{
		"I/Am/Dependency1",
		"I/Am/Dependency2",
	}
}

func (p GoodTaskFileTwoDependencies) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	return tasks.Result{}
}
