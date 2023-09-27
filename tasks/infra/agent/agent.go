package agent

import (
	"net/http"
	"runtime"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Agent/*")

	registrationFunc(InfraAgentVersion{
		runtimeOS:   runtime.GOOS,
		now:         time.Now,
		cmdExecutor: tasks.CmdExecutor,
		httpGetter:  httpHelper.MakeHTTPRequest,
	}, true)
	registrationFunc(InfraAgentDebug{
		blockWithProgressbar: blockWithProgressbar,
		cmdExecutor:          tasks.CmdExecutor,
	}, false)
	registrationFunc(InfraAgentConnect{
		httpGetter: httpHelper.MakeHTTPRequest,
	}, true)

}
