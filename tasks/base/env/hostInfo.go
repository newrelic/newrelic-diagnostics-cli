// +build !windows

package env

import (
	"fmt"
	"github.com/newrelic/NrDiag/tasks"
)



// Gets information on a host
type BaseEnvHostInfo struct {
	HostInfoProvider HostInfoProviderFunc
	HostInfoProviderWithContext HostInfoProviderWithContextFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseEnvHostInfo) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Env/HostInfo")
}

// Explain - Returns the help text for each individual task
func (t BaseEnvHostInfo) Explain() string {
	return "Collect host system info"
}

// Dependencies - Returns the dependencies for ech task.
func (t BaseEnvHostInfo) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BaseEnvHostInfo) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	
	hostInfo, err := t.HostInfoProvider()

	if err != nil {
		return tasks.Result {
			Status: tasks.Warning,
			Summary: fmt.Sprintf("Error collecting complete host information:\n%s", err.Error()),
			Payload: hostInfo,
		}
	}
	
	return tasks.Result {
		Status: tasks.Info,
		Summary: "Collected host information",
		Payload: hostInfo,
	}

}



