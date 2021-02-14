// Class defines structs and methods for timing the run of the Diagnostics CLI integration tests, and sending test result/timing data to Insights.
// This expects INSIGHTS_API_KEY and INSIGHTS_ACCOUNT_ID environment variables to be set, otherwise it skips
// uploading to Insights.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

const insightsEventName = "NRDiagIntegrationTestRun"

// IntegrationTestRun is a struct instantiated for each individual integration test to collect data about that test run.
type IntegrationTestRun struct {
	Name              string
	Context           string
	Build             string
	Status            string
	OperatingSystem   string
	User              string
	CommitAuthor      string
	Error             string
	BatchStartTime    time.Time
	BatchEndTime      time.Time
	BatchDuration     float64
	StartTime         time.Time
	EndTime           time.Time
	Duration          float64
	DockerBuild       testTimer
	DockerRun         testTimer
	StatusDockerBuild string
	StatusDockerRun   string
}

// testTimer is a struct to collect timing data about integration test run sub-tasks: Docker build, Docker run, etc.
type testTimer struct {
	StartTime time.Time
	EndTime   time.Time
	Duration  float64
}

// StartTimer starts the timer for an integration test run sub-task. E.g. test.DockerBuild.StartTimer()
func (t *testTimer) StartTimer() {
	t.StartTime = time.Now()
}

// StopTimer tops the timer for an integration test run sub-task and records the sub-task duration.
func (t *testTimer) StopTimer() {
	t.EndTime = time.Now()
	t.Duration = t.EndTime.Sub(t.StartTime).Seconds()
}

// WrapUp is called when an individual test completes to collect its ending timestmap and calculate duration.
func (i IntegrationTestRun) WrapUp() IntegrationTestRun {
	i.EndTime = time.Now()
	i.Duration = i.EndTime.Sub(i.StartTime).Seconds()
	log.Info(i.Name, " Test took", i.Duration)
	return i
}

// RecordTestTimings is the goroutine responsible for collecting timings
func RecordTestTimings(ch chan IntegrationTestRun, wg *sync.WaitGroup) {
	//Set up slice to keep track of all results
	defer wg.Done()

	allResults := []IntegrationTestRun{}

	BatchStartTime := time.Now()
	log.Info("Started test batch at ", BatchStartTime)
	//Iterate over channel while tests are running
	for result := range ch {
		result.BatchStartTime = BatchStartTime
		allResults = append(allResults, result)
	}

	BatchEndTime := time.Now()
	BatchDuration := BatchEndTime.Sub(BatchStartTime).Seconds()

	//Add job end time and duration to results
	for i := range allResults {
		allResults[i].BatchEndTime = BatchEndTime
		allResults[i].BatchDuration = BatchDuration
	}

	resultsJSON := processTestTimings(allResults)

	insightsAPIKey := os.Getenv("NRDIAG_INSIGHTS_API_KEY")
	insightsAccountID := os.Getenv("INSIGHTS_ACCOUNT_ID")

	// Skip the Insights API call if necessary environment variables are not present.
	if insightsAccountID != "" && insightsAPIKey != "" {
		log.Info("POSTing", len(allResults), "results to Insights account:", insightsAccountID)

		err := insertCustomEvents(insightsAccountID, insightsAPIKey, resultsJSON)

		if err != nil {
			log.Info("Insights event insertion unsuccessful.")
			log.Info(err)
		}

		log.Info("Successfully POSTed events to Insights")
	} else {
		log.Info("Skipping posting data into Insights API because environment variables are not set.")
	}
}

// processTestTimings takes a slice of IntegrationTestRun structs (integration test results)
// and returns a JSON string ready for the Insights API call.
func processTestTimings(resultsToProcess []IntegrationTestRun) string {
	var marshalledResults []string
	for _, result := range resultsToProcess {
		formattedResult, err := json.Marshal(result)
		if err != nil {
			log.Info("Error marshalling JSON for Insights payload")
			continue
		}
		marshalledResults = append(marshalledResults, string(formattedResult))
	}
	return "[" + strings.Join(marshalledResults, ",") + "]"
}

// MarshalJSON is used for custom JSON marshaling for IntegrationTestRun to format it for the Insights custom
// locally sourced small batch cage free artisinal gluten free/free range event POST payload.
func (i IntegrationTestRun) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		EventType           string  `json:"eventType"`
		Name                string  `json:"name"`
		Context             string  `json:"context"`
		OperatingSystem     string  `json:"operatingSystem,omitempty"`
		Build               string  `json:"build"`
		Status              string  `json:"status"`
		User                string  `json:"user,omitempty"`
		CommitAuthor        string  `json:"commitAuthor,omitempty"`
		Error               string  `json:"error,omitempty"`
		BatchStart          int64   `json:"batchStart"`
		BatchEnd            int64   `json:"batchEnd"`
		BatchDuration       float64 `json:"batchDuration"`
		Timestamp           int64   `json:"timestamp"`
		Duration            float64 `json:"duration"`
		DockerBuildDuration float64 `json:"dockerBuildDuration"`
		DockerRunDuration   float64 `json:"dockerRunDuration"`
		DockerBuildStatus   string  `json:"dockerBuildStatus"`
		DockerRunStatus     string  `json:"dockerRunStatus"`
	}{
		EventType:           insightsEventName,
		Name:                i.Name,
		Context:             i.Context,
		OperatingSystem:     i.OperatingSystem,
		User:                i.User,
		CommitAuthor:        i.CommitAuthor,
		Build:               i.Build,
		Status:              i.Status,
		Error:               i.Error,
		Timestamp:           i.StartTime.Unix(),
		BatchStart:          i.BatchStartTime.Unix(),
		BatchEnd:            i.BatchEndTime.Unix(),
		BatchDuration:       i.BatchDuration * 1000,
		Duration:            i.Duration * 1000,
		DockerBuildDuration: i.DockerBuild.Duration * 1000,
		DockerRunDuration:   i.DockerRun.Duration * 1000,
		DockerBuildStatus:   i.StatusDockerBuild,
		DockerRunStatus:     i.StatusDockerRun,
	})
}

// insertCustomEvents makes API POST request to Insights for inserting custom events.
func insertCustomEvents(account, apiKey, payload string) error {
	headers := make(map[string]string)
	headers["X-Insert-Key"] = apiKey
	headers["content-type"] = "application/json"
	headers["cache-control"] = "no-cache"

	var insertURL string
	insertURL = os.Getenv("STAGING_INSIGHTS_URL") + account + "/events"

	wrapper := httpHelper.RequestWrapper{
		Method:  "POST",
		URL:     insertURL,
		Headers: headers,
		Payload: strings.NewReader(payload),
	}

	resp, err := httpHelper.MakeHTTPRequest(wrapper)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Received non 200 response code from insert API: %d", resp.StatusCode)
	}

	return nil

}
