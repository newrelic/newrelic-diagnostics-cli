package collector

import (
	"io"
	"reflect"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BaseCollectorConnectUS - This task connects to collector.newrelic.com and reports the status
type BaseCollectorConnectUS struct {
	upstream   map[string]tasks.Result
	httpGetter requestFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p BaseCollectorConnectUS) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Collector/ConnectUS")
}

// Explain - Returns the help text for each individual task
func (p BaseCollectorConnectUS) Explain() string {
	return "Check network connection to New Relic US region collector endpoint"
}

// Dependencies - This task depends on Base/Config/ProxyDetect
func (p BaseCollectorConnectUS) Dependencies() []string {
	return []string{
		"Base/Config/ProxyDetect", //we are not using the payload of this task, but we want to make sure that it was already detected and set before running any HTTP request
		"Base/Config/RegionDetect",
	}
}

// Execute - Attempts to connect to the US collector endpoint
func (p BaseCollectorConnectUS) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	p.upstream = upstream

	url := "https://collector.newrelic.com/status/mongrel"

	// Was the task not explicitly provided on -t ?
	if !config.Flags.IsForcedTask(p.Identifier().String()) {
		result := p.prepareEarlyResult()
		// Early result received, bailing
		if !(reflect.DeepEqual(result, tasks.Result{})) {
			return result
		}
	}

	// Make request
	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            url,
		TimeoutSeconds: 30,
	}
	resp, err := p.httpGetter(wrapper)

	if err != nil {
		// HTTP error
		return p.prepareCollectorErrorResult(err)
	}

	defer resp.Body.Close()

	// Parse HTTP response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		// body parse error result
		return p.prepareResponseErrorResult(err, strconv.Itoa(resp.StatusCode))
	}

	//Successful request, return result based on status code
	return p.prepareResult(string(body), strconv.Itoa(resp.StatusCode))

}

func (p BaseCollectorConnectUS) prepareEarlyResult() tasks.Result {
	var result tasks.Result
	regions, ok := p.upstream["Base/Config/RegionDetect"].Payload.([]string)
	if ok {
		// If this region was not in the non-empty list of detected region, return early.
		// If no regions were detected, we run all collector connect checks.
		if !tasks.StringInSlice("us01", regions) && len(regions) > 0 {
			result.Status = tasks.None
			result.Summary = "US Region not detected, skipping US collector connect check"
			return result
		}
	}
	return result
}

func (p BaseCollectorConnectUS) prepareCollectorErrorResult(e error) tasks.Result {
	var result tasks.Result
	if e == nil {
		return result
	}
	result.Status = tasks.Failure
	result.Summary = "There was an error connecting to collector.newrelic.com (US Region)"
	result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
	result.Summary += "\nError = " + e.Error()
	result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks"

	return result
}

func (p BaseCollectorConnectUS) prepareResponseErrorResult(e error, statusCode string) tasks.Result {
	var result tasks.Result
	if e == nil {
		return result
	}
	result.Status = tasks.Warning
	result.Summary = "There was an issue reading the body while connecting to the US Region collector."
	result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
	result.Summary += "Error = " + e.Error()
	result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks"

	return result
}

func (p BaseCollectorConnectUS) prepareResult(body, statusCode string) tasks.Result {
	var result tasks.Result

	if statusCode == "404" && body == "{}" {
		log.Debug("Successfully connected (US Region)")
		result.Status = tasks.Success
		result.Summary = "Successfully connected to collector.newrelic.com (US Region)"
	} else {
		log.Debug("Unsuccessful response received from collector.newrelic.com.")
		log.Debug("Body:", body)
		result.Status = tasks.Warning
		result.Summary = "The connection to collector.newrelic.com (US Region) was not successful."
		result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
		result.Summary += "\nResponse Body: " + body
		result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks"
	}

	return result
}
