package minion

import (
	"encoding/json"
	"strconv"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// MinionSettings is struct of parses private-location-settings.json values
type MinionSettings struct {
	Key                   string //Private location key
	Hsm                   bool   //High Security Mode (Verified Script Execution) enabled
	HsmPwd                string //High Security Mode passphrase
	Proxy                 string //Proxy fmt: "host:port"
	ProxyAuth             string //Proxy credentials fmt: "username:password"
	ProxyAcceptSelfSigned bool   //Proxy accepts self signed certificate
}

//Expected keys (case sensitive) in private minion settings.json
var expectedKeys = []string{"key", "hsm", "hsmPwd", "proxy", "proxyAuth", "proxyAcceptSelfSigned"}

// MarshalJSON overrides any marshal to json calls for a MinionSettings struct returning a sanitized json payload with sensitive info stripped out.
// Used so ensure we don't include this sensitive info in the output.json
func (ms MinionSettings) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Key                   string
		Hsm                   bool
		Proxy                 string
		ProxyAcceptSelfSigned bool
	}{
		Key:                   ms.Key,
		Hsm:                   ms.Hsm,
		Proxy:                 ms.Proxy,
		ProxyAcceptSelfSigned: ms.ProxyAcceptSelfSigned,
	})
}

// SyntheticsMinionConfigValidate - Validates private minion configuration
type SyntheticsMinionConfigValidate struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p SyntheticsMinionConfigValidate) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Synthetics/Minion/ConfigValidate")
}

// Explain - Returns the help text for each individual task
func (p SyntheticsMinionConfigValidate) Explain() string {
	return "Validate New Relic Synthetics private minion (legacy) configuration"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p SyntheticsMinionConfigValidate) Dependencies() []string {
	return []string{
		"Base/Config/Validate",
		"Synthetics/Minion/Detect",
	}
}

// Execute - Retrieves private-location-settings.json, parses into struct, returns struct as result.payload
func (p SyntheticsMinionConfigValidate) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if upstream["Synthetics/Minion/Detect"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Not running on synthetics minion. No config to parse"
		return result
	}

	if !upstream["Base/Config/Validate"].HasPayload() {
		log.Debug("[SyntheticsMinionConfigValidate] - Failing because Base/Config/Validate failed")
		result.Status = tasks.Failure
		result.Summary = "private-location-settings.json file not found on private minion. \nCheck settings at: http://<MINION_IP_ADDRESS>/setup"
		result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		return result
	}

	//Grab results from Base/Config/Validate
	validatedConfigs, _ := upstream["Base/Config/Validate"].Payload.([]config.ValidateElement)

	var validateResult config.ValidateElement

	//Search Base/Config/Validate results for settings json.
	for _, config := range validatedConfigs {
		if config.Config.FileName == "private-location-settings.json" {
			if config.Config.FilePath == "/opt/newrelic/synthetics/.newrelic/synthetics/minion/" {
				log.Debug("Found private-location-settings.json setting validateResult")
				validateResult = config
			} else {
				log.Debug("[SyntheticsMinionConfigValidate] - Failing because private-location-settings.json found in incorrect location:" + config.Config.FilePath)
				result.Status = tasks.Failure
				result.Summary = "private-location-settings.json found in incorrect location. \nCheck settings at: http://<MINION_IP_ADDRESS>/setup\nLocation:" + config.Config.FilePath
				result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
				return result
			}
		}
	}

	// If Base/Config/Validate didn't find the settings file..
	if validateResult.Config.FilePath == "" {
		log.Debug("[SyntheticsMinionConfigValidate] - Failing because Base/Config/Validate didn't find the private-location-settings.json settings file")
		result.Status = tasks.Failure
		result.Summary = "private-location-settings.json file not found on private minion. \nCheck settings at: http://<MINION_IP_ADDRESS>/setup"
		result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		return result
	}

	// Convert validateResult.ParsedResult -> MinionSettings struct
	settings, invalidKeys := parseToStruct(validateResult.ParsedResult)

	log.Debug("SyntheticsMinionConfigValidate settings parsed: ")
	log.Debug(settings)

	// Validate key length
	log.Debug("Validating private location key length...")
	if !isKeyValid(settings.Key) {
		log.Debug("[SyntheticsMinionConfigValidate] - Failing because private-location-settings.json contains invalid private minion location key length: " + strconv.Itoa(len(settings.Key)))
		result.Status = tasks.Failure
		result.Summary = "Configured private location key length: " + strconv.Itoa(len(settings.Key)) + ". Expected length: 36\nCheck settings at: http://<MINION_IP_ADDRESS>/setup"
		result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		return result
	}

	// Validate HSM config
	log.Debug("Validating verified script execution configuration...")
	if settings.Hsm && settings.HsmPwd == "" {
		result.Status = tasks.Failure
		result.Summary = "Verified Script Execution is enabled, but no password is provided"
		result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		return result
	}
	log.Debug(validateResult.Config.FileName + " parsed and validated.")

	// Throw warning if invalid keys were found, but we otherwise were successful in the above tests.
	if len(invalidKeys) > 0 {
		result.Status = tasks.Warning
		result.Summary = "private-location-settings.json contains invalid key/value pairs: " + strings.Join(invalidKeys, ", ") + "\nValid keys (case sensitive):" + strings.Join(expectedKeys, ", ")
		result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		result.Payload = settings
		return result
	}

	result.Status = tasks.Success
	result.Summary = "Private minion settings retrieved and validated."
	result.Payload = settings

	return result
}

// parseToStruct - Takes parsed base/config/validate result and returns a MinionSetting struct
func parseToStruct(in tasks.ValidateBlob) (MinionSettings, []string) {

	unparsed := walkSyntheticsConfig(in)
	var settingsStruct MinionSettings
	var invalidKeys []string

	//Keys in private-minion-settings.json are case sensitive
	for k, v := range unparsed {
		switch k {
		case expectedKeys[0]:
			settingsStruct.Key = v.(string)
		case expectedKeys[1]:
			settingsStruct.Hsm = v.(bool)
		case expectedKeys[2]:
			settingsStruct.HsmPwd = v.(string)
		case expectedKeys[3]:
			settingsStruct.Proxy = v.(string)
		case expectedKeys[4]:
			settingsStruct.ProxyAuth = v.(string)
		case expectedKeys[5]:
			settingsStruct.ProxyAcceptSelfSigned = v.(bool)
		default:
			log.Debug("Unknown settings key found:", k)
			invalidKeys = append(invalidKeys, k)
		}
	}
	return settingsStruct, invalidKeys
}

// isKeyValid - Determines if private location key is valid. Currently just checks length
func isKeyValid(key string) bool {
	return len(key) == 36
}

func walkSyntheticsConfig(in tasks.ValidateBlob) map[string]interface{} {

	unparsed := make(map[string]interface{})

	for _, child := range in.Children {
		unparsed[child.Key] = child.RawValue
	}
	return unparsed
}
