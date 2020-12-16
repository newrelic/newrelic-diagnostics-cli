package config

import (
	"os"
	"runtime"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Infra/Config/*")

	registrationFunc(InfraConfigDataDirectoryCollect{dataDirectoryGetter: getDataDir, dataDirectoryPathGetter: getDataDirPath}, true)
	registrationFunc(InfraConfigAgent{validationChecker: checkValidation, configChecker: checkConfig, binaryChecker: checkForBinary}, true)
	registrationFunc(InfraConfigIntegrationsCollect{fileFinder: tasks.FindFiles}, true)
	registrationFunc(InfraConfigIntegrationsValidate{fileReader: os.Open}, true)
	registrationFunc(InfraConfigIntegrationsMatch{
		runtimeOS: runtime.GOOS,
	}, true)
	registrationFunc(InfraConfigIntegrationsValidateJson{}, true)
	registrationFunc(InfraConfigValidateJMX{
		mCmdExecutor: tasks.MultiCmdExecutor,
	}, true)
}
