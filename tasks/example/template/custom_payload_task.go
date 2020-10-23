package template

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// ExampleTemplateCustomPayloadTask - This struct defined the sample plugin which can be used as a starting point
type ExampleTemplateCustomPayloadTask struct {
}

// WiFiAuthPayload - stores information about various WiFi access points which might get used by other tasks
type WiFiAuthPayload struct {
	Company    string
	Location   string
	SSID       string
	Passphrase string //this seems like something we shouldn't output, but we still need... see custom_payload_json_task.go
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p ExampleTemplateCustomPayloadTask) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Example/Template/CustomPayloadTask")
}

// Explain - Returns the help text for each individual task
func (p ExampleTemplateCustomPayloadTask) Explain() string {
	return "This task doesn't do anything."
}

// Dependencies - Returns the dependencies for each task.
func (p ExampleTemplateCustomPayloadTask) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (p ExampleTemplateCustomPayloadTask) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.None,
		Summary: "I succeeded in doing nothing.",
	}

	result.Status = tasks.Success
	result.Payload = WiFiAuthPayload{
		Company:    "New Relic, Inc.",
		Location:   "PDX",
		SSID:       "NR-GUEST",
		Passphrase: "RubyOnRails!",
	}

	return result
}
