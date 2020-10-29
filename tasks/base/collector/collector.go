package collector

import (
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Collector/*")

	registrationFunc(BaseCollectorConnectUS{
		httpGetter: httpHelper.MakeHTTPRequest,
	}, true)
	registrationFunc(BaseCollectorConnectEU{
		httpGetter: httpHelper.MakeHTTPRequest,
	}, true)
}
