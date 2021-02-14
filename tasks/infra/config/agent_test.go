package config

import (
	"os"
	"reflect"
	"testing"

	baseConfig "github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func TestMain(m *testing.M) {
	//Toggle to enable verbose logging
	baseConfig.LogLevel = baseConfig.Info
	os.Exit(m.Run())
}

var validationFromDefaultYML []config.ValidateElement

func Test_checkValidation(t *testing.T) {
	type args struct {
		validations []config.ValidateElement
	}

	validationFromDefaultYML = validateElementFromFile("fixtures/default_infra_config/newrelic-infra.yml")
	validationFromInvalidYML := validateElementFromFile("fixtures/invalid_infra_config/newrelic-infra.yml")
	validationFromMinimalYML := validateElementFromFile("fixtures/minimal_infra_config/newrelic-infra.yml")
	file, _ := os.Open("fixtures/java_config/newrelic.yml")
	parsedFromJavaConfigYML, _ := config.ParseYaml(file)
	validationFromJavaConfigYML := []config.ValidateElement{config.ValidateElement{
		Config: config.ConfigElement{FileName: "newrelic.yml", FilePath: ""}, Status: tasks.Success, ParsedResult: parsedFromJavaConfigYML}}

	type want struct {
		wantValidation []config.ValidateElement
		wantBool       bool
	}

	tests := []struct {
		name string
		args args
		want
	}{
		{"Infra Agent from valid YML", args{validations: validationFromDefaultYML}, want{validationFromDefaultYML, true}},
		{"Invalid YML file", args{validations: validationFromInvalidYML}, want{validationFromInvalidYML, false}},
		{"Infra Agent minimal config file", args{validations: validationFromMinimalYML}, want{validationFromInvalidYML, true}},
		{"Java Agent config file should not pass", args{validations: validationFromJavaConfigYML}, want{[]config.ValidateElement{}, false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log.Debug("running", tt.name)
			gotValid, gotBool := checkValidation(tt.args.validations)
			if gotBool != tt.want.wantBool {
				if gotValid[0].ParsedResult.String() != tt.want.wantValidation[0].ParsedResult.String() {
					t.Errorf("checkValidation() validateElement = %v, want %v", gotValid, tt.want)
					t.Errorf("checkValidation() bool = %v, want %v", gotBool, tt.want)

				}

			}
		})
	}
}

