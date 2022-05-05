package usage

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	l "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/example/template"
)

var sampleFileCopyEnvelope = []tasks.FileCopyEnvelope{}

var sampleTaskResult = registration.TaskResult{
	Task: template.ExampleTemplateMinimalTask{},
	Result: tasks.Result{
		Status:      tasks.Success,
		Summary:     "Summary string",
		URL:         "http://www.example.com",
		FilesToCopy: sampleFileCopyEnvelope,
		Payload:     template.ExamplePayload{}},
	WasOverride: false,
}

var sampleSingleTaskResultSlice = []registration.TaskResult{sampleTaskResult}
var sampleMultipleTaskResultSlice = []registration.TaskResult{
	sampleTaskResult,
	sampleTaskResult,
	sampleTaskResult,
}

var samplePreparedSingleTaskResult = []taskResult{
	{Identifier: "Example/Template/MinimalTask", Status: "Success", URL: "http://www.example.com"},
}

var emptyTaskResultSlice = []registration.TaskResult{}

var samplePreparedPayload = payload{
	Protocol: "1.0",
	Data: runData{
		MetaData:      sampleMetaData,
		Configuration: samplePreparedConfig,
		Results:       samplePreparedSingleTaskResult,
	},
}

var samplePreparedConfig = []config.ConfigFlag{
	{Name: "verbose", Value: true},
	{Name: "interactive", Value: true},
	{Name: "quiet", Value: true},
	{Name: "veryQuiet", Value: false},
	{Name: "help", Value: false},
	{Name: "version", Value: true},
	{Name: "yesToAll", Value: false},
	{Name: "showOverrideHelp", Value: true},
	{Name: "proxy", Value: true},
	{Name: "proxyUser", Value: true},
	{Name: "proxyPassword", Value: false},
	{Name: "tasks", Value: "string"},
	{Name: "configFile", Value: true},
	{Name: "override", Value: false},
	{Name: "outputPath", Value: false},
	{Name: "filter", Value: "string"},
	{Name: "fileUpload", Value: true},
	{Name: "browserURL", Value: true},
	{Name: "attachmentEndpoint", Value: true},
}

var apiStubReflectBody = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	body, _ := ioutil.ReadAll(r.Body)
	w.Write(body)
}))

