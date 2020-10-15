package haberdasher

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
)

func TestTasksService_ValidateLicenseKeys(t *testing.T) {
	setup()
	defer teardown()
	testAPIEndpoint := "/tasks/license-key"

testLicenseKeys := []string{"08a2ad66c637a29c3982469a3fe8d1982d002c4a", "bad_key"}
	expectedParsedResults := []LicenseKeyResult{
		{
			LicenseKey: "08a2ad66c637a29c3982469a3fe8d1982d002c4a",			
      IsValid:    true,
		},
		{
			LicenseKey: "bad_key",
			IsValid:    false,
		},
	}

	rawResponse, err := ioutil.ReadFile("./mocks/license_key_response_success.json")
	if err != nil {
		t.Error(err.Error())
	}

	testMux.HandleFunc(testAPIEndpoint, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		testMethod(t, r, "POST")
		testRequestURL(t, r, testAPIEndpoint)
		fmt.Fprint(w, string(rawResponse))
	})

	results, _, err := testClient.Tasks.ValidateLicenseKeys(testLicenseKeys)
	if err != nil {
		t.Errorf("Error received: %v", err)
	}

	if !reflect.DeepEqual(results, expectedParsedResults) {
		t.Errorf("Default client base URL: %#v, want %#v", results, expectedParsedResults)
	}

}
