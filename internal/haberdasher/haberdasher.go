package haberdasher

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	defaultBaseURL   = "http://localhost:3000"
	defaultUserAgent = "haberdasher-go/1.0"
	contentType      = "application/json"
)

//DefaultClient - Singleton instance of Haberdasher API client
var DefaultClient = &Client{}

// Client is the primary data structure that an implementer would interface
// with.  It contains all the details for host to contact the API, and which
// service endpoints are implemented.
type Client struct {
	httpClient *http.Client

	// Base URL for API requests. Defaults to localhost
	// can be set to different endpoint to use for local development, etc.
	BaseURL *url.URL

	// User Agent used when communication with Haberdasher API
	UserAgent string

	// RunID is unique to each run of the Diagnostics CLI and is required as a header for many haberdasher endpoints
	RunID string

	// InsertKey is a value seeded by the RunID, and is required as a header for many haberdasher endpoints
	InsertKey string

	common service // Reuse a single struct instead of allocating one for each service on the heap.

	// Services used for talking to different parts of the Haberdasher API
	// Usage       *UsageService
	// Attachments *AttachmentsService
	Tasks *TasksService
}

type service struct {
	client *Client
}

// Response is a Haberdasher API response. This wraps the http.Response
// returned from Haberdasher and provides convenient access to things
type Response struct {
	*http.Response
	Body []byte
}

// InitializeDefaultClient initializes the single public client (DefaultClient) with default options.
func InitializeDefaultClient() {
	DefaultClient = newClientWithDefaults()
}

// newClientWithDefaults will create a new haberdasher client using default values
func newClientWithDefaults() *Client {
	baseURL, _ := url.Parse(defaultBaseURL)

	newClient := &Client{
		httpClient: http.DefaultClient,
		BaseURL:    baseURL,
		UserAgent:  defaultUserAgent,
	}

	// Create client after figuring out defaults
	newClient.common.client = newClient

	// Assign services to client
	newClient.Tasks = (*TasksService)(&newClient.common)

	return newClient
}

// NewRequest creates an API request
func (c *Client) NewRequest(method, url string, body interface{}) (*http.Request, error) {
	// Parse BaseURL with path
	u, err := c.BaseURL.Parse(url)
	if err != nil {
		return nil, err
	}

	// Encode body of request to JSON
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	// Set up HTTP request
	req, err := http.NewRequest(method, u.String(), buf)
	if err != nil {
		return nil, err
	}

	// Set headers required by Haberdasher service
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Run-Id", c.RunID)
	req.Header.Set("Insert-Key", c.InsertKey)

	// Set Content-Type for body
	if body != nil {
		req.Header.Set("Content-Type", contentType)
	}

	return req, nil
}

// Do sends an API request and returns the API response. The response is JSON
// decoded.
func (c *Client) Do(req *http.Request, respStruct interface{}) (*Response, error) {

	// Make HTTP request to API
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close() // nolint: errcheck

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %v", err)
	}

	if resp.StatusCode >= 300 {
		bodyString := string(bodyBytes)
		return nil, fmt.Errorf("Expected StatusCode < 300 got %d: %v", resp.StatusCode, bodyString)
	}

	response := newResponse(resp)
	response.Body = bodyBytes

	// Decode HTTP response JSON
	if respStruct != nil {
		decErr := json.Unmarshal(bodyBytes, &respStruct)

		if decErr == io.EOF {
			decErr = nil
		}

		if decErr != nil {
			err = fmt.Errorf("Error unmarshalling to %T: %v", respStruct, decErr)
		}

	}

	return response, err
}

// newResponse creates a new Response for the provided http.Response
// r must not be nil
func newResponse(r *http.Response) *Response {
	response := &Response{Response: r}
	return response
}

// SetRunID sets the Diagnostics CLI run ID for a client
func (c *Client) SetRunID(runID string) {
	c.RunID = runID
	c.InsertKey = generateInsertKey(runID)
}

// SetBaseURL sets the Haberdasher base URL for a client
func (c *Client) SetBaseURL(baseURL string) {
	parsed, _ := url.Parse(baseURL)
	c.BaseURL = parsed
}

// SetUserAgent sets the User-Agent value for a client
func (c *Client) SetUserAgent(userAgent string) {
	c.UserAgent = userAgent
}

// generateInsertKey takes a string, and returns a unique deterministic hash
// it generates a SHA512 digest based on every other char of the input, reversed
func generateInsertKey(runID string) string {
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
