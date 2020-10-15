package env

import (
	"context"
	"strconv"
	"time"
	"fmt"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// BaseEnvHostInfo - Gets information on a host
type BaseEnvHostInfo struct {
	HostInfoProvider HostInfoProviderFunc
	HostInfoProviderWithContext HostInfoProviderWithContextFunc
}

//On Windows, you can specify a timeout with '-o Base/Env/HostInfo.timeout=N', where N is a number between 1 and 60. The default timeout is 3 seconds.
// set some consts for max, min and default timeout
const TimeoutMax = 60
const TimeoutMin = 1
const TimeoutDefault = 3

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
func (t BaseEnvHostInfo) Execute(options tasks.Options, upstream map[string]tasks.Result) (result tasks.Result) {
	timeoutString := options.Options["timeout"]

	// set the timeout to the default
	timeout := time.Duration(TimeoutDefault) * time.Second

	if timeoutString != "" {
		timeoutInt, err := strconv.Atoi(timeoutString)
		if err == nil {
			if timeoutInt >= TimeoutMin || timeoutInt <= TimeoutMax {
				log.Debug("Base/Env/HostInfo - Using custom timeout of", timeoutInt, "seconds.")
				timeout = time.Duration(timeoutInt) * time.Second
			} else {
				log.Debug("Base/Env/HostInfo - Invalid timeout, valid options are", TimeoutMin, "through", TimeoutMax, ". Using default of", TimeoutDefault, "seconds.")
			}
		} else {
			log.Debug("Base/Env/HostInfo - Error converting timeout from string to int, using default of", TimeoutDefault, "seconds.")
			log.Debug(err.Error())
		}
	}

	result = t.getInfo(timeout)
	return
}


func (t BaseEnvHostInfo) getInfo(timeout time.Duration) (result tasks.Result) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	hostInfo, err := t.HostInfoProviderWithContext(ctx)

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
