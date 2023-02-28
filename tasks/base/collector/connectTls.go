package collector

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"golang.org/x/net/html"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type BaseCollectorTLS struct {
	upstream   map[string]tasks.Result
	httpGetter requestFunc
}

func (p BaseCollectorTLS) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Collector/ConnectTLS")
}

func (p BaseCollectorTLS) Explain() string {
	return "Check network connection to New Relic US region collector endpoint"
}

func (p BaseCollectorTLS) Dependencies() []string {
	return []string{
		"Base/Config/ProxyDetect",
		"Base/Config/RegionDetect",
	}
}

func (p BaseCollectorTLS) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	p.upstream = upstream

	url := "https://connection-test.newrelic.com/"

	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            url,
		TimeoutSeconds: 30,
	}
	resp, err := p.httpGetter(wrapper)
	if err != nil {
		return p.prepareCollectorErrorResult(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return p.prepareResponseErrorResult(err, strconv.Itoa(resp.StatusCode))
	}

	return p.prepareResult(string(body), strconv.Itoa(resp.StatusCode))

}

func (p BaseCollectorTLS) prepareCollectorErrorResult(e error) tasks.Result {
	var result tasks.Result
	if e == nil {
		return result
	}
	result.Status = tasks.Failure
	result.Summary = "There was an error connecting to connection-test.newrelic.com."
	result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
	result.Summary += "\nError = " + e.Error()
	result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks"

	return result
}

func (p BaseCollectorTLS) prepareResponseErrorResult(e error, statusCode string) tasks.Result {
	var result tasks.Result
	if e == nil {
		return result
	}
	result.Status = tasks.Warning
	result.Summary = "Status = " + statusCode + ". There was an issue reading the body when connecting to connection-test.newrelic.com."
	result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
	result.Summary += "Error = " + e.Error()
	result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks"

	return result
}

func (p BaseCollectorTLS) prepareResult(body, statusCode string) tasks.Result {
	var result tasks.Result

	if statusCode != "200" {
		log.Debug("Non-200 response received from connection-test.newrelic.com:", statusCode)
		log.Debug("Body:", body)
		result.Status = tasks.Warning
		result.Summary = "connection-test.newrelic.com returned a non-200 STATUS CODE: " + statusCode
		result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
		result.Summary += "\nResponse Body: " + body
		result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks"
		return result
	}

	tlsVer, err := p.parseTlsStringFromHtml(body)
	if err != nil {
		log.Debug("Unable to parse TLS version from connection-test.newrelic.com output.")
		log.Debug("Body:", body)
		result.Status = tasks.Warning
		result.Summary = "Unable to parse TLS version from connection-test.newrelic.com output."
		result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
		result.Summary += "\nResponse Body: " + body
		result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks"
		return result
	}

	tlsMeetsReqs := p.checkTlsVerMeetsReqs(tlsVer)
	if !tlsMeetsReqs {
		log.Debug("TLS version didn't meet requirements or failed to parse.")
		log.Debug("Body:", body)
		result.Status = tasks.Failure
		result.Summary = "connection-test.newrelic.com detected an unsupported TLS Version."
		result.Summary += "\nPlease check network and proxy settings and try again or see -help for more options."
		result.Summary += "\nResponse Body: " + body
		result.URL = "https://docs.newrelic.com/docs/new-relic-solutions/get-started/networks"
		return result
	}

	log.Debug("Successfully connected")
	result.Status = tasks.Success
	result.Summary = "TLS Version: " + tlsVer
	return result
}

func (p BaseCollectorTLS) checkTlsVerMeetsReqs(ver string) bool {
	splitVer := strings.Split(ver, "TLSv")
	if len(splitVer) == 2 {
		verFl, err := strconv.ParseFloat(splitVer[1], 64)
		if err != nil {
			return false
		}
		if verFl == 1.2 || verFl == 1.3 {
			return true
		}
	}

	return false
}

func (p BaseCollectorTLS) parseTlsStringFromHtml(body string) (string, error) {
	var isTlsSection bool
	tkn := html.NewTokenizer(strings.NewReader(body))
	for {
		tt := tkn.Next()

		switch {
		case tt == html.ErrorToken:
			return "", errors.New("tls version not found")
		case tt == html.TextToken:
			t := strings.TrimSpace(tkn.Token().String())
			if isTlsSection {
				return t, nil
			}
			if t == "TLS Version" {
				isTlsSection = true
			}
		}
	}
}
