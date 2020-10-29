package agent

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// InfraAgentConnect - This struct tests the connector to Infrastructure
type InfraAgentConnect struct {
	httpGetter requestFunc
}

type requestFunc func(wrapper httpHelper.RequestWrapper) (*http.Response, error)

//RequestResult - contains HTTP response and error status data, Id is to distinguish requests from many, in this case region
type RequestResult struct {
	Id         string
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

	regionURLs := map[string]string{
		"us01": "https://infra-api.newrelic.com",
		"eu01": "https://infra-api.eu01.nr-data.net",
	}

	var result tasks.Result
	var requestResults map[string]RequestResult
	requestURLs := make(map[string]string)

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Infrastructure Agent config not present."
		return result
	}

	regions, ok := upstream["Base/Config/RegionDetect"].Payload.([]string)

	if (!ok) || len(regions) == 0 {
		requestURLs = regionURLs
	} else {
		for _, region := range regions {
			requestURLs[region] = regionURLs[region]
		}
	}

	requestResults = makeRequests(requestURLs, p.httpGetter)
	summary, status := validateResponses(requestResults)

	return tasks.Result{
		Status:  status,
		Summary: summary,
		URL:     "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks",
		Payload: requestURLs,
	}
}

func makeRequests(urls map[string]string, HTTPagent requestFunc) map[string]RequestResult {
	var requestResults = make(map[string]RequestResult)

	for id, url := range urls {
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
			body, err = ioutil.ReadAll(response.Body)

			if err != nil {
				err = errors.New("Read error - There was an issue reading the body: " + err.Error())
			}
			response.Body.Close()
		}

		requestResults[id] = RequestResult{
			Id:         id,
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
	for id, requestResult := range requestResults {
		log.Debug(id)
		if requestResult.Err != nil {
			summary += "\nThere was an error connecting to " + requestResult.URL
			summary += "\nPlease check network and proxy settings and try again or see -help for more options."
			summary += "\nError = " + requestResult.Err.Error()
			return summary, tasks.Failure
		} else if requestResult.StatusCode == 404 {
			log.Debug("Successfully connected")
			summary += fmt.Sprintf(" Successfully connected to %s Infrastructure API endpoint.", id)
		} else {
			summary += fmt.Sprintf(" Was not able to connect to the Infrastructure API endpoint. Unexpected Response: %d %s ", requestResult.StatusCode, requestResult.Status)
			return summary, tasks.Failure
		}
	}
	return summary, tasks.Success
}
