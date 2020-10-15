package config

import (
	"fmt"
	"errors"
	"github.com/newrelic/NrDiag/tasks"
	log "github.com/newrelic/NrDiag/logger"
	"strings"
	"strconv"
	"github.com/newrelic/NrDiag/internal/haberdasher"
)

var HSM_CONFIG_NAMES = []string{
	"highSecurity", //DotNet
	"high_security", // Ruby, Java, Node, Python
	"newrelic.high_security", // PHP
}

var HSM_ENV_VAR = "NEW_RELIC_HIGH_SECURITY" //Ruby, Node, Python

type HSMvalidation struct {
	LicenseKey string
	AccountHSM bool
	LocalHSM map[string]bool
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

type HSMservice func ([]string) ([]haberdasher.HSMresult, *haberdasher.Response, error)	

// BaseConfigLicenseKey - Struct for task definition
type BaseConfigValidateHSM struct {
	configElements []ValidateElement
	hsmService HSMservice
	envVars map[string]string
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
		log.Debug("Could not check env vars for HSM validationn")
	} else {
		t.envVars = envVars
	}

	validLKtoSources, ok := upstream["Base/Config/ValidateLicenseKey"].Payload.(map[string][]string)

	if (!ok || (len(validLKtoSources) == 0)){
		return tasks.Result {
			Status: tasks.None,
			Summary: "No validated license keys to check high security mode against. Task did not run.",
		}
	}

	t.configElements, ok = upstream["Base/Config/Validate"].Payload.([]ValidateElement)

	if (!ok || (len(t.configElements) == 0)){
		return tasks.Result {
			Status: tasks.None,
			Summary: "No New Relic configuration files to check high security mode against. Task did not run.",
		}
	}
	
	licenseKeys := validLKtoSourcesMapToSlice(validLKtoSources)

	lkToHSM, accountHSMerr := t.getAccountHighSecurity(licenseKeys)

	if accountHSMerr != nil {
		return tasks.Result {
			Status: tasks.None,
			Summary: fmt.Sprintf("Error reaching New Relic to check high security mode status on accounts: %s", accountHSMerr.Error()),
		}
	}

	hsmValidations := t.createHSMValidations(lkToHSM, validLKtoSources)

	taskErrorSummary := ""
	taskErrorSummaryPattern := "High Security Mode setting (%v) for account with license key:\n\n%s\n\nmismatches configuration in:\n%s\n\n"

	for _, hsmValidation := range hsmValidations {
		isValid, misMatchedSources := hsmValidation.Validate()
		if !isValid {
			taskErrorSummary = taskErrorSummary + fmt.Sprintf(taskErrorSummaryPattern, hsmValidation.AccountHSM, hsmValidation.LicenseKey, strings.Join(misMatchedSources, "\n"))
		} 
	}

	if len(taskErrorSummary) > 0 {
		return tasks.Result {
			Status: tasks.Failure,
			Summary: taskErrorSummary,
			Payload: hsmValidations,
		}
	}

	return tasks.Result {
		Status: tasks.Success,
		Summary: "High Security Mode setting for accounts associated with found license keys match local configuration.",
		Payload: hsmValidations,
	}

}

func (t BaseConfigValidateHSM) getAccountHighSecurity(licenseKeys []string) (map[string]bool, error) {
	licenseKeyHSM := make(map[string]bool)

	hsmResults, _, err := t.hsmService(licenseKeys)

	if err != nil {
		return licenseKeyHSM, err
	}

	for _, hsmResult := range hsmResults {
		licenseKeyHSM[hsmResult.LicenseKey] = hsmResult.IsEnabled
	}

	return licenseKeyHSM, nil
}


func (t BaseConfigValidateHSM) createHSMValidations(lkToHSM map[string]bool, validLKtoSources map[string][]string) []HSMvalidation {
	hsmValidations := []HSMvalidation{}

	for lk, accountHSM := range lkToHSM {
		
		sources := validLKtoSources[lk]

		sourceConfigElements := t.matchSourcesToConfigElement(sources)
		
		configSourcesToHSM := t.getHSMConfigurations(sourceConfigElements)
		


		hsmValidation := HSMvalidation {
			LicenseKey: lk,
			AccountHSM: accountHSM,
			LocalHSM: configSourcesToHSM,
		}

		hsmValidations = append(hsmValidations, hsmValidation)
	}

	return hsmValidations
}


func (t BaseConfigValidateHSM) matchSourcesToConfigElement(sources []string) []ValidateElement {
	configElements := []ValidateElement{}

	for _, source := range sources {

		if tasks.PosString(licenseKeyEnvVars, source) > -1 {
			continue
		}

		configElement, err := t.matchSourceToConfigElement(source)

		if err != nil {
			continue
		}
		configElements = append(configElements, configElement)
	}

	return configElements
}

func (t BaseConfigValidateHSM) matchSourceToConfigElement(source string) (ValidateElement, error) {
	for _, configElement := range t.configElements {
		fullPath := configElement.Config.FilePath + configElement.Config.FileName
		if fullPath == source {
			return configElement, nil
		}
	}

	return ValidateElement{}, errors.New("No ConfigElement matches source filepath")
}

func (t BaseConfigValidateHSM) getHSMConfigurations(configElements []ValidateElement) map[string]bool {
	configSourcesToHSM := make(map[string]bool)

	hsmEnv, ok := t.envVars[HSM_ENV_VAR]; if ok {
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



func validLKtoSourcesMapToSlice(validLKtoSourcesMap map[string][]string) []string{
	licenseKeys := []string {}

	for k, _ := range validLKtoSourcesMap {
		licenseKeys = append(licenseKeys, k)
	}

	return licenseKeys
}

