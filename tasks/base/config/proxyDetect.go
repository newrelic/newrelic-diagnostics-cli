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
	proxyURL      string
	proxyScheme   string
	processID     int32
	proxySource   string
}

var proxyEnvVarsKeys = []string{
	"NRIA_PROXY",               //Infra: https://proxy_user.access_10@proxy_01:1080
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

var nrProxyConfigs = []ProxyConfig{ //This is a slice with all the possible entries that we will search through
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
	proxyScheme:   "proxy_scheme",
}

var dotnetProxyKeys = ProxyConfig{
	proxyHost:     "-host",
	proxyPort:     "-port",
	proxyUser:     "-user",
	proxyPassword: "-password",
}

var infraOrNodeProxyKeys = ProxyConfig{
	proxyURL: "proxy", //Infra agent and Node support the single proxy item names (which take as a value a url: 'http://user:pass@10.0.0.1:8000/') or the standard config keys so we look for both. This setting will override the other standardProxyKeys in the config file
}

// minion has combined auth key "proxyAuth".
var minionProxyKeys = ProxyConfig{
	proxyHost: "proxy",
	proxyUser: "proxyAuth",
}

var phpProxyKeys = ProxyConfig{
	proxyHost: "newrelic.daemon.proxy",
}

var httpsProxyKeys = []string{
	"HTTP_PROXY",
	"HTTPS_PROXY",
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

	//check if the customer has http_proxy or https_proxy in their environment. If they don't, later we'll set the env var using the proxy values found via newrelic proxy settings; this is env var will allow us to connect nrdiag to newrelic and upload their data into a ticket
	httpsProxyKey, httpsProxyVal := checkForHttpORHttpsProxies()

	validations, ok := upstream["Base/Config/Validate"].Payload.([]ValidateElement) //data coming from config files found

	if ok {
		proxyConfig, multipleProxyErr := getProxyConfig(validations, options, upstream)

		if multipleProxyErr != nil {
			return tasks.Result{
				Status:  tasks.Warning,
				Summary: "We had difficulties retrieving proxy settings from your New Relic config file: " + multipleProxyErr.Error(),
				Payload: proxyConfig,
			}
		}

		if (proxyConfig != ProxyConfig{}) {
			proxyURL := ""
			if (proxyConfig.proxyHost != "") || (proxyConfig.proxyURL != "") {
				proxyURL = proxyURL + setProxyURL(proxyConfig)
			}
			if httpsProxyKey == "" {
				os.Setenv("HTTP_PROXY", proxyURL) //Set this env var temporarily to be used by: https://github.com/newrelic/newrelic-diagnostics-cli/blob/main/processOptions.go#L39
			}
			log.Debug(proxyConfig)
			return tasks.Result{
				Status:  tasks.Success,
				Summary: fmt.Sprintf("We have successfully detected a proxy URL set %s via New Relic proxy settings using %s\n", proxyURL, proxyConfig.proxySource),
				Payload: proxyConfig,
			}
		}
	}

	if httpsProxyKey != "" {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("We have detected a proxy set via %s: %s\nThough this may be a valid configuration for you app and it is supported by New Relic Infinite Tracing, keep in mind that New Relic agents support their own specific proxy settings.", httpsProxyKey, httpsProxyVal),
			URL:     "https://docs.newrelic.com/docs/using-new-relic/cross-product-functions/install-configure/configure-agent",
		}
	}

	return tasks.Result{
		Status:  tasks.None,
		Summary: "No proxy server settings found for this app",
	}
}

func checkForHttpORHttpsProxies() (string, string) {
	for _, httpsProxyKey := range httpsProxyKeys {
		httpsProxyVal := os.Getenv(httpsProxyKey)
		if httpsProxyVal != "" {
			return httpsProxyKey, httpsProxyVal
		}
	}
	return "", ""
}

