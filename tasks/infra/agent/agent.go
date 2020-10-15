package agent

import (
	"runtime"

	"github.com/newrelic/NrDiag/helpers/httpHelper"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
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
