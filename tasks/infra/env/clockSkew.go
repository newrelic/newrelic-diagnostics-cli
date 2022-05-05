package env

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var (
	errResponseMissingDateHeader = errors.New("date header not found in collector response")
)

const (
	// Clock Skew over 60 seconds is not manageable
	skewThresholdSeconds = 60

	//OS specific links to provide user if clock skew is detected
	troubleshootingURLwindows = "https://docs.microsoft.com/en-us/windows-server/networking/windows-time-service/configuring-systems-for-high-accuracy?tabs=MinPollInterval"
	troubleshootingURLlinux   = "https://www.maketecheasier.com/sync-linux-time-with-ntp-server/"
)

// InfraEnvClockSkew - This struct defined the sample plugin which can be used as a starting point
type InfraEnvClockSkew struct {
	httpGetter        func(httpHelper.RequestWrapper) (*http.Response, error)
	checkForClockSkew func(time.Time) (bool, int, time.Time)
	runtimeOS         string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraEnvClockSkew) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Env/ClockSkew")
}

// Explain - Returns the help text for each individual task
func (p InfraEnvClockSkew) Explain() string {
	return "Detect if host has clock skew from New Relic collector"
}

// Dependencies - Returns the dependencies for each task.
func (p InfraEnvClockSkew) Dependencies() []string {
	return []string{"Infra/Agent/Connect", "Base/Config/ProxyDetect"}
}

// Execute - Returns result containing the log_file value(s) parsed from any found newrelic-infra.yml files previously collected.
func (p InfraEnvClockSkew) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Infra/Agent/Connect"].Status == tasks.None {
		return tasks.Result{
			Status: tasks.None,
			Summary: "Unable to urls from Infra/Agent/Connect. This task did not run",
		}
	}

	requestURLs, ok := upstream["Infra/Agent/Connect"].Payload.(map[string]string)

	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	var apiEndpoint string

	for _, value := range requestURLs {
		apiEndpoint = value
		break
	}

	if len(apiEndpoint) == 0 {
		apiEndpoint = "https://infra-api.newrelic.com"
	}

	collectorTime, err := p.getCollectorTime(apiEndpoint)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Unable to determine New Relic collector time",
		}
	}

	isClockDiffRelevant, diffSeconds, hostTime := p.checkForClockSkew(collectorTime)

	if isClockDiffRelevant {
		summary := fmt.Sprintf("Detected clock skew of %v seconds between host and New Relic collector. This could lead to chart irregularities:", diffSeconds)
		summary += fmt.Sprintf("\n\t%-16v%s", "Host time:", hostTime.String())
		summary += fmt.Sprintf("\n\t%-16v%s", "Collector time:", collectorTime.String())
		summary += "\nYour host may be affected by clock skew. Please consider using NTP to keep your host clocks in sync."

		result := tasks.Result{
			Status:  tasks.Failure,
			Summary: summary,
		}

		if p.runtimeOS == "windows" {
			result.URL = troubleshootingURLwindows
		} else {
			result.URL = troubleshootingURLlinux
		}

		return result
	}

	return tasks.Result{
		Status:  tasks.Success,
		Summary: "System clock in sync with New Relic collector",
	}
}

func (p InfraEnvClockSkew) getCollectorTime(apiEndpoint string) (time.Time, error) {

	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            apiEndpoint,
		TimeoutSeconds: 30,
	}

	//Make request
	resp, err := p.httpGetter(wrapper)
	if err != nil {
		//Return error result
		return time.Time{}, err
	}
	defer resp.Body.Close()

	//Get server time
	responseHeaderDate, ok := resp.Header["Date"]
	if !ok {
		//Return error here
		return time.Time{}, errResponseMissingDateHeader
	}

	serverTime, err := time.Parse(time.RFC1123, responseHeaderDate[0])
	if err != nil {
		return time.Time{}, err
	}

	return serverTime.In(time.UTC), nil
}

func checkForClockSkew(collectorTime time.Time) (bool, int, time.Time) {
	hostTime := time.Now().In(time.UTC)
	diff := hostTime.Sub(collectorTime)
	diffSeconds := int(math.Abs(diff.Seconds()))
	return (diffSeconds > skewThresholdSeconds), diffSeconds, hostTime
}
