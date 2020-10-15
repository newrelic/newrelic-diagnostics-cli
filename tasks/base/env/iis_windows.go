package env

import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"golang.org/x/sys/windows/registry"
)

var iisRegLoc = `Software\Microsoft\InetStp`
var iisVersionKey = `VersionString`

// ExampleTemplateMinimalTask - This struct defined the sample plugin which can be used as a starting point
type BaseEnvIisCheck struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseEnvIisCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/IisCheck")
}

// Explain - Returns the help text for each individual task
func (t BaseEnvIisCheck) Explain() string {
	return "Determine version of IIS"
}

// Dependencies - Returns the dependencies for ech task.
func (t BaseEnvIisCheck) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BaseEnvIisCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	return checkIIS()
}

func checkIIS() (result tasks.Result) {

	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, iisRegLoc, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug("Error opening IIS Reg Key (InetStp). Error = ", err.Error())
		result.Status = tasks.None
		result.Summary = "Could not open Environment Reg Key " + err.Error()
		return result

	}

	defer regKey.Close()

	regValue, _, regErr := regKey.GetStringValue(iisVersionKey)

	if regErr != nil {
		log.Debug("Error opening IIS version Reg Sub Key. Error = ", regErr.Error())
		result.Status = tasks.Error
		return result
	}

	result.Status = tasks.Info
	result.Summary = regValue
	return result

}
