package config

import (
	"strconv"

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
		createHSMLocalValidation: createHSMLocalValidation,
		getHSMConfiguration:      getHSMConfiguration,
	}, true)
}

func createHSMLocalValidation(configElements []ValidateElement, t BaseConfigValidateHSM) map[string]bool {
	return t.GetHSMConfigurations(configElements)
}

func getHSMConfiguration(configElement ValidateElement) bool {
	for _, hsmName := range HSM_CONFIG_NAMES {
		foundKeys := configElement.ParsedResult.FindKey(hsmName)

		if len(foundKeys) == 0 {
			continue
		}

		hsmStatus, ok := strconv.ParseBool(foundKeys[0].Value())

		//If found field is set to something other than a bool value - implicit false
		if ok != nil {
			return false
		}

		return hsmStatus
	}

	return false //If no HSM field found then implicit false
}
