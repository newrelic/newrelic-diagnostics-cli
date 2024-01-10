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

type HSMvalidation struct {
	LicenseKey string
	AccountHSM bool
	LocalHSM   map[string]bool
}
type HSMLocalValidation struct {
	LicenseKey string
	LocalHSM   map[string]bool
}

func (h HSMvalidation) Validate() (bool, []string) {
	misMatchedSources := []string{}
	isValid := true

	for source, sourceHSM := range h.LocalHSM {
		if sourceHSM != h.AccountHSM {
			misMatchedSources = append(misMatchedSources, source)
			isValid = false
		}
	}

	return isValid, misMatchedSources
}

type HSMservice func([]string) ([]haberdasher.HSMresult, *haberdasher.Response, error)

// BaseConfigLicenseKey - Struct for task definition
type BaseConfigValidateHSM struct {
	configElements []ValidateElement
	hsmService     HSMservice
	envVars        map[string]string
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
		"Base/Config/ValidateLicenseKey",
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
	if !ok || (len(t.configElements) == 0) {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No New Relic configuration files to check high security mode against. Task did not run.\n",
		}
	}

	hsmValidations := t.createHSMLocalValidation()

	localHSMSummary := ""
	localHSMSummaryPattern := "Local High Security Mode setting (%v) for configuration filepath:\n\n%s\n\n"

	for k, v := range hsmValidations {
		localHSMSummary += fmt.Sprintf(localHSMSummaryPattern, v, k)
	}
	return tasks.Result{
		Status:  tasks.Info,
		Summary: localHSMSummary,
		Payload: hsmValidations,
	}

}

func (t BaseConfigValidateHSM) createHSMLocalValidation() map[string]bool {
	elementsToPass := t.configElements
	return t.getHSMConfigurations(elementsToPass)
}

func (t BaseConfigValidateHSM) getHSMConfigurations(configElements []ValidateElement) map[string]bool {

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

func (t BaseConfigValidateHSM) getHSMConfiguration(configElement ValidateElement) bool {
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
