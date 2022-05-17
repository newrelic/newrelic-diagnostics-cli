package config

import (
	"github.com/newrelic/newrelic-diagnostics-cli/internal/haberdasher"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Config/*")

	registrationFunc(BaseConfigValidate{}, true)
	registrationFunc(BaseConfigCollect{}, true)
	registrationFunc(BaseConfigLogLevel{}, false)
	registrationFunc(BaseConfigProxyDetect{}, true)
	registrationFunc(BaseConfigLicenseKey{}, true)
	registrationFunc(BaseConfigValidateLicenseKey{
		validateAgainstAccount: validateAgainstAccount,
	}, true)
	registrationFunc(BaseConfigAppName{}, true)
	registrationFunc(BaseConfigRegionDetect{}, true)
	registrationFunc(BaseConfigValidateHSM{
		hsmService: haberdasherHSMService,
	}, true)
}

func haberdasherHSMService(licenseKeys []string) ([]haberdasher.HSMresult, *haberdasher.Response, error) {
	return haberdasher.DefaultClient.Tasks.CheckHSM(licenseKeys)
}
