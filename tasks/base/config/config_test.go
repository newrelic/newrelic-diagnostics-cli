package config

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func Test_getHSMConfiguration(t *testing.T) {
	type args struct {
		configElement ValidateElement
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{
			name: "No found keys with empty configElement",
			args: args{
				configElement: ValidateElement{},
			},
			want: false,
		},
		{
			name: "No found keys with parseBool error",
			args: args{
				configElement: ValidateElement{
					ParsedResult: tasks.ValidateBlob{
						Key:      "high_security",
						RawValue: "nottrue",
					},
					Config: ConfigElement{
						FileName: "newrelic.yml",
						FilePath: "/etc/",
					},
				},
			},
			want: false,
		},
		{
			name: "Found key true",
			args: args{
				configElement: ValidateElement{
					ParsedResult: tasks.ValidateBlob{
						Key:      "high_security",
						RawValue: "true",
					},
					Config: ConfigElement{
						FileName: "newrelic.yml",
						FilePath: "/etc/",
					},
				},
			},
			want: true,
		},
		{
			name: "Found key false",
			args: args{
				configElement: ValidateElement{
					ParsedResult: tasks.ValidateBlob{
						Key:      "high_security",
						RawValue: "false",
					},
					Config: ConfigElement{
						FileName: "newrelic.yml",
						FilePath: "/etc/",
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getHSMConfiguration(tt.args.configElement); got != tt.want {
				t.Errorf("getHSMConfiguration() = %v, want %v", got, tt.want)
			}
		})
	}
}
