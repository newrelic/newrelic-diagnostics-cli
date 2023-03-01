package agent

import (
	"errors"
	"fmt"
	"io"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/exp/maps"
)

// InfraAgentConnect - This struct tests the connector to Infrastructure
type InfraAgentConnect struct {
	httpGetter requestFunc
}

// RequestResult - contains HTTP response and error status data, Id is to distinguish requests from many, in this case region
type RequestResult struct {
	URL        string
	Status     string
	StatusCode int
	Body       string
	Err        error
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraAgentConnect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Agent/Connect")
}

// Explain - Returns the help text for each individual task
func (p InfraAgentConnect) Explain() string {
	return "Check network connection to New Relic Infrastructure collector endpoint"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraAgentConnect) Dependencies() []string {
	return []string{
		"Base/Config/ProxyDetect",
		"Base/Config/RegionDetect",
		"Infra/Config/Agent",
	}
}

// Execute - The core work within each task
func (p InfraAgentConnect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	domains := map[string]string{
		"us01": ".newrelic.com",
		"eu01": ".eu.newrelic.com",
	}
	endpoints := []string{
		"infra-api",
		"identity-api",
		"infrastructure-command-api",
		"log-api",
		"metric-api",
	}

	var result tasks.Result
	var requestResults map[string]RequestResult
	var requestURLs []string

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Infrastructure Agent config not present."
		return result
	}

	regions, ok := upstream["Base/Config/RegionDetect"].Payload.([]string)

	if (!ok) || len(regions) == 0 {
		requestURLs = buildRequestURLs(endpoints, maps.Values(domains)...)
	} else {
		for _, region := range regions {
			requestURLs = append(requestURLs, buildRequestURLs(endpoints, domains[region])...)
		}
	}

	requestResults = makeRequests(requestURLs, p.httpGetter)
	summary, status := validateResponses(requestResults)

	return tasks.Result{
		Status:  status,
		Summary: summary,
		URL:     "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks/#infrastructure",
		Payload: requestURLs,
	}
}

func buildRequestURLs(endpoints []string, domains ...string) []string {
	var urls []string
	for _, domain := range domains {
		for _, endpoint := range endpoints {
			url := "https://" + endpoint + domain
			urls = append(urls, url)
		}
	}
	return urls
}

func makeRequests(urls []string, HTTPagent requestFunc) map[string]RequestResult {
	var requestResults = make(map[string]RequestResult)

	for _, url := range urls {
		var body []byte
		var status string
		var statusCode int

		wrapper := httpHelper.RequestWrapper{
			Method:         "GET",
			URL:            url,
			TimeoutSeconds: 30,
		}

		response, err := HTTPagent(wrapper)

		if err == nil {
			status = response.Status
			statusCode = response.StatusCode
			body, err = io.ReadAll(response.Body)

			if err != nil {
				err = errors.New("read error - There was an issue reading the body: " + err.Error())
			}
			response.Body.Close()
		}

		requestResults[url] = RequestResult{
			URL:        url,
			Status:     status,
			StatusCode: statusCode,
			Body:       string(body),
			Err:        err,
		}

	}
	return requestResults
}

func validateResponses(requestResults map[string]RequestResult) (string, tasks.Status) {
	var summary string
	for url, requestResult := range requestResults {
		log.Debug(url)
		if requestResult.Err != nil {
			summary += "\nThere was an error connecting to " + url
			summary += "\nPlease check network and proxy settings and try again or see -help for more options."
			summary += "\nError = " + requestResult.Err.Error()
			return summary, tasks.Failure
		} else if requestResult.StatusCode == 404 {
			log.Debug("Successfully connected")
			summary += fmt.Sprintf(" Successfully connected to %s.", url)
		} else {
			summary += fmt.Sprintf(" Was not able to connect to %s. Unexpected Response: %d %s ", url, requestResult.StatusCode, requestResult.Status)
			return summary, tasks.Failure
		}
	}
	return summary, tasks.Success
}
