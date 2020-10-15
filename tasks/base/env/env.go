package env

import (
	"path/filepath"
	"runtime"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Env/*")

	registrationFunc(BaseEnvHostInfo{
		HostInfoProvider:            NewHostInfo,
		HostInfoProviderWithContext: NewHostInfoWithContext,
	}, true)
	registrationFunc(BaseEnvCollectEnvVars{}, true)
	registrationFunc(BaseEnvDetectAWS{
		httpGetter: tasks.HTTPRequester,
	}, true)
	registrationFunc(BaseEnvInitSystem{
		runtimeOs:   runtime.GOOS,
		evalSymlink: filepath.EvalSymlinks,
	}, true)
	registrationFunc(BaseEnvDetectAzure{}, true)
}