func getProxyConfig(validations []ValidateElement, options tasks.Options, upstream map[string]tasks.Result) (ProxyConfig, error) {

	proxyConfig := findProxyValuesFromEnvVars(upstream)

	for _, validation := range validations { // Go through each config file validation to see if the proxy is configured anywhere in there or to at least find out which agent are we dealing with based on the file extension
		if filepath.Ext(validation.Config.FileName) != ".ini" && (proxyConfig != ProxyConfig{}) {
			return proxyConfig, nil //early exit because env vars take precendence for all agents except python. PHP does not use env vars
		}
		if filepath.Ext(validation.Config.FileName) == ".yml" && (validation.Config.FileName != "newrelic-infra.yml") {
			//applicable only to Java not Ruby:
			proxyConfig := findProxyValuesFromSysProps(upstream)
			if (proxyConfig != ProxyConfig{}) {
				return proxyConfig, nil //early exit because system properties take precendence over config file
			}
			//Check for proxy values in yml file, applicable to both Java and Ruby
			proxyConfig = findProxyValuesFromYmlFile(validation, options)

			if (proxyConfig != ProxyConfig{}) {
				proxyConfig.proxySource = validation.Config.FilePath + validation.Config.FileName
			}
			return proxyConfig, nil
		}
		// now let's look into infra, .NET, PHP, minion and other standard settings
		for _, proxyConfigKeys := range nrProxyConfigs {
			proxyConfig, multipleProxyConfigs := findProxyValuesFromConfigFile(proxyConfigKeys, validation)
			log.Debug("ProxyConfig found through config file: ", proxyConfig)
			log.Debug("Detected multipleProxyConfigs: ", multipleProxyConfigs)
			if proxyConfig.proxyHost != "" {
				return proxyConfig, nil //return as soon as we have a match
			}
			if len(multipleProxyConfigs) > 1 {
				return proxyConfig, errors.New("multiple proxy values found within a config File")
			}
		}
	} // end of iterating through config validations
	return proxyConfig, nil
}

func setProxyURL(proxy ProxyConfig) string {
	//this single setting is only available for a couple of agents and they ovewrite other proxy settings
	if proxy.proxyURL != "" {
		return proxy.proxyURL
	}
	//build the URL by putting together all the proxy setting values they have used
	var proxyURL string
	if proxy.proxyUser != "" {
		proxyURL += proxy.proxyUser
		//No password found case for combined auth in single key (e.g. private minion's proxyAuth)
		if proxy.proxyPassword != "" {
			proxyURL += ":" + proxy.proxyPassword
		}
		proxyURL += "@"
	}

	proxyURL += proxy.proxyHost

	if proxy.proxyPort != "" {
		proxyURL += ":" + proxy.proxyPort
	}
	//Some customers will preprend a protocol to their host configuration. So want to avoid building a url such as this one: http://https://myuser:mypassword@myproxy.mycompany.com:8080
	if strings.Contains(proxyURL, "http") {
		return proxyURL
	}

	if proxy.proxyScheme != "" { //setting option only for python and java
		proxyURL = proxy.proxyScheme + "://" + proxyURL
		return proxyURL
	}
	//default to http
	proxyURL = "http://" + proxyURL
	log.Debug("Setting proxy via detected config to", proxyURL)
	return proxyURL
}

func findProxyValuesFromEnvVars(upstream map[string]tasks.Result) ProxyConfig {
	proxyConfig := ProxyConfig{}

	if upstream["Base/Env/CollectEnvVars"].Status == tasks.Info {
		envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
		if ok {
			for _, proxyEnvVarKey := range proxyEnvVarsKeys {
				proxyEnvVarVal, isPresent := envVars[proxyEnvVarKey]
				if isPresent {
					lowerCaseEnvVar := strings.ToLower(proxyEnvVarKey)
					if strings.Contains(lowerCaseEnvVar, "host") {
						proxyConfig.proxyHost = proxyEnvVarVal
					} else if strings.Contains(lowerCaseEnvVar, "port") {
						proxyConfig.proxyPort = proxyEnvVarVal
					} else if strings.Contains(lowerCaseEnvVar, "user") {
						proxyConfig.proxyUser = proxyEnvVarVal
					} else if strings.Contains(lowerCaseEnvVar, "pass") { //should match pass or password
						proxyConfig.proxyPassword = proxyEnvVarVal
					} else if strings.Contains(lowerCaseEnvVar, "scheme") {
						proxyConfig.proxyScheme = proxyEnvVarVal
					} else {
						proxyConfig.proxyURL = proxyEnvVarVal
					}
				}
			}
			if (proxyConfig != ProxyConfig{}) {
				proxyConfig.proxySource = "New Relic Environment variables"
			}
		}
	}
	return proxyConfig
}

func findProxyValuesFromSysProps(upstream map[string]tasks.Result) ProxyConfig {
	proxyConfig := ProxyConfig{}

	if upstream["Base/Env/CollectSysProps"].Status == tasks.Info {
		processes, ok := upstream["Base/Env/CollectSysProps"].Payload.([]tasks.ProcIDSysProps)
		if ok {
			for _, process := range processes {
				for _, proxySysPropKey := range proxySysPropsKeys {
					proxySysPropVal, isPresent := process.SysPropsKeyToVal[proxySysPropKey]
					if isPresent {
						if strings.Contains(proxySysPropKey, "host") {
							proxyConfig.proxyHost = proxySysPropVal
						} else if strings.Contains(proxySysPropKey, "port") {
							proxyConfig.proxyPort = proxySysPropVal
						} else if strings.Contains(proxySysPropKey, "user") {
							proxyConfig.proxyUser = proxySysPropVal
						} else {
							proxyConfig.proxyPassword = proxySysPropVal
						}
					}
				}
				if (proxyConfig != ProxyConfig{}) { // we want at least one of the proxyConfig fiels to be populated
					proxyConfig.processID = process.ProcID
					proxyConfig.proxySource = "system properties"
					return proxyConfig //early exit, we got the proxyConfig info that we needed
				}
			} //end of iterating through java processes
		}
	}
	return proxyConfig
}

