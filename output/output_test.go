package output

import (
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func Test_GetResultsJSON(t *testing.T) {
	OutputNow = func() time.Time {
		return time.Date(2000, 12, 15, 17, 8, 00, 0, time.UTC)
	}

	fakeResults := generateResultArray()

	expected := readFile("fixtures/test-output.json")
	if runtime.GOOS == "windows" {
		expected = readFile("fixtures/test-output_windows.json")
	}
	observed := getResultsJSON(fakeResults)

	//if you intended to make changes to the output JSON:
	// - uncomment the next line of code for one run
	// - inspect new-output.json to make sure it looks like what you expect
	// - replace test-output.json with new-output.json
	// - comment the line and run the test again
	//ioutil.WriteFile("fixtures/new-output.json", []byte(observed), 0644)

	if expected != observed {
		t.Error("Expected:", expected, "Observed:", observed)
	}
}
func Test_StreamDataOutput(t *testing.T) {

	dataChannel := make(chan string)

	fakeResults := []registration.TaskResult{
		{
			Task: registration.TasksForIdentifierString("Base/Log/Collect")[0],
			Result: tasks.Result{
				Status:  tasks.Success,
				Summary: "Streamed data",
				URL:     "",
				FilesToCopy: []tasks.FileCopyEnvelope{
					{Path: "data.txt", Stream: dataChannel, Identifier: "Base/Log/Collect"},
				},
			},
		},
	}

	go streamData(dataChannel)

	expected := readFile("fixtures/test-stream-output.json")
	if runtime.GOOS == "windows" {
		expected = readFile("fixtures/test-stream-output_windows.json")
	}
	observed := getResultsJSON(fakeResults)

	//if you intended to make changes to the output JSON:
	// - uncomment the next line of code for one run
	// - inspect new-output.json to make sure it looks like what you expect
	// - replace test-stream-output.json with new-stream-output.json
	// - comment the line and run the test again
	//ioutil.WriteFile("fixtures/new-stream-output.json", []byte(observed), 0644)

	if expected != observed {
		t.Error("Expected:", expected, "Observed:", observed)
	}
}

func streamData(ch chan string) {
	ch <- "line 1\n"
	ch <- "line 2\n"
	ch <- "line 3\n"
	ch <- "line 4\n"
	ch <- "line 5\n"

	close(ch)
}

func Test_WriteLineResults(t *testing.T) {
	tests := []struct {
		name string
		want []tasks.Result
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WriteLineResults(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WriteLineResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func generateResultArray() []registration.TaskResult {
	resA := registration.TaskResult{
		Task: registration.TasksForIdentifierString("Base/Config/Collect")[0],
		Result: tasks.Result{
			Status:  tasks.Success,
			Summary: "4 config files(s) found",
			URL:     "",
			FilesToCopy: []tasks.FileCopyEnvelope{
				{Path: "./fixtures/java/newrelic/newrelic.yml", Identifier: "Base/Config/Collect"},
				{Path: "./fixtures/ruby/config/newrelic.yml", Identifier: "Base/Config/Collect"},
			},
			Payload: "[{\"FileName\":\"newrelic.yml\",\"FilePath\":\"/Users/btribbia/dev/go/src/github.com/newrelic/newrelic-diagnostics-cli/fixtures/java/newrelic/\"},{\"FileName\":\"newrelic.yml\",\"FilePath\":\"/Users/btribbia/dev/go/src/github.com/newrelic/newrelic-diagnostics-cli/fixtures/ruby/config/\"}]",
		},
	}

	resB := registration.TaskResult{
		Task: registration.TasksForIdentifierString("Base/Config/Validate")[0],
		Result: tasks.Result{
			Status:      tasks.Success,
			Summary:     "",
			URL:         "",
			FilesToCopy: nil,
			Payload:     nil,
		},
	}

	resC := registration.TaskResult{
		Task: registration.TasksForIdentifierString("Base/Collector/ConnectUS")[0],
		Result: tasks.Result{
			Status:      tasks.Success,
			Summary:     "200 OK",
			URL:         "",
			FilesToCopy: nil,
			Payload:     nil,
		},
	}
	results := []registration.TaskResult{resA, resB, resC}

	return results
}

func readFile(file string) string {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Info("error reading file", err)
	}
	//This is to fix line ending in Windows
	replaced := strings.Replace(string(content), "\r\n", "\n", -1)
	return replaced
}
