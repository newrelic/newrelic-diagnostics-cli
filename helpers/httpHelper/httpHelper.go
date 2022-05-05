package httpHelper

import (
	"errors"
	"io"
	"net/http"
	"time"
	"net"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	pb "gopkg.in/cheggaaa/pb.v1"
)

//RequestWrapper is a basic wrapper to streamline the use of http requests within the project
type RequestWrapper struct {
	Method         string
	URL            string
	Headers        map[string]string
	Payload        io.Reader
	Length         int64
	TimeoutSeconds int16
	BypassProxy bool
}

//NewHTTPRequestWrapper - returns a new request wrapper for creating an http request
func NewHTTPRequestWrapper() RequestWrapper {
	var wrapper RequestWrapper
	wrapper.Method = "GET"
	return wrapper
}

//Default request timeout of 30 seconds if no timeout value is passed to helper.
const defaultTimeoutSeconds = 30

//MakeHTTPRequest -  takes the basics of a request and makes it
func MakeHTTPRequest(wrapper RequestWrapper) (*http.Response, error) {

	if wrapper.URL == "" || wrapper.Method == "" {
		log.Info("Error: URL or method are not set")
		return nil, errors.New("error: URL or method are not set")
	}
	reader := wrapper.Payload
	// set up a progress bar if length is set
	if wrapper.Length != 0 {
		bar := pb.New(int(wrapper.Length)).SetUnits(pb.U_BYTES)
		bar.Start()
		defer bar.Finish()
		reader = bar.NewProxyReader(wrapper.Payload)
	}

	//Now create our request object
	req, _ := http.NewRequest(wrapper.Method, wrapper.URL, reader)

	// Setting the content length header if supplied
	if wrapper.Length != 0 {
		req.ContentLength = wrapper.Length
	}

	// Now add headers if they exist
	for k, v := range wrapper.Headers {
		req.Header.Set(k, v)
	}

	//If no request timeout is provided, set timeout to default value.
	if wrapper.TimeoutSeconds == 0 {
		wrapper.TimeoutSeconds = defaultTimeoutSeconds
	}

	var transport http.RoundTripper
	
	//All HTTP requests use the same http transport. 
	//However, when bypassing the detected system proxy, the request will be using its own unique http transport.
	if wrapper.BypassProxy {
		//These are the http.DefaultTransport values minus the Proxy
		transport = &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
	} else {
		transport = http.DefaultTransport
	}	

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Duration(wrapper.TimeoutSeconds) * time.Second,
	}

	req.Header.Set("User-Agent", "Nrdiag_/"+config.Version)

	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil

}
