package profiler

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// How the system env vars look as of December 2019
// Used in (DotNet/Profiler/WasRegKey) to check registry keys at SYSTEM\CurrentControlSet\Services\WAS
// Used in (DotNet/Profiler/W3svcRegKey) to check registry keys at SYSTEM\CurrentControlSet\Services\W3SVC\
var (
	expectedRegKeyExists = []string{
	"NEWRELIC_INSTALL_PATH",
	}
	
	expectedRegKeyWithVals = map[string]string{
	"COR_ENABLE_PROFILING" : "1",
	"COR_PROFILER" : "{71DA0A04-7777-4EC6-9643-7D28B46A8A41}",
	}
)

// RegisterWith - will register any plugins in this package 
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Profiler/*")

	registrationFunc(DotNetProfilerInstrumentationPossible{}, true)
	registrationFunc(DotNetProfilerW3svcRegKey{}, true)
	registrationFunc(DotNetProfilerWasRegKey{}, true)
	registrationFunc(DotNetProfilerEnvVarKey{}, true)
}