func findProxyValuesFromYmlFile(fileElement ValidateElement, options tasks.Options) ProxyConfig {
	proxyValues, proxyMap := findProxyValuesFromConfigFile(standardProxyKeys, fileElement)
	if proxyValues.proxyHost != "" {
		//This indicates only one environment section found in our yml config file and we did not have to bother creating a proxyMap
		return proxyValues
	}
	//More than one set of values found so we're going to default to trying to find them based on production
	var envOverride string
	if options.Options["environment"] != "" { //is an nrdiag cmdline option to set their environment to production, staging, etc.
		envOverride = (string(options.Options["environment"]))
	} else {
		envOverride = "production"
	}

	var proxyConfig ProxyConfig
	// Set host
	path := "/" + envOverride + "/" + standardProxyKeys.proxyHost
	proxyConfig.proxyHost = proxyMap[path]

	// Set Port
	path = "/" + envOverride + "/" + standardProxyKeys.proxyPort
	proxyConfig.proxyPort = proxyMap[path]

	// Set User
	path = "/" + envOverride + "/" + standardProxyKeys.proxyUser
	proxyConfig.proxyUser = proxyMap[path]

	// Set Pass or Password
	path = "/" + envOverride + "/" + standardProxyKeys.proxyPassword
	if val, ok := proxyMap[path]; ok {
		proxyConfig.proxyPassword = val
	} else {
		path = "/" + envOverride + "/" + standardProxyKeys.proxyPassword + "word"
		proxyConfig.proxyPassword = proxyMap[path]
	}
	proxyConfig.proxyPassword = proxyMap[path]

	// Set Scheme
	path = "/" + envOverride + "/" + standardProxyKeys.proxyScheme
	proxyConfig.proxyScheme = proxyMap[path]

	log.Debug("returning found proxy values of", proxyConfig)
	return proxyConfig
}

func findProxyValuesFromConfigFile(proxyConfigKeys ProxyConfig, fileElement ValidateElement) (ProxyConfig, map[string]string) {
	var proxyConfig ProxyConfig
	var proxyKeyPathToVal = make(map[string]string)
	//This will return possibly multiple key instances if the config file has an instance per environment (common, production, staging and development)
	hostValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyHost) //we make the assumption that proxyConfigKeys will have at least a proxyHost field because that is a field all different agent proxy settings have in common
	for _, v := range hostValues {
		if len(hostValues) == 1 {
			proxyConfig.proxyHost = v.Value()
		} else {
			proxyKeyPathToVal[v.PathAndKey()] = v.Value() //Ex: proxyKeyPathToVal["/Common/proxy_host"]
		}
	}

	if proxyConfigKeys.proxyPort != "" {
		portValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyPort)
		for _, v := range portValues {
			if len(portValues) == 1 {
				proxyConfig.proxyPort = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
	}

	if proxyConfigKeys.proxyUser != "" {
		userValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyUser)
		for _, v := range userValues {
			if len(userValues) == 1 {
				proxyConfig.proxyUser = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
	}

	if proxyConfigKeys.proxyPassword != "" {
		passValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyPassword)
		for _, v := range passValues {
			if len(passValues) == 1 {
				proxyConfig.proxyPassword = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
		// Do second search for password because java and ruby match in everything except on this proxy key: proxy_pass vs proxy_password
		passwordValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyPassword + "word")
		for _, v := range passwordValues {
			if len(passwordValues) == 1 {
				proxyConfig.proxyPassword = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
	}

	if proxyConfigKeys.proxyURL != "" {
		proxyURLValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyURL)
		for _, v := range proxyURLValues {
			if len(proxyURLValues) == 1 {
				proxyConfig.proxyURL = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
	}

	if proxyConfigKeys.proxyScheme != "" {
		proxySchemeValues := fileElement.ParsedResult.FindKey(proxyConfigKeys.proxyScheme)
		for _, v := range proxySchemeValues {
			if len(proxySchemeValues) == 1 {
				proxyConfig.proxyScheme = v.Value()
			} else {
				proxyKeyPathToVal[v.PathAndKey()] = v.Value()
			}
		}
	}

	log.Debug("combined proxy map is", proxyKeyPathToVal)
	return proxyConfig, proxyKeyPathToVal
}
