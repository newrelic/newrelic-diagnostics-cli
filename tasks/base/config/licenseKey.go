package config

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var licenseKeyConfigNames = []string{
	"license_key",      //infra, java, python, ruby
	"licenseKey",       //dotnet, node
	"-licenseKey",      //dotnetcore
	"newrelic.license", //PHP
}

var licenseKeyEnvVars = []string{
	"NRIA_LICENSE_KEY",      //infra
	"NEW_RELIC_LICENSE_KEY", //all other agents
}

var licenseKeySysProp = "-Dnewrelic.config.license_key"

type LicenseKey struct {
	Value  string
	Source string
}

// BaseConfigLicenseKey - Struct for task definition
type BaseConfigLicenseKey struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseConfigLicenseKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/LicenseKey")
}

// Explain - Returns the help text for each individual task
func (t BaseConfigLicenseKey) Explain() string {
	return "Determine New Relic license key(s)"
}

// Dependencies - Returns the dependencies for each task.
func (t BaseConfigLicenseKey) Dependencies() []string {
	return []string{
		"Base/Config/Validate",
		"Base/Env/CollectEnvVars",
		"Base/Env/CollectSysProps",
	}
}

// Execute - The core work within each task
func (t BaseConfigLicenseKey) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var (
		licenseKeys           []LicenseKey
		licenseKeysFromConfig []LicenseKey
		licenseKeysFromEnv    []LicenseKey
		licenseKeyFromSysProp LicenseKey //there can only be one system property for a license key
	)

	configElements, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement)
	if ok {
		licenseKeysFromConfig = getLicenseKeysFromConfig(configElements, licenseKeyConfigNames)
		licenseKeys = append(licenseKeys, licenseKeysFromConfig...)
	}

	envVarValues, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if ok {
		licenseKeysFromEnv = getLicenseKeysFromEnv(envVarValues)
		licenseKeys = append(licenseKeys, licenseKeysFromEnv...)
	}

	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		procIDSysProps, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			licenseKeyFromSysProp = getLicenseKeysFromSysProps(procIDSysProps)
			if licenseKeyFromSysProp.Value != "" {
				licenseKeys = append(licenseKeys, licenseKeyFromSysProp)
			}
		}
	}

	if len(licenseKeys) == 0 {

		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "No New Relic licenses keys were found. Please ensure a license key is set in your New Relic agent configuration or environment.",
			URL:     "https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/install-configure/configure-agent",
			Payload: licenseKeys,
		}
	}

	deDupedLicenseKeys := dedupeLicenseKeySlice(licenseKeys)

	if len(deDupedLicenseKeys) > 1 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Multiple license keys detected:\n" + summarizeLicenseKeySources(deDupedLicenseKeys) + envOverrideMessage(deDupedLicenseKeys),
			URL:     "https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/install-configure/configure-agent",
			Payload: licenseKeys,
		}
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: strconv.Itoa(len(deDupedLicenseKeys)) + " unique New Relic license key(s) found." + envOverrideMessage(licenseKeys),
		Payload: licenseKeys,
	}
}

func envOverrideMessage(licenseKeys []LicenseKey) string {
	envMessage := ""
	sysPropMessage := ""

	for _, licenseKey := range licenseKeys {
		if licenseKey.Source == "NRIA_LICENSE_KEY" {
			envMessage = envMessage + fmt.Sprintf("\n     '%s' from '%s' will be used by New Relic Infrastructure Agent", licenseKey.Value, licenseKey.Source)
		} else if licenseKey.Source == "NEW_RELIC_LICENSE_KEY" {
			envMessage = envMessage + fmt.Sprintf("\n     '%s' from '%s' will be used by New Relic APM Agents", licenseKey.Value, licenseKey.Source)
		} else if licenseKey.Source == "-Dnewrelic.config.license_key" {
			sysPropMessage = fmt.Sprintf("\n     '%s' from '%s' will be used by New Relic APM Agents", licenseKey.Value, licenseKey.Source)
		}
	}

	if len(envMessage) > 0 {
		return envMessage
	}
	return sysPropMessage
}

func getLicenseKeysFromEnv(envVariables map[string]string) []LicenseKey {
	results := []LicenseKey{}

	for envVar, v := range envVariables {
		for _, expectedEnvKey := range licenseKeyEnvVars {
			if envVar == expectedEnvKey {
				licenseKey := LicenseKey{
					Value:  v,
					Source: envVar,
				}
				results = append(results, licenseKey)
			}
		}
	}
	return results
}

