package config

import (
	"testing"

	"os"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

func Test_checkValidation(t *testing.T) {
	type args struct {
		validations []config.ValidateElement
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"Java Agent from valid YML", args{validations: validateElementFromFile("fixtures/agent_testdata1.yml")}, true},
		{"Invalid YML file", args{validations: validateElementFromFile("fixtures/agent_testdata2.yml")}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, gotBool := checkValidation(tt.args.validations)
			if gotBool != tt.want {
				log.Info("Config found was", config)
				t.Errorf("checkValidation() = %v, want %v", gotBool, tt.want)
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
		{name: "Java Agent from invalid YML", args: args{configs: []config.ConfigElement{{FileName: "agent_testdata2.yml", FilePath: "fixtures/"}}}, want: true},
		{name: "Java Agent config not found", args: args{configs: []config.ConfigElement{{FileName: "agent_testdata3", FilePath: "fixtures/"}}}, want: false},
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

func Test_checkForJar(t *testing.T) {

	tests := []struct {
		name       string
		createFile bool
		want       bool
	}{
		{name: "jar present", createFile: true, want: true},
		{name: "jar missing", createFile: false, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createFile(tt.createFile)
			log.Debug("running", tt.name)
			if got := checkForJar(); got != tt.want {
				t.Errorf("checkForJar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func validateElementFromFile(file string) []config.ValidateElement {
	var payload []config.ValidateElement

	fileHandle, _ := os.Open(file)
	parsedConfig, _ := config.ParseYaml(fileHandle)
	payload = append(payload, config.ValidateElement{Config: config.ConfigElement{FileName: "test.yml", FilePath: ""}, Status: tasks.Success, ParsedResult: parsedConfig})
	return payload
}
func createFile(create bool) {
	if create {
		log.Debug("create File")
		file, err := os.Create("newrelic.jar") // For read access.
		defer file.Close()
		if err != nil {
			log.Info("Error creating file", err)
		}
	} else {
		log.Debug("delete file")
		err := os.Remove("newrelic.jar")
		if err != nil {
			log.Info("Error deleting file", err)
		}

	}

}
