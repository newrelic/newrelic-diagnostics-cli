package env

import (
	"runtime"

	"github.com/newrelic/NrDiag/helpers/httpHelper"
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Env/*")

	registrationFunc(InfraEnvClockSkew{
		httpGetter:        httpHelper.MakeHTTPRequest,
		checkForClockSkew: checkForClockSkew,
		runtimeOS:         runtime.GOOS,
	}, true)

}