func getLicenseKeysFromSysProps(sysProps []tasks.ProcIDSysProps) LicenseKey {

	for i := 0; i < len(sysProps); i++ {
		sysPropMap := sysProps[i].SysPropsKeyToVal
		lk, isPresent := sysPropMap[licenseKeySysProp]
		if isPresent {
			return LicenseKey{
				Value:  lk,
				Source: licenseKeySysProp,
			}
		}
	}
	return LicenseKey{}
}

func getLicenseKeysFromConfig(configElements []ValidateElement, configKeys []string) []LicenseKey {
	result := []LicenseKey{}
	for _, configKey := range configKeys {
		for _, configFile := range configElements {

			foundKeys := configFile.ParsedResult.FindKey(configKey)
			keyDedupe := make(map[string]bool)

			for _, key := range foundKeys {
				// We only want to create unique licenseKey instances from a single config file
				// This helps avoid the same license key being reported multiple times from same source
				// Due to environment (prod/staging/developement etc) configuration
				if _, ok := keyDedupe[key.Value()]; ok {
					continue
				}

				keyDedupe[key.Value()] = true

				licenseKey := LicenseKey{
					Value:  key.Value(),
					Source: configFile.Config.FilePath + configFile.Config.FileName,
				}

				if detectEnvLicenseKey(licenseKey.Value) {
					retrievedKey, err := retrieveEnvLicenseKey(licenseKey.Value, os.Getenv)
					if err != nil {
						continue // Unable to retrieve, skip this one
					} else {
						licenseKey.Value = retrievedKey
					}
				}
				//Commenting out for now as sanitization may undermine our validation downstream
				//sanitizedKey := sanitizeLicenseKey(licenseKeyVal)
				result = append(result, licenseKey)
			}
		}
	}
	return result
}

// detectEnvLicenseKey attempts to detect configuration file references to environment variables
func detectEnvLicenseKey(key string) bool {
	re := regexp.MustCompile(`<%= ENV|process.env`)
	return re.MatchString(key)
}

// EnvReader is meant for dependency injecting into retrieveEnvLicenseKey
type EnvReader func(string) string

// retrieveEnvLicenseKey parses references to license keys stored as environment
// variables and fishes them out of the ENV. It takes os.Getenv as an argument
func retrieveEnvLicenseKey(keyEnvReference string, readEnv EnvReader) (string, error) {
	rubyReg := regexp.MustCompile(`ENV\[["']([^'"]+)["']\]`)
	nodeReg := regexp.MustCompile(`process.env.(.+)`)

	matches := rubyReg.FindStringSubmatch(keyEnvReference)

	if len(matches) == 0 {
		matches = nodeReg.FindStringSubmatch(keyEnvReference)
	}

	if len(matches) > 0 {
		envName := matches[1]
		envVal := readEnv(envName)
		if envVal != "" {
			return envVal, nil
		} else {
			return "", errors.New("licenseKey var " + keyEnvReference + " not found in environment")
		}
	}
	return "", errors.New("unable to parse key from provided input")
}

func summarizeLicenseKeySources(licenseKeys []LicenseKey) string {
	summary := ""

	for _, licenseKey := range licenseKeys {
		summary = summary + fmt.Sprintf("     '%s' from '%s'\n", licenseKey.Value, licenseKey.Source)
	}

	return summary
}

func dedupeLicenseKeySlice(licenseKeys []LicenseKey) []LicenseKey {
	deDuped := []LicenseKey{}
	uniques := map[string]LicenseKey{}

	for _, licenseKey := range licenseKeys {

		//If key already exists in dedupe map skip if its not an env variable
		//We want to give precedence to env variable and system property over config file
		if lk, present := uniques[licenseKey.Value]; present {
			if (tasks.PosString(licenseKeyEnvVars, licenseKey.Source) == -1) && licenseKeySysProp != licenseKey.Source {
				continue //skip value of config files
			}
			if licenseKey.Source == licenseKeySysProp && lk.Source == "NEW_RELIC_LICENSE_KEY" {
				continue //skip value of system property if the already present lk.Source was from an env var
			}
		}
		uniques[licenseKey.Value] = licenseKey
	}

	for _, v := range uniques {
		deDuped = append(deDuped, v)
	}
	return deDuped
}
