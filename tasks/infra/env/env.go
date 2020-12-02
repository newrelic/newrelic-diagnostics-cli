package env

import (
	"runtime"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Env/*")

	registrationFunc(InfraEnvClockSkew{
		httpGetter:        httpHelper.MakeHTTPRequest,
		checkForClockSkew: checkForClockSkew,
		runtimeOS:         runtime.GOOS,
	}, true)
	registrationFunc(InfraEnvNrjmxMbeans{
		getMBeanQueriesFromJMVMetricsYml: getMBeanQueriesFromJMVMetricsYml,
		executeNrjmxCmdToFindBeans:       executeNrjmxCmdToFindBeans,
	}, true)
	registrationFunc(InfraEnvKafkaBrokers{}, true)
}
