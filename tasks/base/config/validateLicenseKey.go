package config

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/internal/haberdasher"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var (
	licenseRegex = regexp.MustCompile("^[[:alnum:]]+$")
)

// BaseConfigValidateLicenseKey - Struct for task definition
type BaseConfigValidateLicenseKey struct {
	validateAgainstAccount func(map[string][]string) (map[string][]string, map[string][]string, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseConfigValidateLicenseKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/ValidateLicenseKey")
}

// Explain - Returns the help text for each individual task
func (p BaseConfigValidateLicenseKey) Explain() string {
	return "Determine New Relic license key(s) have a proper format and that they are valid for a determined account"
}

// Dependencies - Returns the dependencies for each task.
func (p BaseConfigValidateLicenseKey) Dependencies() []string {
	return []string{
		"Base/Config/LicenseKey",
	}
}

// Execute - The core work within each task
func (p BaseConfigValidateLicenseKey) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	licenseKeys, ok := upstream["Base/Config/LicenseKey"].Payload.([]LicenseKey)

	if !ok {
		log.Debug(`upstream["Base/Config/LicenseKey"].Payload.([]LicenseKey) failed data type assertion in validateLicenseKey task`)
	}
	if len(licenseKeys) == 0 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No New Relic licenses keys were found. Task to validate license key did not run",
		}
	}
	//we'll start by validating license key format and collecting the proper result summary based on our findings
	validatedLKCounter := 0
	validFormatLKToSources := make(map[string][]string)
	var successSummary, warningSummary, failureSummary string

	//For all cases, except Python, env vars will override config settings. If we find env vars let's skip validating config files because they will contain either empty strings or unmodified sample license keys that will fail validation.
	envVarsFound := findLKFromEnvVarSources(licenseKeys)
	if len(envVarsFound) > 0 {
		for _, envVar := range envVarsFound {
			isEnvVarFormatValid, errMsg := checkEnvVarFormat(envVar)
			validatedLKCounter++
			if isEnvVarFormatValid {
				sources, isPresent := validFormatLKToSources[envVar.Value]
				if isPresent {
					sources = append(sources, envVar.Source)
					validFormatLKToSources[envVar.Value] = sources
					continue
				}
				validFormatLKToSources[envVar.Value] = []string{envVar.Source}
			} else {
				failureSummary += errMsg
			}
		}
	} else {
		uniqueLKToSources := dedupeLicenseKeys(licenseKeys) //we should expect some repeats because when customers are using multiple NR products, most probably they will use the same license key for each product's config file.
		for lk, sources := range uniqueLKToSources {
			isConfigFormatValid, errMsg := checkConfigFormat(lk, sources)
			validatedLKCounter++
			if isConfigFormatValid {
				validFormatLKToSources[lk] = sources
			} else {
				failureSummary += errMsg
			}
		}
	}

	//Only if we have collected a license key with a valid format, then we can move into checking that the customer's account agrees that this is a valid key
	var resultsPayload map[string][]string
	if len(validFormatLKToSources) > 0 {
		validAccountLKToSources, invalidAccountLKToSources, err := p.validateAgainstAccount(validFormatLKToSources)

		if err != nil {
			for lk, sources := range validFormatLKToSources {
				warningSummary += fmt.Sprintf("The license key found in %s has a valid New Relic format: %s. \nThough we ran into an error (%s) while trying to validate against your account. Only if your agent is reporting an 'Invalid license key' log entry, reach out to New Relic Support.\n\n", strings.Join(sources, ",\n "), lk, err.Error())
			}
			resultsPayload = validFormatLKToSources
		}
		if len(invalidAccountLKToSources) > 0 {
			for lk, sources := range invalidAccountLKToSources {
				failureSummary += fmt.Sprintf("The license key found in %s did not pass our validation check when verifying against your account:\n%s\nIf your agent is reporting an 'Invalid license key' log entry, please reach out to New Relic Support.\n\n", strings.Join(sources, ",\n "), lk)
			}
		}
		if len(validAccountLKToSources) > 0 {
			for lk, sources := range validAccountLKToSources {
				successSummary += fmt.Sprintf("The license key found in %s passed our validation check when verifying against your account:\n %s"+"\n", strings.Join(sources, ",\n "), lk)
				if isRegionEU(lk) {
					successSummary += fmt.Sprintf(`Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, please verify that your agent version is compatible with New Relic license keys that are 'region aware': https://docs.newrelic.com/docs/using-new-relic/welcome-new-relic/get-started/our-eu-us-region-data-centers. Reach out to Support if this is not the issue.` + "\n")
				} else {
					successSummary += fmt.Sprintf(`Note: If your agent is reporting an 'Invalid license key' log entry for this valid License key, reach out to New Relic support to verify any issues in our end.` + "\n\n")
				}
			}
			resultsPayload = validAccountLKToSources
		}
	}

	resultStatus := determineResultStatus(successSummary, warningSummary, failureSummary)

	return tasks.Result{
		Status:  resultStatus,
		Summary: fmt.Sprintf("We validated %s license key(s):\n"+successSummary+failureSummary+warningSummary, strconv.Itoa(validatedLKCounter)),
		Payload: resultsPayload,
	}
}

