package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

//ProxyConfig represents an specific proxy server settings/configuration. The processID field mostly serves a purpose for agents like the Java Agent
type ProxyConfig struct {
	proxyHost     string
	proxyPort     string
	proxyUser     string
	proxyPassword string
	processID     int32
}

var proxyEnvVarsKeys = []string{
	"NEW_RELIC_PROXY_HOST",     //Java, Node, Python, Ruby
	"NEW_RELIC_PROXY_PASS",     //Node, Python, Ruby
	"NEW_RELIC_PROXY_PORT",     //Java, Node, Python, Ruby
	"NEW_RELIC_PROXY_URL",      //Node
	"NEW_RELIC_PROXY_USER",     //Java, Node, Python, Ruby
	"NEW_RELIC_PROXY_SCHEME",   //Java, Python : Setting proxy_scheme: "https" will allow the agent to connect through proxies using the HTTPS scheme
	"NEW_RELIC_PROXY_PASSWORD", //Java
}

var proxySysPropsKeys = []string{
	"-Dnewrelic.config.proxy_host", //=somehost.com
	"-Dnewrelic.config.proxy_port",
	"-Dnewrelic.config.proxy_user", //=username
	"-Dnewrelic.config.proxy_password",
}

var proxyConfigKeys = []ProxyConfig{ //This is a slice with all the possible entries that we will search through
	minionProxyKeys,
	standardProxyKeys,
	dotnetProxyKeys,
	infraOrNodeProxyKeys,
	phpProxyKeys,
}

var standardProxyKeys = ProxyConfig{ //This is the values used by Java, Ruby, Node, Python
	proxyHost:     "proxy_host",
	proxyPort:     "proxy_port",
	proxyUser:     "proxy_user",
	proxyPassword: "proxy_pass",
}

var dotnetProxyKeys = ProxyConfig{
	proxyHost:     "-host",
	proxyPort:     "-port",
	proxyUser:     "-user",
	proxyPassword: "-password",
}

var infraOrNodeProxyKeys = ProxyConfig{
	proxyHost: "proxy", //Infra agent and Node support the single proxy item names (which take as a value a url: 'http://user:pass@10.0.0.1:8000/') or the standard config keys so we look for both. This setting will override the other standardProxyKeys in the config file
}

// minion has combined auth key "proxyAuth".
var minionProxyKeys = ProxyConfig{
	proxyHost: "proxy",
	proxyUser: "proxyAuth",
}

var phpProxyKeys = ProxyConfig{
	proxyHost: "newrelic.daemon.proxy",
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
		"Base/Env/CollectEnvVars",
		"Base/Env/CollectSysProps",
	}
}

// Execute - This task will search for config files based on the string array defined and walk the directory tree from the working directory searching for additional matches
func (p BaseConfigProxyDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	//Check options to see if proxy was configured via command line or env

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

	validations, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement)
	if ok {
		log.Debug("Base/Config/Validate payload is correct type")
	}
	// Loop through validations to see if the proxy is configured anywhere in there or to at least find out which agent are we dealing with based on the filename
	for _, validation := range validations { //We'll exit the iteration as soon as we set a proxy
		proxyConfigs, err := findProxyValues(validation, envOverride, upstream)
		if err != nil {
			log.Debug("Error getting proxy. Error was ", err)
			result.Status = tasks.Warning
			result.Summary = "Error retrieving proxy from config file" + err.Error()
		}
		log.Debug("execute proxyValue is", proxyConfigs)

		proxySuccessSummary := ""
		for _, proxyConfig := range proxyConfigs {
			if proxyConfig.proxyHost != "" {
				proxyURL := setProxyFromConfig(proxyConfig)
				proxySuccessSummary = proxySuccessSummary + fmt.Sprintf("Set proxy to: %s\n", proxyURL)
			}
		}
		if len(proxySuccessSummary) > 0 {
			return tasks.Result{
				Status:  tasks.Success,
				Summary: proxySuccessSummary,
				Payload: proxyConfigs,
			}
		}

		//Check proxy configuration, if set in multiple locations and NOT matching, flag a warning
	}

	return result
}