func Test_checkConfig(t *testing.T) {
	type args struct {
		configs []config.ConfigElement
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{name: "Infra Agent from invalid YML", args: args{configs: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/invalid_infra_config/"}}}, want: true},
		{name: "Infra Agent config not found", args: args{configs: []config.ConfigElement{{FileName: "newrelic-infra", FilePath: "fixtures/empty_yml/"}}}, want: false},
		{name: "Infra Agent minimal yml", args: args{configs: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/minimal_infra_config/"}}}, want: true},
		{name: "Java Agent should not pass", args: args{configs: []config.ConfigElement{{FileName: "newrelic.yml", FilePath: "fixtures/java_config/"}}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			_, gotBool := checkConfig(tt.args.configs)
			if gotBool != tt.want {
				t.Errorf("checkConfig() = %v, want %v", gotBool, tt.want)
			}
		})
	}
}

func Test_checkForBinary(t *testing.T) {
	tests := []struct {
		name       string
		createFile bool
		want       bool
	}{
		{name: "binary present", createFile: true, want: true},
		{name: "binary missing", createFile: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createFile(t, tt.createFile)
			log.Debug("running", tt.name)
			got, _ := checkForBinary()
			if got != tt.want {
				t.Errorf("checkForBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validateElementFromFile(file string) []config.ValidateElement {
	var payload []config.ValidateElement
	fileHandle, _ := os.Open(file)
	defer fileHandle.Close()
	parsedConfig, _ := config.ParseYaml(fileHandle)
	payload = append(payload, config.ValidateElement{Config: config.ConfigElement{FileName: "newrelic-infra.yml", FilePath: ""}, Status: tasks.Success, ParsedResult: parsedConfig, Error: ""})
	return payload
}

func createFile(t *testing.T, create bool) {
	t.Helper()
	if create {
		log.Debug("create File")
		file, err := os.Create("newrelic-infra") // For read access.
		defer file.Close()
		if err != nil {
			log.Debug("Error creating file", err)
		}
	} else {
		log.Debug("delete file")
		err := os.Remove("newrelic-infra")
		if err != nil {
			log.Debug("Error deleting file", err)
		}

	}

}

func checkPayload(payload interface{}, expected interface{}) bool {
	return true
}

func TestInfraConfigAgent_Execute(t *testing.T) {
	successResultConfig := tasks.Result{Status: tasks.Success,
		Summary: "Infra agent identified as present on system from validated config file",
	}
	successResultParsed := tasks.Result{Status: tasks.Success,
		Summary: "Infra agent identified as present on system from parsed config file",
	}
	successResultBinary := tasks.Result{Status: tasks.Success,
		Summary: "Infra agent identified as present on system from existence of binary file: newrelic-infra.exe",
	}
	emptyResult := tasks.Result{Status: tasks.None, Summary: tasks.NoAgentDetectedSummary, URL: "", FilesToCopy: nil, Payload: nil}
	mockValidationTrue := func([]config.ValidateElement) ([]config.ValidateElement, bool) {
		return []config.ValidateElement{}, true
	}

	mockValidationFalse := func([]config.ValidateElement) ([]config.ValidateElement, bool) {
		return []config.ValidateElement{}, false
	}
	mockConfigTrue := func([]config.ConfigElement) ([]config.ConfigElement, bool) {
		return []config.ConfigElement{}, true
	}

	mockConfigFalse := func([]config.ConfigElement) ([]config.ConfigElement, bool) {
		return []config.ConfigElement{}, false
	}
	mockBinaryTrue := func() (bool, string) {
		return true, "newrelic-infra.exe"
	}

	mockBinaryFalse := func() (bool, string) {
		return false, ""
	}

	type fields struct {
		validationChecker validationFunc
		configChecker     configFunc
		binaryChecker     binaryFunc
	}
	type args struct {
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   tasks.Result
	}{
		{name: "It should return successful result from validation",
			fields: fields{validationChecker: mockValidationTrue, configChecker: mockConfigTrue, binaryChecker: mockBinaryTrue},
			args: args{upstream: map[string]tasks.Result{
				"Base/Config/Collect": {
					Status:  tasks.Success,
					Payload: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/default_infra_config/"}},
				},
				"Base/Config/Validate": {
					Status:  tasks.Success,
					Payload: validationFromDefaultYML,
				},
			}},
			want: successResultConfig},
		{name: "It should return successful result from parsed config",
			fields: fields{validationChecker: mockValidationFalse, configChecker: mockConfigTrue, binaryChecker: mockBinaryTrue},
			args: args{upstream: map[string]tasks.Result{
				"Base/Config/Collect": {
					Status:  tasks.Success,
					Payload: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/default_infra_config/"}},
				},
				"Base/Config/Validate": {
					Status: tasks.Failure,
				},
			}},
			want: successResultParsed},
		{name: "It should return successful result from binary file",
			fields: fields{validationChecker: mockValidationFalse, configChecker: mockConfigFalse, binaryChecker: mockBinaryTrue},
			args: args{upstream: map[string]tasks.Result{
				"Base/Config/Collect": {
					Status:  tasks.Success,
					Payload: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/default_infra_config/"}},
				},
				"Base/Config/Validate": {
					Status: tasks.Failure,
				},
			}},
			want: successResultBinary},
		{name: "It should return empty result from binary file not found",
			fields: fields{validationChecker: mockValidationFalse, configChecker: mockConfigFalse, binaryChecker: mockBinaryFalse},
			args: args{upstream: map[string]tasks.Result{
				"Base/Config/Collect": {
					Status:  tasks.Success,
					Payload: []config.ConfigElement{{FileName: "newrelic-infra.yml", FilePath: "fixtures/default_infra_config/"}},
				},
				"Base/Config/Validate": {
					Status: tasks.Failure,
				},
			}},
			want: emptyResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigAgent{
				validationChecker: tt.fields.validationChecker,
				configChecker:     tt.fields.configChecker,
				binaryChecker:     tt.fields.binaryChecker,
			}

			gotResult := p.Execute(tasks.Options{}, tt.args.upstream)

			if !gotResult.Equals(tt.want) {
				t.Errorf("checkForInfraAgent() didn't match: \n%v\n%v", gotResult, tt.want)
				t.Logf("Got Result is: %v", gotResult)
				t.Logf("tt.wantResult is: %v", tt.want)

			}

		})
	}
}

func TestInfraConfigAgent_Dependencies(t *testing.T) {
	type fields struct {
		name           string
		upstream       map[string]tasks.Result
		validationFunc validationFunc
		configFunc     configFunc
		binaryFunc     binaryFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{name: "It returns expected dependencies", fields: fields{}, want: []string{"Base/Config/Collect", "Base/Config/Validate"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigAgent{
				name:              tt.fields.name,
				upstream:          tt.fields.upstream,
				validationChecker: tt.fields.validationFunc,
				configChecker:     tt.fields.configFunc,
				binaryChecker:     tt.fields.binaryFunc,
			}
			if got := p.Dependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfraConfigAgent.Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfraConfigAgent_Explain(t *testing.T) {
	type fields struct {
		name              string
		upstream          map[string]tasks.Result
		validationChecker validationFunc
		configChecker     configFunc
		binaryChecker     binaryFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "It returns the expected explain", fields: fields{}, want: "Detect New Relic Infrastructure agent"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigAgent{
				name:              tt.fields.name,
				upstream:          tt.fields.upstream,
				validationChecker: tt.fields.validationChecker,
				configChecker:     tt.fields.configChecker,
				binaryChecker:     tt.fields.binaryChecker,
			}
			if got := p.Explain(); got != tt.want {
				t.Errorf("InfraConfigAgent.Explain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInfraConfigAgent_Identifier(t *testing.T) {
	type fields struct {
		name              string
		upstream          map[string]tasks.Result
		validationChecker validationFunc
		configChecker     configFunc
		binaryChecker     binaryFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   tasks.Identifier
	}{
		{name: "It returns the expected Identifier", fields: fields{}, want: tasks.IdentifierFromString("Infra/Config/Agent")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := InfraConfigAgent{
				name:              tt.fields.name,
				upstream:          tt.fields.upstream,
				validationChecker: tt.fields.validationChecker,
				configChecker:     tt.fields.configChecker,
				binaryChecker:     tt.fields.binaryChecker,
			}
			if got := p.Identifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InfraConfigAgent.Identifier() = %v, want %v", got, tt.want)
			}
		})
	}
}