func findLKFromEnvVarSources(licenseKeys []LicenseKey) []LicenseKey {
	var envVars []LicenseKey
	for _, lk := range licenseKeys {
		for _, envVar := range licenseKeyEnvVars {
			if lk.Source == envVar {
				envVars = append(envVars, lk)
			}
		}
	}
	return envVars
}

func checkEnvVarFormat(licenseKey LicenseKey) (bool, string) {
	if licenseKeyUsingQuotes(licenseKey.Value) {
		errMsg := fmt.Sprintf(`Using quotes around %s may cause inconsistent behavior. We highly recommend removing those quotes, and running the `+tasks.ThisProgramFullName+` again.`+"\n\n", licenseKey.Source)
		return false, errMsg
	}
	if isFormatValid(strings.TrimSpace(licenseKey.Value)) {
		return true, ""
	}
	errMsg := fmt.Sprintf("The license key found in %s does not have a valid format: %s. \nThe NR license key is 40 alphanumeric characters. \nReview this documentation to make sure that you have the proper format of a New Relic Personal API key: \nhttps://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys\n\n", licenseKey.Source, licenseKey.Value)
	return false, errMsg
}

func checkConfigFormat(licenseKey string, sources []string) (bool, string) {
	sanitizedLicenseKey := sanitizeLicenseKey(licenseKey)
	if isFormatValid(sanitizedLicenseKey) {
		return true, ""
	}
	errMsg := fmt.Sprintf("The license key found in %s does not have a valid format: %s. \nThe NR license key is 40 alphanumeric characters. \nReview this documentation to make sure that you have the proper format of a New Relic Personal API key: \nhttps://docs.newrelic.com/docs/apis/get-started/intro-apis/types-new-relic-api-keys\n\n", strings.Join(sources, ",\n "), sanitizedLicenseKey)
	return false, errMsg
}

// sanitizeLicenseKey strips whitespace and single and double quotes from a string
func sanitizeLicenseKey(key string) string {
	sanitizedKey := strings.TrimSpace(key)
	re := regexp.MustCompile(`['" ]`)
	sanitizedKey = re.ReplaceAllString(sanitizedKey, ``)
	return sanitizedKey
}

func licenseKeyUsingQuotes(licenseKey string) bool {
	matched, err := regexp.Match(`['" ]`, []byte(licenseKey))
	if err != nil {
		log.Debug(err)
	}
	return matched
}

// IsValid return true if license is in valid format and incorrect length.
func isFormatValid(licenseKey string) bool {
	return licenseRegex.MatchString(licenseKey) && len(licenseKey) == 40
}

func validateAgainstAccount(LKToSources map[string][]string) (map[string][]string, map[string][]string, error) {

	var licenseKeys []string
	for lk := range LKToSources {
		licenseKeys = append(licenseKeys, lk)
	}
	results, _, err := haberdasher.DefaultClient.Tasks.ValidateLicenseKeys(licenseKeys)
	if err != nil {
		return map[string][]string{}, map[string][]string{}, err
	}
	validLicenseKeys := map[string][]string{}
	invalidLicenseKeys := map[string][]string{}
	for _, result := range results {
		if result.IsValid {
			validLicenseKeys[result.LicenseKey] = LKToSources[result.LicenseKey]
		} else {
			invalidLicenseKeys[result.LicenseKey] = LKToSources[result.LicenseKey]
		}
	}
	return validLicenseKeys, invalidLicenseKeys, nil
}

// IsRegionEU returns true if license region is EU.
func isRegionEU(license string) bool {
	r := getRegion(license)
	// only EU supported
	if len(r) > 1 && r[:2] == "eu" {
		return true
	}
	return false
}

// GetRegion returns license region or empty if none.
func getRegion(licenseKey string) string {
	//regionLicenseRegex is defined in regionDetect.go
	matches := regionLicenseRegex.FindStringSubmatch(licenseKey)
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}

func determineResultStatus(successSummary, warningSummary, failureSummary string) tasks.Status {
	if len(failureSummary) > 0 {
		return tasks.Failure
	}

	if len(warningSummary) > 0 {
		return tasks.Warning
	}

	return tasks.Success
}

func dedupeLicenseKeys(licenseKeys []LicenseKey) map[string][]string {
	uniqueLicenseKeys := map[string][]string{}

	for _, lk := range licenseKeys {
		sources, isPresent := uniqueLicenseKeys[lk.Value]

		if isPresent {
			sources = append(sources, lk.Source)
			uniqueLicenseKeys[lk.Value] = sources
			continue
		}
		uniqueLicenseKeys[lk.Value] = []string{lk.Source}
	}
	return uniqueLicenseKeys
}
