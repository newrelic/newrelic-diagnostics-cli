package agent

import (
	"regexp"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BrowserAgentDetect - This struct defined the sample plugin which can be used as a starting point
type BrowserAgentDetect struct {
}

// BrowserAgentPayload - formatted data for json output
type BrowserAgentPayload struct {
	AppReporting      string
	AgentVersion      string
	BrowserLicenseKey string
	BrowserLoader     string
	AgentType         string
	TransactionName   string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BrowserAgentDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Browser/Agent/Detect")
}

// Explain - Returns the help text for each individual task
func (t BrowserAgentDetect) Explain() string {
	return "Detect New Relic Browser agent from provided URL"
	// gorun -t browser/agent/detect -v -o Browser/Agent/Detect.url=http://localhost:3000
}

// Dependencies - Returns the dependencies for ech task.
func (t BrowserAgentDetect) Dependencies() []string {
	return []string{
		"Browser/Agent/GetSource",
	}
}

// Execute - The core work within each task
func (t BrowserAgentDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	//log.Debug(upstream)
	// first type assert upstream back into a payload we can work with
	source, ok := upstream["Browser/Agent/GetSource"].Payload.(BrowserAgentSourcePayload)
	if ok {
		log.Debug("correct type")
		//log.Debug(source)
	}
	log.Debug("source of loaders", len(source.Loader))
	if len(source.Loader) == 0 {
		log.Debug("No loaders detected, setting to failed")
		result.Status = tasks.Failure
		result.Summary = "Failed to detect browser agent"
		result.URL = "https://docs.newrelic.com/docs/browser/new-relic-browser/installation/install-new-relic-browser-agent"
		return result
	} else if len(source.Loader) > 2 {
		log.Debug("More than 1 browser agent detected")
		result.Status = tasks.Warning
		result.Summary = "More than one browser agent detected, please check to ensure only one browser agent is configured per page"
	} else {
		result.Status = tasks.Success
		result.Summary = "Found values for browser agent"
	}

	var payload BrowserAgentPayload
	payload.AgentVersion = getAgentVersion(source.Loader)
	payload.AppReporting = getAppReporting(source.Loader)
	payload.BrowserLicenseKey = getBrowserKey(source.Loader)
	payload.BrowserLoader = getBrowserLoader(source.Loader)
	payload.AgentType, payload.TransactionName = getAgentType(source.Loader)
	log.Debug("payload is ", payload)

	result.Payload = payload

	return result
}

func getAppReporting(scripts []string) string {
	regex := regexp.MustCompile(`applicationID[":\s]*"([0-9,]*)",`)

	for _, loader := range scripts {
		//log.Debug("loader is ", loader)
		for _, value := range regex.FindAllStringSubmatch(loader, -1) {
			log.Debug("App reporting ID found: ", value[1])
			return value[1]
		}
	}
	log.Debug("reporting app ID not found, returning empty value")
	return ""
}

func getAgentVersion(scripts []string) string {
	regex := regexp.MustCompile(`agent:"js-agent\.newrelic\.com.*-([0-9]*)\.min\.js"`)

	for _, loader := range scripts {
		for _, value := range regex.FindAllStringSubmatch(loader, -1) {
			log.Debug("agent version found: ", value[1])
			return value[1]
		}
	}
	log.Debug("browser key not found, returning empty value")
	return ""
}

func getBrowserKey(scripts []string) string {
	regex, _ := regexp.Compile(`licenseKey[":\s]*"([a-z0-9]*)"`)

	for _, loader := range scripts {
		for _, value := range regex.FindAllStringSubmatch(loader, -1) {
			log.Debug("browser license key found", value[1])
			return value[1]
		}

	}
	log.Debug("browser key not found, returning empty value")
	return ""
}

//detects loader in use
func getBrowserLoader(scripts []string) string {
	spa := regexp.MustCompile(`agent:"js-agent\.newrelic\.com/nr-spa-[0-9]*\.min\.js"`)
	pro := regexp.MustCompile(`\(NREUM={}\)\).loader_config|window.onerror|UncaughtException`)

	for _, loader := range scripts {
		// First check for SPA agent
		if spa.MatchString(loader) {
			log.Debug("Found SPA agent")
			return "SPA"
		} else if pro.MatchString(loader) { //Now check for Pro vs Lite since SPA was not found

			log.Debug("Found Pro loader")
			return "Pro"
		}

	}
	log.Debug("Pro loader not found, returning Lite")
	return "Lite"

}

// detects copy/paste vs injected, returns type and Transaction name (if injected)
func getAgentType(scripts []string) (string, string) {
	regex, _ := regexp.Compile(`transactionName"[:\s]"*([0-9a-zA-Z=]*)"`)

	for _, loader := range scripts {
		for _, value := range regex.FindAllStringSubmatch(loader, -1) {
			log.Debug("agent type detected as injected with transactionName", value[1])
			return "injected", value[1]
		}
		log.Debug("transactionName not found, checking for applicationTime")
		if regexp.MustCompile(`applicationTime`).MatchString(loader) {
			log.Debug("Found injected browser agent via applicationTime")
			return "injected", ""
		}

	}
	log.Debug("agent type detected as copy/paste")
	return "copy/paste", ""
}
