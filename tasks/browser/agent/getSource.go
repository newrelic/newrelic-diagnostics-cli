package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"regexp"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// BrowserAgentGetSource - This struct defined the sample plugin which can be used as a starting point
type BrowserAgentGetSource struct {
}

type BrowserAgentSourcePayload struct {
	Source string
	URL    string
	Loader []string
}

func (payload BrowserAgentSourcePayload) MarshalJSON() ([]byte, error) {
	//note: this technique can be used to return anything you want, including modified values or nothing at all.
	//anything that gets returned here ends up in the output json file
	return json.Marshal(&struct {
		URL string
	}{
		URL: payload.URL,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BrowserAgentGetSource) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Browser/Agent/GetSource")
}

// Explain - Returns the help text for each individual task
func (t BrowserAgentGetSource) Explain() string {
	return "Determine New Relic Browser agent loader script from provided URL"
	// gorun -t browser/agent/detect -v -o Browser/Agent/Detect.url=http://localhost:3000
}

// Dependencies - Returns the dependencies for ech task.
func (t BrowserAgentGetSource) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BrowserAgentGetSource) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	log.Debug(options)
	url := options.Options["url"]

	request := httpHelper.RequestWrapper{
		Method: "GET",
		URL:    url,
	}
	resp, err := httpHelper.MakeHTTPRequest(request)
	if err != nil {
		log.Debug("Error was", err)
		result.Status = tasks.Failure
		result.Summary = "Failed to connect to " + url + " Please check your URL and verify connectivity"
		return result
	}
	stream := make(chan string)
	responseBody, er := ioutil.ReadAll(resp.Body)
	go streamSource(responseBody, stream)
	if er != nil {
		log.Debug("Error reading body was", er)
		result.Status = tasks.Failure
		result.Summary = "Failed to connect to " + url + " Please check your URL and verify connectivity"
		return result
	}
	result.FilesToCopy = []tasks.FileCopyEnvelope{tasks.FileCopyEnvelope{Path: "nrdiag-output/source.html",
		Stream:     stream,
		Identifier: t.Identifier().String(),
	}}

	result.Payload = BrowserAgentSourcePayload{
		URL:    url,
		Source: string(responseBody),
		Loader: getLoader(string(responseBody)),
	}
	result.Status = tasks.Success
	result.Summary = "Successfully downloaded source from " + url

	return result
}

func streamSource(responseBody []byte, ch chan string) {
	//defer reader.Close()
	defer close(ch)

	scanner := bufio.NewScanner(bytes.NewReader(responseBody))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}

}

func getLoader(data string) []string {
	var scripts []string
	// First we grab the script tags
	regex := regexp.MustCompile(`(?m:<script\b[^>]*>([\s\S]*?)<\/script>)`)

	matches := regex.FindAllString(data, -1)
	// Now loop through matches to find the browser agent loader
	nreum := regexp.MustCompile("window.NREUM")
	for _, script := range matches {
		if nreum.MatchString(script) {
			log.Debug("Browser Loader found")
			scripts = append(scripts, script)
		}
	}

	if len(scripts) == 0 {
		log.Debug("No loader found")
	}

	return scripts
}
