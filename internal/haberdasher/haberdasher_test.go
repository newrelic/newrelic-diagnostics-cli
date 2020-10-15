package haberdasher

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

var (
	// testMux is the HTTP request multiplexer used with the test server.
	testMux *mux.Router

	// testClient is the Haberdasher client being tested.
	testClient *Client

	// testServer is a test HTTP server used to provide mock API responses.
	testServer *httptest.Server
)

// setup sets up a test HTTP server along with a haberdasher.Client that is configured to talk to that test server.
// Tests should register handlers on mux which provide mock responses for the API method being tested.
func setup() {
	// Test server
	testMux = mux.NewRouter()
	testServer = httptest.NewServer(testMux)

	// haberdasher client configured to use test server
	testClient = newClientWithDefaults()
	testClient.SetBaseURL(testServer.URL)

}

func teardown() {
	testServer.Close()
}

func TestBadResponseCode(t *testing.T) {
	setup()
	defer teardown()
	testAPIEndpoint := "/tasks/license-key"

	//Define mock response
	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		testMethod(t, r, "POST")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, "Not found")
	})

	//Create request to mock server
	req, err := testClient.NewRequest("POST", testAPIEndpoint, nil)
	if err != nil {
		t.Fatalf("Failed to setup http.NewRequest: %v", err)
	}

	//Perform request to mock server
	_, err = testClient.Do(req, nil)
	if err == nil {
		t.Error("Expected an error, but we didn't get one")
	}

	testErrorMsg(t, err, "Expected StatusCode < 300 got 404: Not found")

}

// Test that run ID and insert key are included on every request
func TestExpectedHeaders(t *testing.T) {
	setup()
	defer teardown()

	testAPIEndpoint := "/tasks/license-key"
	testRunID := "2ed53550-cc19-11e8-bcf3-f45c8992bc33"
	testUserAgent := "Nrdiag_/test"
	testInsertKey := "c9cffd0a7ccfaadca6882e3f4f7bc8f863d0f4c11e97d26636ac00b3aec9d1aab4ebffe71d71e0cdb13a582916009b8df8b6bc506f8c7dd118f6253f6fd6bf21"

	testClient.SetRunID(testRunID)
	testClient.SetUserAgent(testUserAgent)

	//Define mock response
	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		testHeader(t, r, "Run-Id", testRunID)
		testHeader(t, r, "Insert-Key", testInsertKey)
		testHeader(t, r, "User-Agent", testUserAgent)

		fmt.Fprint(w, nil)
	})

	//Create request to mock server
	req, err := testClient.NewRequest("POST", testAPIEndpoint, nil)
	if err != nil {
		t.Fatalf("Failed to setup http.NewRequest: %v", err)
	}

	//Perform request to mock server
	_, err = testClient.Do(req, nil)
	if err != nil {
		t.Errorf("Did not expect error, but received: %s", err.Error())
	}

}

// Test that client is setup with defaults
func TestClientDefaults(t *testing.T) {

	clientWithDefaults := newClientWithDefaults()

	expectedBaseURL, _ := url.Parse("http://localhost:3000")
	expectedUserAgent := "haberdasher-go/1.0"

	if !reflect.DeepEqual(clientWithDefaults.BaseURL, expectedBaseURL) {
		t.Errorf("Default client base URL: %s, want %s", clientWithDefaults.BaseURL, expectedBaseURL)
	}

	if !strings.EqualFold(clientWithDefaults.UserAgent, expectedUserAgent) {
		t.Errorf("Default client user agent: %s, want %s", clientWithDefaults.UserAgent, expectedUserAgent)
	}

}

//Insert key generation
func Test_generateInsertKey(t *testing.T) {
	tests := []struct {
		runID string
		want  string
	}{
		{runID: "2ed53550-cc19-11e8-bcf3-f45c8992bc33",
			want: "c9cffd0a7ccfaadca6882e3f4f7bc8f863d0f4c11e97d26636ac00b3aec9d1aab4ebffe71d71e0cdb13a582916009b8df8b6bc506f8c7dd118f6253f6fd6bf21"},
		{runID: "303a05c5-cc19-11e8-9c2b-f45c8992bc33",
			want: "b031fdd9010ec21bd0e15d3f0ec93d458cbd7d39dc5822621cd478744977b186e854c5b720aa6b95acc9ced851516af1daf8531f02c6c2d07396bfcdefe79332"},
		{runID: "31502c2a-cc19-11e8-8edf-f45c8992bc33",
			want: "760b740efd3aa0cfc8e2f55872cb8f86dba5603e1d57a988778eeae8d3566e505470929766fe4dc5ef1044bd05730486618f4cb309f6f933471f993ba8cda2dd"},
		{runID: "325a303f-cc19-11e8-9f24-f45c8992bc33",
			want: "5e0c83e49f35bed6ca092f71426e20cdd656ceec60c816e588d714c22b484530113dc5d741bd83d2d2cb5ff0bfe80e1f9def86abd05aa3df96a3f7b2d4a93999"},
		{runID: "335c0279-cc19-11e8-bff6-f45c8992bc33",
			want: "9934cd99e4fd82b6acd75398e7efb2f558417bdc5060afac77956885eda115001c095da9d5c7d0c3594bb7179c73675cdd0deac693e2cc735f6202064afa9a6c"},
		{runID: "347647e3-cc19-11e8-bffe-f45c8992bc33",
			want: "b9f30f02f258e47eb356a9389a29ed55a554f0648e82c0f4f9df75f7519e908b16d613a01c2da1bce1691528bffc004f2ae34ab35dcf2c9ce5b0d53e2a9e8886"},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test #%v", i), func(t *testing.T) {
			if got := generateInsertKey(tt.runID); got != tt.want {
				t.Errorf("generateInsertKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ====== Test helpers

func testHeader(t *testing.T, r *http.Request, header string, want string) {
	t.Helper()

	if got := r.Header.Get(header); got != want {
		t.Errorf("Header.Get(%q) returned %q, want %q", header, got, want)
	}
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testErrorMsg(t *testing.T, e error, want string) {
	t.Helper()
	if got := e.Error(); got != want {
		t.Errorf("Request method: %v, want %v", got, want)
	}
}

func testRequestURL(t *testing.T, r *http.Request, want string) {
	t.Helper()
	vars := mux.Vars(r)
	for k, v := range vars {
		want = strings.Replace(want, fmt.Sprintf("{%s}", k), v, -1)
	}

	if got := r.URL.String(); !strings.HasPrefix(got, want) {
		t.Errorf("Request URL: %v, want %v", got, want)
	}
}
