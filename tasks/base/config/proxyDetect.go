package config

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type proxyConfig struct {
	proxyHost     string
	proxyPort     string
	proxyUser     string
	proxyPassword string
}

// BaseConfigProxyDetect - Primary task to search for and find config file. Will optionally take command line input as source
type BaseConfigProxyDetect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseConfigProxyDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/ProxyDetect")
}

// Explain - Returns the help text for each individual task
func (p BaseConfigProxyDetect) Explain() string {
	return "Determine and use configured proxy for New Relic agent"
}

// Dependencies - No dependencies since this is generally one of the first tasks to run
func (p BaseConfigProxyDetect) Dependencies() []string {
	// no dependencies!
	return []string{
		"Base/Config/Validate",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseConfigProxyDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result

	if envProxy := os.Getenv("HTTP_PROXY"); envProxy != "" {
		log.Debug("Proxy set via ENV variable. Task did not run.")
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Proxy set via ENV variable. Task did not run.",
		}
	}
	var envOverride string
	if options.Options["environment"] != "" {
		envOverride = (string(options.Options["environment"]))
		log.Debug("Setting environment override to :", envOverride)
	}

	validations, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement) //This is a type assertion to cast my upstream results back into data I know the structure of and can now work with. In this case, I'm casting it back to the []validateElements{} I know it should return
	if ok {
		log.Debug("Base/Config/Validate payload is correct type")
		//		log.Debug(configs) //This may be useful when debugging to log the entire results to the screen
	}

	//Check options to see if proxy was configured via command line or env

	// Loop through validations to see if the proxy is configured anywhere

	for _, validation := range validations {

		proxyValue, err := findProxyValues(validation, envOverride)
		if err != nil {
			log.Debug("Error getting proxy. Error was ", err)
			result.Status = tasks.Warning
			result.Summary = "Error retrieving proxy from config file" + err.Error()
		}
		log.Debug("execute proxyValue is", proxyValue)
		if proxyValue.proxyHost != "" {
			proxyURL := setProxyFromConfig(proxyValue)
			result.Status = tasks.Success
			result.Summary = "Set proxy to: " + proxyURL + " from file: " + validation.Config.FilePath + validation.Config.FileName
			result.Payload = proxyValue
			return result //Exit as soon as we set a proxy
		}

	}

	//Check proxy configuration, if set in multiple locations and NOT matching, flag a warning

	//Set proxy if not set via env or command line

	//Report proxy used to payload and summary, regardless if successful or not

	return result
}

var standardProxyKeys = proxyConfig{ //This is the values used by Java, Ruby
	proxyHost:     "proxy_host",
	proxyPort:     "proxy_port",
	proxyUser:     "proxy_user",
	proxyPassword: "proxy_pass",
}

var dotnetProxyKeys = proxyConfig{
	proxyHost:     "-host",
	proxyPort:     "-port",
	proxyUser:     "-user",
	proxyPassword: "-password",
}

var infraProxyKeys = proxyConfig{
	proxyHost: "proxy", //node also supports the single proxy item names or the standard config keys so we look for both
}

// minion has combined auth key "proxyAuth".
var minionProxyKeys = proxyConfig{
	proxyHost: "proxy",
	proxyUser: "proxyAuth",
}

var phpProxyKeys = proxyConfig{
	proxyHost: "newrelic.daemon.proxy",
}

var proxyConfigKeys = []proxyConfig{ //This is a slice with all the possible entries that we will search through
	minionProxyKeys,
	standardProxyKeys,
	dotnetProxyKeys,
	infraProxyKeys,
	phpProxyKeys,
}

func setProxyFromConfig(proxy proxyConfig) string {
	//Build proxy url. Some agents and customers will preprend a protocol to their host configuration. So want to avoid building a url such as this one: http://https://myuser:mypassword@myproxy.mycompany.com:8080
	var proxyURL string
	defaultProxyProtocol := "http://"
	//Proxy user credential found?
	if proxy.proxyUser != "" {

		proxyURL += proxy.proxyUser

		//No pass found case for combined auth in single key (e.g. private minion's proxyAuth)
		if proxy.proxyPassword != "" {
			proxyURL += ":" + proxy.proxyPassword
		}
		proxyURL += "@"
	}

	proxyURL += proxy.proxyHost

	if proxy.proxyPort != "" {
		proxyURL += ":" + proxy.proxyPort
	}

	if !(strings.Contains(proxyURL, "http")) {
		proxyURL = defaultProxyProtocol + proxyURL
	}
	//This is how we set the proxy and all default connections will use it
	log.Debug("Setting proxy via detected config to", proxyURL)
	os.Setenv("HTTP_PROXY", proxyURL)
	return proxyURL
}

