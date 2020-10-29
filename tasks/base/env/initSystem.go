package env

import (
	"fmt"
	"regexp"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseEnvInitSystem - This struct defined the sample plugin which can be used as a starting point
type BaseEnvInitSystem struct {
	runtimeOs   string
	evalSymlink func(string) (string, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseEnvInitSystem) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/InitSystem")
}

// Explain - Returns the help text for each individual task
func (p BaseEnvInitSystem) Explain() string {
	return "Determine Linux init system"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseEnvInitSystem) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p BaseEnvInitSystem) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	if p.runtimeOs == "windows" {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task does not apply to Windows",
		}
	}
	if p.runtimeOs == "darwin" {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task does not apply to Mac OS",
		}
	}

	initPath, err := p.evalSymlink("/sbin/init")
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to read symbolic link for /sbin/init: %s", err.Error()),
		}
	}

	initSystem := parseInitSystem(initPath)
	if initSystem == "" {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to parse init system from: %s", initPath),
		}
	}

	return tasks.Result{
		Status:  tasks.Info,
		Summary: fmt.Sprintf("%s detected", initSystem),
		Payload: initSystem,
	}
}

func parseInitSystem(initPath string) string {
	//https://linuxconfig.org/detecting-which-system-manager-is-running-on-linux-system
	regexSysD := regexp.MustCompile(`.*(\/systemd$)`)
	regexUpstart := regexp.MustCompile(`.*(upstart$)`)

	if initPath == "/sbin/init" {
		return "SysV"
	}

	if regexSysD.Match([]byte(initPath)) {
		return "Systemd"
	}

	if regexUpstart.Match([]byte(initPath)) {
		return "Upstart"
	}

	return ""

}