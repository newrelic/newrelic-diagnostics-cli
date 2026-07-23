package agentcontrol

import (
	"errors"
	"fmt"
	"io"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// AgentControlStatusServer collects the agent-control local status server output.
type AgentControlStatusServer struct {
	httpGetter   requestFunc
	configReader func(string) ([]byte, error)
}

// RequestResult contains HTTP response data for a single request.
type RequestResult struct {
	URL        string
	Status     string
	StatusCode int
	Body       string
	Err        error
}

func (p AgentControlStatusServer) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("AgentControl/Agent/StatusServer")
}

func (p AgentControlStatusServer) Explain() string {
	return "Collects agent-control status server output"
}

func (p AgentControlStatusServer) Dependencies() []string {
	return []string{}
}

func (p AgentControlStatusServer) Execute(_ tasks.Options, _ map[string]tasks.Result) tasks.Result {
	statusURL := p.resolveStatusURL()

	result := makeRequest(statusURL, p.httpGetter)

	if result.Err != nil {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: fmt.Sprintf("Could not connect to agent-control status server at %s: %s", statusURL, result.Err.Error()),
			Payload: result,
		}
	}

	if result.StatusCode == 200 {
		stream := make(chan string)
		go tasks.StreamBlob(result.Body, stream)

		return tasks.Result{
			Status:  tasks.Success,
			Summary: fmt.Sprintf("agent-control status server response:\n%s", result.Body),
			Payload: result,
			FilesToCopy: []tasks.FileCopyEnvelope{{
				Path:       "agent-control-status.json",
				Stream:     stream,
				Identifier: "AgentControl/Agent/StatusServer",
			}},
		}
	}

	return tasks.Result{
		Status:  tasks.Warning,
		Summary: fmt.Sprintf("Unexpected response from agent-control status server at %s: HTTP %d", statusURL, result.StatusCode),
		Payload: result,
	}
}

// resolveStatusURL reads host and port from the config file, falling back to defaults.
func (p AgentControlStatusServer) resolveStatusURL() string {
	host := defaultStatusHost
	port := defaultStatusPort

	cfg, err := readACConfig(p.configReader)
	if err != nil {
		log.Debug("Could not read agent-control config for status server, using defaults: " + err.Error())
	} else {
		if !cfg.Server.Enabled {
			log.Debug("agent-control status server is disabled in config")
		}
		if cfg.Server.Host != "" {
			host = cfg.Server.Host
		}
		if cfg.Server.Port != 0 {
			port = cfg.Server.Port
		}
	}

	return fmt.Sprintf("http://%s:%d/status", host, port)
}

func makeRequest(url string, HTTPagent requestFunc) RequestResult {
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
			err = errors.New("read error: " + err.Error())
		}
		_ = response.Body.Close()
	}

	return RequestResult{
		URL:        url,
		Status:     status,
		StatusCode: statusCode,
		Body:       string(body),
		Err:        err,
	}
}
