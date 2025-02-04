package config

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Base/Config/ValidateHSM", func() {
	var p BaseConfigValidateHSM

	Describe("Identifier()", func() {
		It("Should return correct identifier", func() {
			expectedIdentifier := tasks.Identifier{
				Category:    "Base",
				Subcategory: "Config",
				Name:        "ValidateHSM",
			}

			Expect(p.Identifier()).To(Equal(expectedIdentifier))
		})

	})

	Describe("Explain()", func() {
		It("Should return correct task explanations string", func() {
			expectedExplanation := "Validate High Security Mode agent configuration against account configuration"
			Expect(p.Explain()).To(Equal(expectedExplanation))
		})
	})

	Describe("Dependencies()", func() {
		It("Should return an expected slice of dependencies", func() {
			expectedDependencies := []string{"Base/Config/Validate", "Base/Env/CollectEnvVars"}
			Expect(p.Dependencies()).To(Equal(expectedDependencies))
		})
	})

})

func TestBaseConfigValidateHSM_Execute(t *testing.T) {
	type fields struct {
		configElements []ValidateElement
		envVars        map[string]string
	}
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   tasks.Result
		retVal map[string]bool
	}{
		// TODO: Add test cases.
		{
			name: "envVars Could not be checked for validation",
			fields: fields{
				configElements: []ValidateElement{
					{
						ParsedResult: tasks.ValidateBlob{
							Key:      "high_security",
							RawValue: "true",
						},
						Config: ConfigElement{
							FileName: "newrelic.yml",
							FilePath: "/etc/",
						},
					},
					{
						ParsedResult: tasks.ValidateBlob{
							Key:      "high_security",
							RawValue: "true",
						},
						Config: ConfigElement{
							FileName: "newrelic.js",
							FilePath: "/app/",
						},
					},
				},
				envVars: map[string]string{},
			},
			args: args{
				options: tasks.Options{},
				upstream: map[string]tasks.Result{
					"Base/Config/Validate": {
						Status: tasks.Success,
						Payload: []ValidateElement{
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
							{
								ParsedResult: tasks.ValidateBlob{
									Key:      "high_security",
									RawValue: "true",
								},
								Config: ConfigElement{
									FileName: "newrelic.js",
									FilePath: "/app/",
								},
							},
						},
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: "Local High Security Mode setting (true) for configuration:\n\nnewrelic.yml\n\n" + "Local High Security Mode setting (false) for configuration:\n\nnewrelic.js\n\n",
				Payload: map[string]bool{
					"newrelic.yml": true,
					"newrelic.js":  false,
				},
			},
			retVal: map[string]bool{
				"newrelic.yml": true,
				"newrelic.js":  false,
			},
		},
		{
			name: "No config elements or env vars to validate",
			fields: fields{
				configElements: []ValidateElement{},
				envVars:        map[string]string{},
			},
			args: args{
				options: tasks.Options{},
				upstream: map[string]tasks.Result{
					"Base/Config/Validate": {
						Status:  tasks.Success,
						Payload: []ValidateElement{},
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "No New Relic configuration files or environment variables to check high security mode against. Task did not run.\n",
			},
		},
		{
			name: "No config elements to validate but env vars exist",
			fields: fields{
				configElements: []ValidateElement{},
				envVars: map[string]string{
					"NEW_RELIC_HIGH_SECURITY": "false",
				},
			},
			args: args{
				options: tasks.Options{},
				upstream: map[string]tasks.Result{
					"Base/Config/Validate": {
						Status:  tasks.Success,
						Payload: []ValidateElement{},
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: "Local High Security Mode setting (false) for configuration:\n\nNEW_RELIC_HIGH_SECURITY\n\n",
				Payload: map[string]bool{
					"NEW_RELIC_HIGH_SECURITY": true,
				},
			},
			retVal: map[string]bool{"NEW_RELIC_HIGH_SECURITY": true},
		},
		{
			name: "No config elements to validate but env vars exist but none contain hsm",
			fields: fields{
				configElements: []ValidateElement{},
				envVars: map[string]string{
					"NOTSECURITYMODE": "false",
				},
			},
			args: args{
				options: tasks.Options{},
				upstream: map[string]tasks.Result{
					"Base/Config/Validate": {
						Status:  tasks.Success,
						Payload: []ValidateElement{},
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "No configurations for high security mode found.\n",
			},
			retVal: map[string]bool{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createHSMLocalValTest := func([]ValidateElement, BaseConfigValidateHSM) map[string]bool {
				return tt.retVal
			}
			tr := BaseConfigValidateHSM{
				createHSMLocalValidation: createHSMLocalValTest,
				envVars:                  tt.fields.envVars,
			}
			got := tr.Execute(tt.args.options, tt.args.upstream)
			if !reflect.DeepEqual(got.Status, tt.want.Status) {
				t.Errorf("BaseConfigValidateHSM.Execute().Status = %v, want %v", got.Status, tt.want.Status)
			}
			if !reflect.DeepEqual(got.Payload, tt.want.Payload) {
				t.Errorf("BaseConfigValidateHSM.Execute().Payload = %v, want %v", got.Payload, tt.want.Payload)
			}
		})
	}
}

func TestBaseConfigValidateHSM_GetHSMConfigurations(t *testing.T) {
	type fields struct {
		getHSMConfiguration GetHSMConfiguration
		envVars             map[string]string
	}
	type args struct {
		configElements []ValidateElement
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]bool
	}{
		// TODO: Add test cases.
		{
			name: "Return empty map",
			fields: fields{

				getHSMConfiguration: func(ve ValidateElement) bool { return false },
				envVars:             map[string]string{},
			},
			args: args{
				configElements: []ValidateElement{},
			},
			want: map[string]bool{},
		},
		{
			name: "Return env vars configuration",
			fields: fields{

				getHSMConfiguration: func(ve ValidateElement) bool { return false },
				envVars: map[string]string{
					"NEW_RELIC_HIGH_SECURITY": "true",
				},
			},
			args: args{
				configElements: []ValidateElement{},
			},
			want: map[string]bool{
				"NEW_RELIC_HIGH_SECURITY": true,
			},
		},
		{
			name: "Return config files configuration",
			fields: fields{
				getHSMConfiguration: func(ve ValidateElement) bool { return true },
				envVars:             map[string]string{},
			},
			args: args{
				configElements: []ValidateElement{
					{
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
			},
			want: map[string]bool{
				"/etc/newrelic.yml": true,
			},
		},
		{
			name: "Return config files and env vars configuration",
			fields: fields{
				getHSMConfiguration: func(ve ValidateElement) bool { return true },
				envVars: map[string]string{
					"NEW_RELIC_HIGH_SECURITY": "false",
				},
			},
			args: args{
				configElements: []ValidateElement{
					{
						ParsedResult: tasks.ValidateBlob{
							Key:      "high_security",
							RawValue: "true",
						},
						Config: ConfigElement{
							FileName: "newrelic.yml",
							FilePath: "/etc/",
						},
					},
					{
						ParsedResult: tasks.ValidateBlob{
							Key:      "high_security",
							RawValue: "true",
						},
						Config: ConfigElement{
							FileName: "newrelic.js",
							FilePath: "/app/",
						},
					},
				},
			},
			want: map[string]bool{
				"NEW_RELIC_HIGH_SECURITY": false,
				"/etc/newrelic.yml":       true,
				"/app/newrelic.js":        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := BaseConfigValidateHSM{
				getHSMConfiguration: tt.fields.getHSMConfiguration,
				envVars:             tt.fields.envVars,
			}
			if got := tr.GetHSMConfigurations(tt.args.configElements); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseConfigValidateHSM.GetHSMConfigurations() = %v, want %v", got, tt.want)
			}
		})
	}
}
