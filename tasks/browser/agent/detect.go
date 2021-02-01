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
	// ./nrdiag -browser-url http://thecustomers-website-url --suites browser
}

// Dependencies - Returns the dependencies for each task.
func (t BrowserAgentDetect) Dependencies() []string {
	return []string{
		"Browser/Agent/GetSource",
	}
}

// Execute - The core work within each task
func (t BrowserAgentDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Browser/Agent/GetSource"].Status == tasks.Failure || upstream["Browser/Agent/GetSource"].Status == tasks.Error || upstream["Browser/Agent/GetSource"].Status == tasks.None {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "This tasks did not run because the previous task 'Browser/Agent/GetSource' either did not run or was not successful.",
		}
	}

	pageSourcePayload, ok := upstream["Browser/Agent/GetSource"].Payload.(BrowserAgentSourcePayload)

	if !ok {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "This task was unable to run because we ran into a Type Assertion error from the upstream task Browser/Agent/GetSource",
		}
	}

	if len(pageSourcePayload.Loader) > 1 {
		log.Debug("More than 1 browser agent detected")
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "More than one browser agent detected, please check to ensure only one browser agent is configured per page",
		}
	}

	var payload BrowserAgentPayload
	payload.AgentVersion = getAgentVersion(pageSourcePayload.Loader)
	payload.AppReporting = getAppReporting(pageSourcePayload.Loader)
	payload.BrowserLicenseKey = getBrowserKey(pageSourcePayload.Loader)
	payload.BrowserLoader = getBrowserLoader(pageSourcePayload.Loader)
	payload.AgentType, payload.TransactionName = getAgentType(pageSourcePayload.Loader)
	log.Debug("payload is ", payload)

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "Found values for browser agent",
		Payload: payload,
	}
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