func findProxyValues(input ValidateElement, envOverride string) (proxyConfig, error) {
	//So first we build a map so we can track the results, using the proxy_host as the string of the map itself
	var defaultProxyConfig proxyConfig

	if filepath.Ext(input.Config.FileName) == ".yml" && input.Config.FileName != "newrelic-infra.yml" {
		proxyConfig, err := findValueFromYml(input, envOverride)
		if err != nil {
			return proxyConfig, err
		}
		return proxyConfig, nil
	}

	for _, proxyKey := range proxyConfigKeys {

		proxyValues, proxyMap := findValueFromKeys(proxyKey, input)
		log.Debug("Detected proxyValues", proxyValues)
		log.Debug("Detected proxyMap", proxyMap)
		if len(proxyMap) > 1 {
			return defaultProxyConfig, errors.New("multiple proxy values")
		}
		if proxyValues.proxyHost != "" {
			return proxyValues, nil //return as soon as we have a match
		}
	}

	log.Debug("returning found proxy values of ", defaultProxyConfig)
	return defaultProxyConfig, nil
}

func findValueFromYml(input ValidateElement, envOverride string) (proxyConfig, error) {
	var returnProxyValues proxyConfig

	proxyValues, proxyMap := findValueFromKeys(standardProxyKeys, input)
	if proxyValues.proxyHost != "" {
		//This indicates only one section found in our yml config file
		return proxyValues, nil
	} //More than one set of values found so we're going to default to trying to find them based on production

	if envOverride == "" {
		envOverride = "production"
	}

	// Set host
	path := "/" + envOverride + "/" + standardProxyKeys.proxyHost
	returnProxyValues.proxyHost = proxyMap[path]

	// Set Port
	path = "/" + envOverride + "/" + standardProxyKeys.proxyPort
	returnProxyValues.proxyPort = proxyMap[path]

	// Set User
	path = "/" + envOverride + "/" + standardProxyKeys.proxyUser
	returnProxyValues.proxyUser = proxyMap[path]
	// Set Pass
	path = "/" + envOverride + "/" + standardProxyKeys.proxyPassword
	if val, ok := proxyMap[path]; ok {
		returnProxyValues.proxyPassword = val
	} else {
		path = "/" + envOverride + "/" + standardProxyKeys.proxyPassword + "word"
		returnProxyValues.proxyPassword = proxyMap[path]
	}

	returnProxyValues.proxyPassword = proxyMap[path]
	log.Debug("returning found proxy values of", returnProxyValues)
	return returnProxyValues, nil
}

func findValueFromKeys(searchKeys proxyConfig, input ValidateElement) (proxyConfig, map[string]string) {
	var proxyValues proxyConfig
	var combinedMap = make(map[string]string)

	//This will return 1 or more found keys that we need to map to the slice
	hostValue := input.ParsedResult.FindKey(searchKeys.proxyHost)

	// call out to build the map from the value

	for _, v := range hostValue {
		if len(hostValue) == 1 {
			proxyValues.proxyHost = v.Value()
		} else {
			combinedMap[v.PathAndKey()] = v.Value()
		}
	}
	if searchKeys.proxyPort != "" {
		portValue := input.ParsedResult.FindKey(searchKeys.proxyPort)
		for _, v := range portValue {
			if len(portValue) == 1 {
				proxyValues.proxyPort = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
	}
	if searchKeys.proxyUser != "" {
		userValue := input.ParsedResult.FindKey(searchKeys.proxyUser)
		for _, v := range userValue {
			if len(userValue) == 1 {
				proxyValues.proxyUser = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
	}
	if searchKeys.proxyPassword != "" {
		passValue := input.ParsedResult.FindKey(searchKeys.proxyPassword)
		for _, v := range passValue {
			if len(passValue) == 1 {
				proxyValues.proxyPassword = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
		// Do second search for password because java and ruby are dumb and almost match
		passwordValue := input.ParsedResult.FindKey(searchKeys.proxyPassword + "word")
		for _, v := range passwordValue {
			if len(passwordValue) == 1 {
				proxyValues.proxyPassword = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}

	}


	log.Debug("combined proxy map is", combinedMap)
	log.Debug("proxyValues found are", proxyValues)
	return proxyValues, combinedMap
}

func convertToMapStruct(incomingMap map[string]string, proxyKey *proxyConfig) map[string]proxyConfig {
	var returnedConfig = make(map[string]proxyConfig)

	//We're going to default to assuming environment of prod
	//support an override to pass in a different environment sectrion to activate but otherwise just use prod

	log.Debug("Proxy key is ", proxyKey.proxyHost)

	for key := range incomingMap {
		if strings.Contains(key, proxyKey.proxyHost) {
			split := strings.Split(key, "/")
			var mapKey string
			for i := 0; i <= len(split)-2; i++ { //minus two because it starts at 0 and we don't want the last segment
				mapKey = mapKey + "/" + split[i]
			}
			log.Debug("total mapKey is ", mapKey)
		}
	}

	return returnedConfig
}

func setProxyMap(hostValue map[string]string, searchKey string, found *map[string]proxyConfig) {
	var prefixes []string

	for key := range hostValue {
		log.Debug("Key found is", key)
		split := strings.Split(key, "/")
		//Now we reconstruct the key to match port with the host
		var mapKey string
		for i := 0; i <= len(split)-2; i++ { //minute two because it starts at 0 and we don't want the last segment
			mapKey = mapKey + "/" + split[i]
		}
		prefixes = append(prefixes, mapKey)

	}

}
