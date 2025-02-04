package agent

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	// ./nrdiag -browser-url http://thecustomers-website-url --suites browser
}

// Dependencies - Returns the dependencies for ech task.
func (t BrowserAgentGetSource) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BrowserAgentGetSource) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	log.Debug(options)
	url := options.Options["url"]

	if url == "" {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "This health check requires the usage of the command option '-browser-url'. Please re-run " + tasks.ThisProgramFullName + " using the flag in this manner: ./nrdiag -browser-url http://YOUR-WEBSITE-URL -suites browser",
		}
	}

	request := httpHelper.RequestWrapper{
		Method: "GET",
		URL:    url,
	}
	resp, err := httpHelper.MakeHTTPRequest(request)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Failed to connect to %s. Please make sure to add a protocol to the URL or verify connectivity. Encountered error: %s", url, err.Error()),
		}
	}

	stream := make(chan string)
	responseBody, errReadingBody := io.ReadAll(resp.Body)
	go streamSource(responseBody, stream)
	if errReadingBody != nil {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("Failed to connect to %s. Please check your URL and verify connectivity: %s", url, errReadingBody.Error()),
		}
	}
	resultFilesToCopy := []tasks.FileCopyEnvelope{{Path: "nrdiag-output/source.html",
		Stream:     stream,
		Identifier: t.Identifier().String(),
	}}
	//goodScripts means that we 1 or more loader scripts inside the head tag and badScripts means that we found instance of loader scripts outside of the head tag
	goodScripts, badScripts := getLoaderScript(string(responseBody))

	loaderFoundInheadTag := len(goodScripts) > 0
	loaderFoundOutsideHead := len(badScripts) > 0

	if !loaderFoundInheadTag && !loaderFoundOutsideHead {
		return tasks.Result{
			Status:      tasks.Failure,
			Summary:     fmt.Sprintf("We were unable to find the Browser agent script in the HTML source code for %s. If this has not been added to the application yet, follow one of the deployment options suggested in our documentation. However, if the browser script is being loaded via an external file, keep in mind that this can cause issues in collecting the data for New Relic.", url),
			URL:         "https://docs.newrelic.com/docs/browser/browser-monitoring/installation/install-browser-monitoring-agent",
			FilesToCopy: resultFilesToCopy,
		}
	}

	if loaderFoundInheadTag && loaderFoundOutsideHead {
		return tasks.Result{
			Status:      tasks.Warning,
			Summary:     "We found at least one instance of the New Relic Browser script element outside of the </head> tag. This script should only be present in the </head> tag or it can cause monitoring issues.",
			URL:         "https://docs.newrelic.com/docs/browser/browser-monitoring/installation/install-browser-monitoring-agent",
			FilesToCopy: resultFilesToCopy,
			Payload: BrowserAgentSourcePayload{
				URL:    url,
				Source: string(responseBody),
				Loader: goodScripts,
			},
		}
	}

	if loaderFoundOutsideHead {
		return tasks.Result{
			Status:      tasks.Failure,
			Summary:     "We found the browser agent in the source code but it is not inline in the <head>. This can cause issues in collecting the data. The best practices for installation are to copy the snippet of code and paste it as close to the top of the HEAD as possible, but after any position-sensitive META tags (X-UA-Compatible and charset). The script needs to load very early in order to wrap the browser's built-in APIs",
			URL:         "https://docs.newrelic.com/docs/browser/browser-monitoring/installation/install-browser-monitoring-agent",
			FilesToCopy: resultFilesToCopy,
		}
	}
	//Looks like we only have good scripts
	return tasks.Result{
		Status:      tasks.Success,
		Summary:     "We successfully found the New Relic Browser script element in the following page's source: " + url + ". The body of the page has been included in the nrdiag-zip file",
		FilesToCopy: resultFilesToCopy,
		Payload: BrowserAgentSourcePayload{
			URL:    url,
			Source: string(responseBody),
			Loader: goodScripts,
		},
	}
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

func getLoaderScript(data string) ([]string, []string) {
	var goodScripts, badScripts []string
	nrScriptRgx := regexp.MustCompile("window.NREUM")
	// check for Browser agent in head tag
	scriptInHeadRgx := regexp.MustCompile(`(?m:<script\b[^>]*>([\s\S]*?)<\/script>[\s\S]*<\/head>)`)
	scriptInHeadMatches := scriptInHeadRgx.FindAllString(data, -1)

	if len(scriptInHeadMatches) > 0 {
		for _, match := range scriptInHeadMatches {
			if nrScriptRgx.MatchString(match) {
				// it is considered a good script because the loader was found inside the Head tag
				goodScripts = append(goodScripts, match)
			}
		}
	}
	//This would match any script: (?m:<script\b[^>]*>([\s\S]*?)<\/script>)
	scriptAfterHeadRgx := regexp.MustCompile(`(<\/head>[\s\S]*<script\b[^>]*>([\s\S]*?)<\/script>)`)
	scriptAfterheadMatches := scriptAfterHeadRgx.FindAllString(data, -1)
	// Now loop through scriptAfterheadMatches to find the browser agent loader outside of head tag
	for _, match := range scriptAfterheadMatches {
		if nrScriptRgx.MatchString(match) {
			badScripts = append(badScripts, match)
		}
	}

	return goodScripts, badScripts

}
