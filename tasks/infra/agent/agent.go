package agent

import (
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Agent/*")

	registrationFunc(InfraAgentVersion{
		runtimeOS:   runtime.GOOS,
		cmdExecutor: tasks.CmdExecutor,
	}, true)
	registrationFunc(InfraAgentDebug{
		blockWithProgressbar: blockWithProgressbar,
		cmdExecutor:          tasks.CmdExecutor,
		runtimeOS:            runtime.GOOS,
	}, false)
	registrationFunc(InfraAgentConnect{
		httpGetter: httpHelper.MakeHTTPRequest,
	}, true)

}
