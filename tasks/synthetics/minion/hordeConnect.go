package minion

import (
	"io/ioutil"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// HTTPResponse is struct of parsed http response data used to evaluate horde connection
type HTTPResponse struct {
	ResponseCode int
	ResponseBody string
}

type SyntheticsMinionHordeConnect struct { // This defines the task itself and should be named according to the standard CategorySubcategoryTaskname in camelcase
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p SyntheticsMinionHordeConnect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Synthetics/Minion/HordeConnect")
}

// Explain - Returns the help text for each individual task
func (p SyntheticsMinionHordeConnect) Explain() string {
	return "Check network connection to New Relic Synthetics horde endpoint for private minions (legacy)"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p SyntheticsMinionHordeConnect) Dependencies() []string {
	return []string{
		"Synthetics/Minion/ConfigValidate",
		"Base/Config/ProxyDetect",
		"Synthetics/Minion/Detect",
	}
}

// Execute - Uses parsed private location settings key to perform a simple HTTP request to horde
func (p SyntheticsMinionHordeConnect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	if upstream["Synthetics/Minion/Detect"].Status != tasks.Success {
		result.Status = tasks.None
		result.Summary = "Not running on private minion"
		return result
	}

	settings, ok := upstream["Synthetics/Minion/ConfigValidate"].Payload.(MinionSettings)
	if ok {
		log.Debug("correct type: MinionSettings")
	}
	log.Debug(settings, "minion settings found")
	log.Debug("Making horde HTTP response with key: " + settings.Key)

	hordeResponse, err := hordeRequest(settings.Key)

	log.Debug("Request complete, evaluating response code. Full response:", hordeResponse)
	switch hordeResponse.ResponseCode {
	case 200:
		if err != nil {
			result.Status = tasks.Warning
			result.Summary = "Succesful connection to synthetics-horde.nr-data.net with current settings, but unable to parse response body"
			result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks#synthetics-private"
			return result
		}
		result.Status = tasks.Success
		result.Summary = "Succesful connection to synthetics-horde.nr-data.net with current settings"

	case 403:
		if upstream["Synthetics/Minion/ConfigValidate"].Status != tasks.Success {
			result.Status = tasks.Failure
			result.Summary = "Private minion has not been configured."
			result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		} else {
			result.Status = tasks.Failure
			result.Summary = "Connection to synthetics-horde.nr-data.net returned 403 Forbidden. Confirm location key is correct (\"" + settings.Key + "\") at: http://<MINION_IP_ADDRESS>/setup"
			result.URL = "https://docs.newrelic.com/docs/synthetics/new-relic-synthetics/private-locations/install-configure-private-minions#configure"
		}
	case -7:
		result.Status = tasks.Failure
		result.Summary = "Unable to complete request to synthetics-horde.nr-data.net: " + err.Error()
		result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks#synthetics-private"
	default:
		result.Status = tasks.Failure
		result.Summary = "Expected 200 response. Received: " + strconv.Itoa(hordeResponse.ResponseCode)
		result.URL = "https://docs.newrelic.com/docs/apm/new-relic-apm/getting-started/networks#synthetics-private"
	}

	return result
}

// hordeRequest - Performs GET request to https://synthetics-horde.nr-data.net/api/v1.0/config using provided private location key
func hordeRequest(privateLocationKey string) (HTTPResponse, error) {
	var httpResponse HTTPResponse

	headers := make(map[string]string)
	headers["X-API-Key"] = privateLocationKey

	log.Debug("Attempting connection to: synthetics-horde.nr-data.net (HTTPS) using location key: " + privateLocationKey)

	wrapper := httpHelper.RequestWrapper{
		Method:  "GET",
		URL:     "https://synthetics-horde.nr-data.net/api/v1.0/config",
		Headers: headers,
	}

	resp, err := httpHelper.MakeHTTPRequest(wrapper)

	// Check for connection error
	if err != nil {
		httpResponse.ResponseCode = -7
		return httpResponse, err
	}
	defer resp.Body.Close()

	httpResponse.ResponseCode = resp.StatusCode
	responseBody, err := ioutil.ReadAll(resp.Body)
	// Check for error parsing response body
	if err != nil {
		return httpResponse, err
	}
	httpResponse.ResponseBody = string(responseBody)

	return httpResponse, nil
}
