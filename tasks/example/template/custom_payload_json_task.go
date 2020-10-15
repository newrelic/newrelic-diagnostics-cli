package template

import (
	"encoding/json"

	"github.com/newrelic/NrDiag/tasks"
)

// ExampleTemplateCustomPayloadJSONTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateCustomPayloadJSONTask struct {
}

// FilteredWiFiAuthPayload - stores information about various WiFi access points which might get used by other tasks
type FilteredWiFiAuthPayload struct {
	Company    string
	Location   string
	SSID       string
	Passphrase string //this seems like something we shouldn't output, but we still need... see below
}

//MarshalJSON - custom JSON marshaling for this task, we'll strip out the passphrase to keep it only in memory, not on disk
func (payload FilteredWiFiAuthPayload) MarshalJSON() ([]byte, error) {
	//note: this technique can be used to return anything you want, including modified values or nothing at all.
	//anything that gets returned here ends up in the output json file
	return json.Marshal(&struct {
		Company  string
		Location string
		SSID     string
	}{
		Company:  payload.Company,
		Location: payload.Location,
		SSID:     payload.SSID,
	})
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateCustomPayloadJSONTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/CustomPayloadJSONTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateCustomPayloadJSONTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateCustomPayloadJSONTask) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p ExampleTemplateCustomPayloadJSONTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	result.Status = tasks.Success
	result.Payload = FilteredWiFiAuthPayload{
		Company:    "New Relic, Inc.",
		Location:   "PDX",
		SSID:       "NR-GUEST",
		Passphrase: "RubyOnRails!",
	}

	return result
}
