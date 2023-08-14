package tasks

import (
	"encoding/json"
	"fmt"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

// Result stores the response from a task execution
type Result struct {
	Status      Status
	Summary     string             // customer facing verbiage
	URL         string             // a URL pointing to documentation about this task; "required" on Warning or Failure, desireable on any status
	FilesToCopy []FileCopyEnvelope // List of files identified by the task to be included in zip file
	Payload     interface{}        // task defined list of returned data. This is what is used by downstream tasks so data format agreements are between tasks
}

// Status statusEnum listing of valid values for status
type Status int

//Constants for use by the status property above
const (
	//None - this task does not apply to this system and has no meaningful data to report.
	None Status = iota
	//Success - A task has completed and identified the system conforms to our expectations
	Success
	//Warning - the system does not conform to our expectations, but may not actually be in a problematic state
	Warning
	//Failure - the system does not conform to expectations and we believe the system is in a broken state
	Failure
	//Error - an internal error has occurred and we were unable to determine the state of this check (eg: permissions error)
	Error
	//Info - A task has completed, but it has only collected information, no "judgments" here
	Info
)

// Equals verifies two Result objects match each other. It purposefully does not verify payloads match exact since ordering may be non-deterministic but all other values are compared.
func (r Result) Equals(result Result) bool {
	if r.Status != result.Status {
		log.Info("Status didn't match")
		return false
	}

	if r.Summary != result.Summary {
		log.Info("Summary didn't match")
		return false
	}

	if r.URL != result.URL {
		log.Info("URL didn't match")
		return false
	}

	return true
}

func (r Result) IsFailure() bool {
	return r.Status != None && r.Status != Success && r.Status != Info
}

// HasPayload will check if a upstream task.Result has a payload we can work with. Notice status 'Warning' is not included here and it's because a lot of the time it has payload. But HasPayload may not be applicable to some tasks.
func (r Result) HasPayload() bool {
	return r.Status != None && r.Status != Error && r.Status != Failure
}

//MarshalJSON - custom JSON marshaling for this task, in this case we ignore the parsed config
func (s Status) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.StatusToString())
}

// Task describes the interface all agent tasks implement
type Task interface {
	Identifier() Identifier
	Explain() string
	Dependencies() []string
	Execute(Options, map[string]Result) Result
}

//ByIdentifier is a sort helper to sort an array of tasks by their identifiers
type ByIdentifier []Task

func (t ByIdentifier) Len() int {
	return len(t)
}
func (t ByIdentifier) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}
func (t ByIdentifier) Less(i, j int) bool {
	t1 := t[i].Identifier()
	t2 := t[j].Identifier()

	if t1.Category == t2.Category && t1.Subcategory == t2.Subcategory {
		return t1.Name < t2.Name
	} else if t1.Category == t2.Category {
		return t1.Subcategory < t2.Subcategory
	} else {
		return t1.Category < t2.Category
	}
}

// Options passes in the options to execute an individual task, likely picked up from command line options to override default values
type Options struct { // what goes in here? maybe this also should be string to pass in arbitrary json?
	Options map[string]string // Map of core and task specific options
}

// Identifier contains the task's name, category and subcategory
type Identifier struct {
	Category    string
	Subcategory string
	Name        string
}

func (i Identifier) String() string {
	return fmt.Sprintf("%s/%s/%s", i.Category, i.Subcategory, i.Name)
}

// IdentifierFromString - converts a sting like "Category/SubCategory/Name" to an Identifier
func IdentifierFromString(s string) Identifier {
	parts := strings.Split(s, "/")
	if len(parts) == 1 {
		return Identifier{}
	}
	return Identifier{Category: parts[0], Subcategory: parts[1], Name: parts[2]}
}
