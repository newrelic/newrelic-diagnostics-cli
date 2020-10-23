package fixtures

import "github.com/newrelic/newrelic-diagnostics-cli/tasks"

type GoodTaskFileNoDependencies struct {
}

func (p GoodTaskFileNoDependencies) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Good/TaskFile/NoDependencies")
}

func (p GoodTaskFileNoDependencies) Explain() string {
	return "This task doesn't do anything."
}

func (p GoodTaskFileNoDependencies) Dependencies() []string {
	return []string{}
}

func (p GoodTaskFileNoDependencies) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	return tasks.Result{}
}