func setProxyFromConfig(proxy ProxyConfig) string {
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

func findProxyValues(fileElement ValidateElement, envOverride string, upstream map[string]tasks.Result) ([]ProxyConfig, error) {
	//So first we build a map so we can track the results, using the proxy_host as the string of the map itself
	var defaultProxyConfig ProxyConfig

	if filepath.Ext(fileElement.Config.FileName) == ".yml" && fileElement.Config.FileName != "newrelic-infra.yml" {
		//applicable only to Java
		proxyConfigs := findValueFromSysProps(upstream) //the only reason I am return an slice of ProxyConfigs is because I am assumming that in some crazy world you can have different JVM processes, each running with its own proxy server settings
		if len(proxyConfigs) > 0 {
			return proxyConfigs, nil
		}
		//Check for proxy values in yml file, applicable to Java and Ruby
		proxyConfig, err := findValueFromYml(fileElement, envOverride)
		if err != nil {
			return []ProxyConfig{proxyConfig}, err
		}
		return []ProxyConfig{proxyConfig}, nil
	}
	// now let's look into infra, .NET, PHP, minion and other standard settings
	for _, proxyKey := range proxyConfigKeys {
		proxyConfig, proxyMap := findValueFromKeys(proxyKey, fileElement)
		log.Debug("Detected proxyConfig", proxyConfig)
		log.Debug("Detected proxyMap", proxyMap)
		if len(proxyMap) > 1 {
			return []ProxyConfig{defaultProxyConfig}, errors.New("multiple proxy values")
		}
		if proxyConfig.proxyHost != "" {
			return []ProxyConfig{proxyConfig}, nil //return as soon as we have a match
		}
	}

	log.Debug("returning found proxy values of ", defaultProxyConfig)
	return []ProxyConfig{defaultProxyConfig}, nil
}

func findValueFromSysProps(upstream map[string]tasks.Result) []ProxyConfig {
	proxyConfigs := []ProxyConfig{}

	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		processes, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			for _, process := range processes {
				config := ProxyConfig{}
				for _, proxySysPropKey := range proxySysPropsKeys {
					proxySysPropVal, isPresent := process.SysPropsKeyToVal[proxySysPropKey]
					if isPresent {
						if strings.Contains(proxySysPropKey, "host") {
							config.proxyHost = proxySysPropVal
						} else if strings.Contains(proxySysPropKey, "port") {
							config.proxyPort = proxySysPropVal
						} else if strings.Contains(proxySysPropKey, "user") {
							config.proxyUser = proxySysPropVal
						} else {
							config.proxyPassword = proxySysPropVal
						}
					}
				}
				if (config != ProxyConfig{}) { // we want at least one of the proxyConfig fiels to be populated
					config.processID = process.ProcID
					proxyConfigs = append(proxyConfigs, config)
				}
			}
		}
		return proxyConfigs
	}

	return proxyConfigs
}

func findValueFromYml(fileElement ValidateElement, envOverride string) (ProxyConfig, error) {
	var returnProxyValues ProxyConfig

	proxyValues, proxyMap := findValueFromKeys(standardProxyKeys, fileElement)
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

func findValueFromKeys(searchKeys ProxyConfig, fileElement ValidateElement) (ProxyConfig, map[string]string) {
	var proxyValues ProxyConfig
	var combinedMap = make(map[string]string)

	//This will return 1 or more found keys that we need to map to the slice
	hostValues := fileElement.ParsedResult.FindKey(searchKeys.proxyHost)

	// call out to build the map from the value

	for _, v := range hostValues {
		if len(hostValues) == 1 {
			proxyValues.proxyHost = v.Value()
		} else {
			combinedMap[v.PathAndKey()] = v.Value()
		}
	}
	if searchKeys.proxyPort != "" {
		portValues := fileElement.ParsedResult.FindKey(searchKeys.proxyPort)
		for _, v := range portValues {
			if len(portValues) == 1 {
				proxyValues.proxyPort = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
	}
	if searchKeys.proxyUser != "" {
		userValues := fileElement.ParsedResult.FindKey(searchKeys.proxyUser)
		for _, v := range userValues {
			if len(userValues) == 1 {
				proxyValues.proxyUser = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
	}
	if searchKeys.proxyPassword != "" {
		passValues := fileElement.ParsedResult.FindKey(searchKeys.proxyPassword)
		for _, v := range passValues {
			if len(passValues) == 1 {
				proxyValues.proxyPassword = v.Value()
			} else {
				combinedMap[v.PathAndKey()] = v.Value()
			}
		}
		// Do second search for password because java and ruby are dumb and almost match
		passwordValues := fileElement.ParsedResult.FindKey(searchKeys.proxyPassword + "word")
		for _, v := range passwordValues {
			if len(passwordValues) == 1 {
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

func convertToMapStruct(incomingMap map[string]string, proxyKey *ProxyConfig) map[string]ProxyConfig {
	var returnedConfig = make(map[string]ProxyConfig)

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

func setProxyMap(hostValue map[string]string, searchKey string, found *map[string]ProxyConfig) {
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
