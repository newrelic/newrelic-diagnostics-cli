package config

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func Test_checkValidation(t *testing.T) {
	type args struct {
		validations []config.ValidateElement
	}
	tests := []struct {
		name string
		args args
		want bool
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := checkValidation(tt.args.validations); got != tt.want {
				t.Errorf("checkValidation() = %v, want %v", got, tt.want)
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
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, got := checkConfig(tt.args.configs); got != tt.want {
				t.Errorf("checkConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}
