package usage

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"regexp"
	"strings"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	l "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
)

const defaultProtocolVersion = "1.0"
const defaultUsageEndpoint = "http://localhost:3000/usage"

type usageResponse struct {
	Survey struct {
		Prompt string
		URI    string
	}
}

type payload struct {
	Protocol string  `json:"protocol"`
	Data     runData `json:"data"`
}

type runData struct {
	MetaData      metaData            `json:"metadata"`
	Configuration []config.ConfigFlag `json:"configuration"`
	Results       []taskResult        `json:"results"`
}

type metaData struct {
	Timestamp     int64    `json:"timestamp"`
	NRDiagVersion string   `json:"nrdiagVersion"`
	RunID         string   `json:"runId"`
	RpmApps       []rpmApp `json:"rpmApps"`
	Hostname      string   `json:"hostname"`
	LicenseKeys   []string `json:"licenseKeys"`
}

type taskResult struct {
	Identifier string `json:"identifier"`
	Status     string `json:"status"`
	URL        string `json:"URL"`
}

type rpmApp struct {
	AppID     string `json:"appId"`
	AccountID string `json:"accountId"`
}

type httpResponse struct {
	statusCode int
	body       string
}

type usageAPI struct {
	serviceEndpoint
}

type serviceEndpoint struct {
	URL string
}

func getUnixTime() int64 {
	return time.Now().Unix()
}

func getProtocolVersion() string {
	return defaultProtocolVersion
}

// genInsertKey takes a string, and returns a unique deterministic hash
// it generates a SHA512 digest based on every other char of the input, reversed
func genInsertKey(runID string) string {
	runIDBytes := []byte(runID) // convert input to slice of bytes
	hashSeed := []byte{}        // initialize accumulator hash

	for i, b := range runIDBytes {
		if i%2 == 0 {
			/* if index is even, push this char in on top of our slice,
			   this implementation has the effect of reversing it */
			hashSeed = append([]byte{b}, hashSeed...)
		}
	}
	h := sha512.New()
	h.Write(hashSeed)       // generate digest
	hashBytes := h.Sum(nil) // output digest
	// convert to human readable lowercase string before returning
	return hex.EncodeToString(hashBytes[:])
}

// SendUsageData generates and POSTS NRdiag usage data to the Haberdasher service
func SendUsageData(results []registration.TaskResult, runID string) {
	usageEndpoint := usageAPI{
		serviceEndpoint{
			URL: config.UsageEndpoint,
		},
	}
	log.Debug("Sending usage data to", config.UsageEndpoint)
	preparedResults := prepareResults(results)
	preparedConfig := config.Flags.UsagePayload()
	preparedMeta := prepareMeta(results, runID)

	protocolVersion := getProtocolVersion()
	preparedPayload := preparePayload(protocolVersion, preparedResults, preparedConfig, preparedMeta)

	payloadJSON, err := json.Marshal(preparedPayload)
	if err != nil {
		log.Debug("Error marshalling payload into JSON", err)
	}

	payloadString := string(payloadJSON)
	preparedHeaders := genRequestHeaders(preparedMeta)
	response, err := usageEndpoint.postData(payloadString, preparedHeaders)
	if err != nil {
		log.Debug("Error sending usage data:", err)
		return
	}

	link := response.Survey.URI
	if link == "" {
		return
	}

	prompt := response.Survey.Prompt
	if prompt == "" {
		prompt = "We'd love to know more about your experience using the " + tasks.ThisProgramFullName + "! Please visit:"
	}

	log.Info(prompt)
	log.Info(link)
	log.Info()
}

func preparePayload(protocolVersion string, results []taskResult, configuration []config.ConfigFlag, metadata metaData) payload {
	packaged := runData{
		MetaData:      metadata,
		Configuration: configuration,
		Results:       results,
	}

	return payload{
		Protocol: protocolVersion,
		Data:     packaged,
	}
}

// prepareResults gathers and converts registration.TaskResults to individual maps
// Returns a slice of maps, each map represents one task
func prepareResults(data []registration.TaskResult) []taskResult {
	prepared := []taskResult{}
	for _, result := range data {
		prepared = append(prepared, taskResult{
			Identifier: result.Task.Identifier().String(),
			Status:     result.Result.StatusToString(),
			URL:        result.Result.URL,
		})
	}

	return prepared
}

