package agentcontrol

import (
	"net/http"
	"os"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering AgentControl/Agent/*")
	registrationFunc(AgentControlStatusServer{
		httpGetter:   httpHelper.MakeHTTPRequest,
		configReader: os.ReadFile,
	}, true)
	registrationFunc(AgentControlAgentConnect{
		httpGetter:   httpHelper.MakeHTTPRequest,
		configReader: os.ReadFile,
	}, true)
	registrationFunc(AgentControlLogCollect{
		cmdExec: tasks.CmdExecutor,
	}, true)
}
