package config

import (
	"fmt"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/internal/haberdasher"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var HSM_CONFIG_NAMES = []string{
	"highSecurity",           //DotNet
	"high_security",          // Ruby, Java, Node, Python
	"newrelic.high_security", // PHP
}

var HSM_ENV_VAR = "NEW_RELIC_HIGH_SECURITY" //Ruby, Node, Python

type HSMLocalValidation struct {
	LicenseKey string
	LocalHSM   map[string]bool
}

type HSMservice func([]string) ([]haberdasher.HSMresult, *haberdasher.Response, error)
type CreateHSMLocalValidation func([]ValidateElement, BaseConfigValidateHSM) map[string]bool
type GetHSMConfiguration func(ValidateElement) bool

// BaseConfigLicenseKey - Struct for task definition
type BaseConfigValidateHSM struct {
	configElements           []ValidateElement
	createHSMLocalValidation CreateHSMLocalValidation
	getHSMConfiguration      GetHSMConfiguration
	envVars                  map[string]string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseConfigValidateHSM) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/ValidateHSM")
}

// Explain - Returns the help text for each individual task
func (t BaseConfigValidateHSM) Explain() string {
	return "Validate High Security Mode agent configuration against account configuration"
}

// Dependencies - Returns the dependencies for each task.
func (t BaseConfigValidateHSM) Dependencies() []string {
	return []string{
		"Base/Config/Validate",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within each task
func (t BaseConfigValidateHSM) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		log.Debug("Could not check env vars for HSM validation")
	} else {
		t.envVars = envVars
	}

	t.configElements, ok = upstream["Base/Config/Validate"].Payload.([]ValidateElement)
	if !ok {
		log.Debug("Could not check configuration files for HSM validation")
	}

	if len(t.configElements) == 0 && len(t.envVars) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No New Relic configuration files or environment variables to check high security mode against. Task did not run.\n",
		}
	}

	hsmValidations := t.createHSMLocalValidation(t.configElements, t)

	localHSMSummary := ""
	localHSMSummaryPattern := "Local High Security Mode setting (%v) for configuration:\n\n\t%s\n\n"

	for k, v := range hsmValidations {
		localHSMSummary += fmt.Sprintf(localHSMSummaryPattern, v, k)
	}
	if localHSMSummary == "" && len(hsmValidations) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No configurations for high security mode found.\n",
		}
	}
	return tasks.Result{
		Status:  tasks.Info,
		Summary: localHSMSummary,
		Payload: hsmValidations,
	}

}

func (t BaseConfigValidateHSM) GetHSMConfigurations(configElements []ValidateElement) map[string]bool {
	configSourcesToHSM := make(map[string]bool)

	hsmEnv, ok := t.envVars[HSM_ENV_VAR]
	if ok {
		hsmEnvBool, parseErr := strconv.ParseBool(hsmEnv)

		if parseErr == nil {
			configSourcesToHSM[HSM_ENV_VAR] = hsmEnvBool

		}

	}

	for _, configElement := range configElements {

		fullPath := configElement.Config.FilePath + configElement.Config.FileName
		isHSM := t.getHSMConfiguration(configElement)

		configSourcesToHSM[fullPath] = isHSM
	}
	return configSourcesToHSM
}