func getRPMdetails(r tasks.Result) []rpmApp {
	result := []rpmApp{}

	rpmAppslice := r.Payload.([]l.LogNameReportingTo)

	regex := `https://rpm.newrelic.com/accounts/(?P<account>\d+)/applications/(?P<application>\d+)$`
	regexKey := regexp.MustCompile(regex)

	for _, logfile := range rpmAppslice {
		for _, line := range logfile.ReportingTo {
			matches := regexKey.FindStringSubmatch(line)
			if len(matches) > 2 {
				result = append(result, rpmApp{
					AccountID: matches[1],
					AppID:     matches[2],
				})
			}
		}
	}

	return result
}

func getHostname(h env.HostInfo) string {
	return h.Hostname
}

// gather and package metadata - currently a passthrough
func prepareMeta(results []registration.TaskResult, runID string) metaData {
	var runTimeMetaData = metaData{
		Timestamp:     getUnixTime(),
		NRDiagVersion: config.Version,
		RunID:         runID,
		RpmApps:       []rpmApp{},
		Hostname:      "",
		LicenseKeys:   []string{},
	}

	for _, result := range results {
		if result.Task.Identifier().String() == "Base/Config/ValidateLicenseKey" && result.Result.Status == tasks.Success {
			licenseKeyToSources, ok := result.Result.Payload.(map[string][]string)
			//We do not need the value of sources(if lk is env var or comes from config file, etc) for this operation
			reducedLicenseKeys := []string{}

			for lk := range licenseKeyToSources {
				reducedLicenseKeys = append(reducedLicenseKeys, lk)
			}

			if ok {
				runTimeMetaData.LicenseKeys = reducedLicenseKeys
			} else {
				log.Info("Unable to send licenseKeys metadata for Haberdasher because of a Type Assertion error")
			}
		}
		if result.Task.Identifier().String() == "Base/Log/ReportingTo" && result.Result.Status == tasks.Success {
			runTimeMetaData.RpmApps = getRPMdetails(result.Result)
		}
		if result.Task.Identifier().String() == "Base/Env/HostInfo" && result.Result.Status == tasks.Info && result.Result.Payload != nil {
			runTimeMetaData.Hostname = getHostname(result.Result.Payload.(env.HostInfo))
		}
	}
	return runTimeMetaData
}

func genRequestHeaders(m metaData) map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = "application/json"
	headers["Usage-Protocol"] = getProtocolVersion()
	headers["Run-Id"] = m.RunID
	headers["Insert-Key"] = genInsertKey(m.RunID)

	return headers
}

// postData initiates an HTTP POST request to the provided *usageAPI.URL endpoint with content-type JSON header
// and the passed string as the request body.
func (u *usageAPI) postData(data string, headers map[string]string) (usageResponse, error) {
	var response httpResponse
	if len(u.URL) == 0 {
		// u.URL is assigned by build script. If not present, fall back to default
		u.URL = defaultUsageEndpoint
	}
	wrapper := httpHelper.RequestWrapper{
		Method:         "POST",
		URL:            u.URL,
		Payload:        strings.NewReader(data),
		Headers:        headers,
		TimeoutSeconds: 15,
	}

	res, err := httpHelper.MakeHTTPRequest(wrapper)

	if err != nil {
		log.Debug("Failed upload: " + err.Error())
		return usageResponse{}, err
	}

	bodyBytes, _ := io.ReadAll(res.Body)
	bodyString := string(bodyBytes)

	response.statusCode = res.StatusCode
	response.body = bodyString

	// Check the response
	if response.statusCode != 200 {
		log.Debug("Unexpected status code from usage endpoint:", response.statusCode)
		return usageResponse{}, errors.New("unexpected status code")
	}

	var responseData usageResponse
	if marshalErr := json.Unmarshal(bodyBytes, &responseData); marshalErr != nil {
		log.Debugf("Error parsing json response when attempting to upload ticket attachments: %s\n", marshalErr.Error())
		log.Debug(response.body)
		return usageResponse{}, marshalErr
	}

	return responseData, nil
}