func getJSONfixture(filename string, t *testing.T) (string, error) {
	t.Helper()
	jsonFile := filepath.Join("test-fixtures", filename)
	jsonData, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func Test_usageAPI_postData(t *testing.T) {

	jsonString, err := getJSONfixture("expectedSurveyResponse.json", t)

	var responseData usageResponse
	if marshalErr := json.Unmarshal([]byte(jsonString), &responseData); marshalErr != nil {
		t.Errorf("Was unable to parse JSON sample response: %s", jsonString)
	}

	sampleHeaders := make(map[string]string)
	sampleHeaders["Content-Type"] = "application/json"
	sampleHeaders["Usage-Protocol"] = "v1"
	sampleHeaders["Run-Id"] = "f82e068d-7cac-44f0-924e-8a1f5539c146"
	sampleHeaders["Insert-Key"] = "6b7b3cf7ae75869fa87fc77ebb8bf5b11612023e2148d8aa2aec4e8025cd7fd8"

	if err != nil {
		t.Errorf("Error retrieving JSON fixture: %v", err)
	}

	type fields struct {
		serviceEndpoint serviceEndpoint
	}
	type args struct {
		data    string
		headers map[string]string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    usageResponse
		wantErr bool
	}{
		{
			name:   "it should post JSON to an external service as POST request body",
			fields: fields{serviceEndpoint: serviceEndpoint{URL: apiStubReflectBody.URL}},
			args:   args{data: jsonString, headers: sampleHeaders},
			want:   responseData,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &usageAPI{
				serviceEndpoint: tt.fields.serviceEndpoint,
			}
			got, err := u.postData(tt.args.data, tt.args.headers)
			if (err != nil) != tt.wantErr {
				t.Errorf("usageAPI.postData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("usageAPI.postData() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_prepareResults(t *testing.T) {
	type args struct {
		data []registration.TaskResult
	}
	tests := []struct {
		name string
		args args
		want []taskResult
	}{
		{
			name: "should parse single task result into expected data structure",
			args: args{data: sampleSingleTaskResultSlice},
			want: []taskResult{
				{Identifier: "Example/Template/MinimalTask", Status: "Success", URL: "http://www.example.com"},
			},
		},
		{
			name: "should parse multiple task results into expected data structure",
			args: args{data: sampleMultipleTaskResultSlice},
			want: []taskResult{
				{Identifier: "Example/Template/MinimalTask", Status: "Success", URL: "http://www.example.com"},
				{Identifier: "Example/Template/MinimalTask", Status: "Success", URL: "http://www.example.com"},
				{Identifier: "Example/Template/MinimalTask", Status: "Success", URL: "http://www.example.com"},
			},
		},
		{
			name: "should parse empty task result into expected data structure",
			args: args{data: emptyTaskResultSlice},
			want: []taskResult{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := prepareResults(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("prepareResults() = %v, want %v", got, tt.want)
			}
		})
	}
}

var sampleMetaData = metaData{
	Timestamp:     1538089691,
	NRDiagVersion: "1.0",
}

func Test_preparePayload(t *testing.T) {
	type args struct {
		protocolVersion string
		results         []taskResult
		configuration   []config.ConfigFlag
		metadata        metaData
	}
	tests := []struct {
		name string
		args args
		want payload
	}{
		{
			name: "It should prepare the expected payload",
			args: args{
				protocolVersion: "1.0",
				results:         samplePreparedSingleTaskResult,
				configuration:   samplePreparedConfig,
				metadata:        sampleMetaData,
			},
			want: samplePreparedPayload,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := preparePayload(tt.args.protocolVersion, tt.args.results, tt.args.configuration, tt.args.metadata); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("preparePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_genInsertKey(t *testing.T) {
	type args struct {
		runID string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should generate expected SHA512",
			args: args{runID: "f82e068d-7cac-44f0-924e-8a1f5539c146"},
			want: "582f577622720778e2513cd6346c34b686ba7a5471bbebaad9d0a69b9c211b8cfd9c75f63c2417e0ba637fb82bb38e11f1a83e6576a58b12d2c8c098cdde8cad",
		},
		{
			name: "it should generate another expected SHA512",
			args: args{runID: "1e4b2da4-e475-4390-94d5-948ec517bd32"},
			want: "8c14eb59d1160a63c1054557c1535f193b0efe80d61058e2fa3f932dd08cd36d3429a4576c85bba2d761745f1368079d02f96917b5782794ca6c95db5ae5e6ff",
		},
		{
			name: "it should handle empty input",
			args: args{runID: ""},
			want: "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := genInsertKey(tt.args.runID); got != tt.want {
				t.Errorf("genInsertKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getRPMdetails(t *testing.T) {

	type args struct {
		r tasks.Result
	}
	tests := []struct {
		name string
		args args
		want []rpmApp
	}{
		{name: "it should parse account and app ID from reporting to log line",
			args: args{r: tasks.Result{
				Status:  tasks.Info,
				Summary: "I found a log with reporting to lines",
				URL:     "",
				Payload: []l.LogNameReportingTo{{
					Logfile:     "new_relic.log",
          ReportingTo: []string{"https://rpm.newrelic.com/accounts/111/applications/21487336"},
				}},
			}},
			want: []rpmApp{{
				AppID:     "21487336",
        AccountID: "111",
			}},
		},
		{name: "it should not parse account and app ID from garbage reporting to log line",
			args: args{r: tasks.Result{
				Status:  tasks.Info,
				Summary: "I found a log with reporting to lines",
				URL:     "",
				Payload: []l.LogNameReportingTo{{
					Logfile:     "new_relic.log",
          ReportingTo: []string{"https://rpm.newrelic.com/accounts/111applications21487336"},
				}},
			}},
			want: []rpmApp{},
		},
		{name: "it should not parse account and app ID from empty array",
			args: args{r: tasks.Result{
				Status:  tasks.Info,
				Summary: "I did not find a log with reporting to lines",
				URL:     "",
				Payload: []l.LogNameReportingTo{{}},
			}},
			want: []rpmApp{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRPMdetails(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getReportingToDetails() = %v, want %v", got, tt.want)
			}
		})
	}
}
