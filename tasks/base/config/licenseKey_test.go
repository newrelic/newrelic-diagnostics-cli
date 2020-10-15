package config

import (
	"reflect"
	"sort"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/newrelic/NrDiag/tasks"
)

var expectedLicenseKeys = []string{
	"APM-LICENSE-KEY",
	"INFRA-LICENSE-KEY",
}

var expectedLicenseKeysFromEnv = []LicenseKey{
	{
		Value:  "APM-LICENSE-KEY",
		Source: "NEW_RELIC_LICENSE_KEY",
	},
	{
		Value:  "INFRA-LICENSE-KEY",
		Source: "NRIA_LICENSE_KEY",
	},
}

var expectedConfigLicenseKeys = []string{
	"license_key",
	"licenseKey",
	"-licenseKey",
	"newrelic.license",
}

func Test_getLicenseKeysFromEnv(t *testing.T) {
	sort.Strings(expectedLicenseKeys)

	type args struct {
		envVariables map[string]string
	}
	tests := []struct {
		name string
		args args
		want []LicenseKey
	}{
		{
			name: "should retrieve infrastructure license key",
			args: args{envVariables: map[string]string{
				"NRIA_LICENSE_KEY": "INFRA-LICENSE-KEY",
			}},
			want: []LicenseKey{
				{
					Value:  "INFRA-LICENSE-KEY",
					Source: "NRIA_LICENSE_KEY",
				},
			},
		},
		{
			name: "should retrieve RPM license key",
			args: args{envVariables: map[string]string{
				"NEW_RELIC_LICENSE_KEY": "APM-LICENSE-KEY",
			}},
			want: []LicenseKey{
				{
					Value:  "APM-LICENSE-KEY",
					Source: "NEW_RELIC_LICENSE_KEY",
				},
			},
		},
		{
			name: "should retrieve RPM and Infra license keys",
			args: args{envVariables: map[string]string{
				"NEW_RELIC_LICENSE_KEY": "APM-LICENSE-KEY",
				"NRIA_LICENSE_KEY":      "INFRA-LICENSE-KEY",
			}},
			want: expectedLicenseKeysFromEnv,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Order of return can be variable, so sort got and want first
			got := getLicenseKeysFromEnv(tt.args.envVariables)
			sort.SliceStable(got, func(i, j int) bool { return got[i].Value < got[j].Value })
			sort.SliceStable(tt.want, func(i, j int) bool { return tt.want[i].Value < tt.want[j].Value })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLicenseKeysFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getLicenseKeysFromConfig(t *testing.T) {
	sort.Strings(expectedConfigLicenseKeys)

	type args struct {
		configElements []ValidateElement
		configKeys     []string
	}
	tests := []struct {
		name string
		args args
		want []LicenseKey
	}{
		{
			name: "should retrieve non-zero-length APM license key",
			args: args{
				configElements: []ValidateElement{
					ValidateElement{
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.license",
							RawValue: "Schnauzer12",
						},
						Config: ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/app/",
						},
					},
				},
				configKeys: []string{
					"newrelic.license",
				},
			},
			want: []LicenseKey{
				{
					Value:  "Schnauzer12",
					Source: "/app/newrelic.ini",
				},
			},
		},
		{
			name: "should not return empty string slice if license key is empty string",
			args: args{
				configElements: []ValidateElement{
					ValidateElement{
						ParsedResult: tasks.ValidateBlob{
							Key:      "newrelic.license",
							RawValue: "",
						},
						Config: ConfigElement{
							FileName: "newrelic.ini",
							FilePath: "/app/",
						},
					},
				},
				configKeys: []string{
					"newrelic.license",
				},
			},
			want: []LicenseKey{
				{
					Value:  "",
					Source: "/app/newrelic.ini",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Order of return can be variable, so sort got and want first
			got := getLicenseKeysFromConfig(tt.args.configElements, tt.args.configKeys)
			sort.SliceStable(got, func(i, j int) bool { return got[i].Value < got[j].Value })
			sort.SliceStable(tt.want, func(i, j int) bool { return tt.want[i].Value < tt.want[j].Value })
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getLicenseKeysFromConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sanitizeLicenseKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "It strips whitespace from the key name",
			args: args{key: " dummy_license_key_here  	"},
			want: "dummy_license_key_here",
		},
		{
			name: "It strips single quotes from the key name",
			args: args{key: "'dummy_license_key_here'"},
			want: "dummy_license_key_here",
		},
		{
			name: "It strips double quotes from the key name",
			args: args{key: "\"dummy_license_key_here\""},
			want: "dummy_license_key_here",
		},
		{
			name: "It strips both quotes and spaces from the key name",
			args: args{key: "' \"   dummy_license_key_here  \"'"},
			want: "dummy_license_key_here",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeLicenseKey(tt.args.key); got != tt.want {
				t.Errorf("sanitizeLicenseKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_detectEnvLicenseKey(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "It detects a Ruby agent reference to an environment variable",
			args: args{key: "<%= ENV[\"LICENSE_KEY_VALUE_RUBY\"] %>"},
			want: true,
		},
		{
			name: "It detects a NodeJS agent reference to an environment variable",
			args: args{key: "process.env.LICENSE_KEY_VALUE_NODE"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectEnvLicenseKey(tt.args.key); got != tt.want {
				t.Errorf("detectEnvLicenseKey() = %v, want %v", got, tt.want)
			}
		})
	}
}


func mockOsGetenv(envKey string) string {
	// passthrough mock function standing in for os.Getenv
	return envKey
}
func Test_retrieveEnvLicenseKey(t *testing.T) {
	type args struct {
		keyEnvReference string
		readEnv         EnvReader
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "It should parse and retrieve a Ruby license key from the ENV",
			args: args{keyEnvReference: "<%= ENV[\"LICENSE_KEY_VALUE_RUBY\"] %>", readEnv: mockOsGetenv},
			want: "LICENSE_KEY_VALUE_RUBY",
		},
		{
			name: "It should parse and retrieve a NodeJS license key from the process.env",
			args: args{keyEnvReference: "process.env.LICENSE_KEY_VALUE_NODE", readEnv: mockOsGetenv},
			want: "LICENSE_KEY_VALUE_NODE",
		},
		{
			name:    "It should return an error if unable to parse a license key from the supplied string",
			args:    args{keyEnvReference: "String that can't be parsed", readEnv: mockOsGetenv},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retrieveEnvLicenseKey(tt.args.keyEnvReference, tt.args.readEnv)
			if (err != nil) != tt.wantErr {
				t.Errorf("retrieveEnvLicenseKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("retrieveEnvLicenseKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ = Describe("BaseConfigLicenseKey", func() {

	var p BaseConfigLicenseKey

	Describe("Execute()", func() {

		var (
			result   tasks.Result
			options  tasks.Options
			upstream map[string]tasks.Result
		)

		JustBeforeEach(func() {
			result = p.Execute(options, upstream)
		})

		Context("When multiple license key sources are found but all use same license key", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []ValidateElement{
							ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.license",
									RawValue: "Schnauzer12",
								},
								Config: ConfigElement{
									FileName: "newrelic.ini",
									FilePath: "/app/",
								},
							}, ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "license_key",
									RawValue: "Schnauzer12",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
						},
					},
					"Base/Env/CollectEnvVars": tasks.Result{
						Status: tasks.Success,
						Payload: map[string]string{
							"NEW_RELIC_LICENSE_KEY": "Schnauzer12",
						},
					},
				}
			})

			expectedResult := tasks.Result{
				Status:  tasks.Success,
				Summary: "1 unique New Relic license key(s) found.\n     'Schnauzer12' will be used by New Relic APM Agents",
				Payload: []LicenseKey{
					{
						Value:  "Schnauzer12",
						Source: "/app/newrelic.ini",
					},
					{
						Value:  "Schnauzer12",
						Source: "/etc/newrelic.yml",
					}, {
						Value:  "Schnauzer12",
						Source: "NEW_RELIC_LICENSE_KEY",
					},
				},
			}

			It("Should return a success status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("Should return a summary reporting unique license keys", func() {
				Expect(result.Summary).To(Equal(expectedResult.Summary))
			})

			It("Should have a payload of found LicenseKeys", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})

		})

		Context("When multiple license key sources are found and all use different license keys", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []ValidateElement{
							ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "newrelic.license",
									RawValue: "Banana",
								},
								Config: ConfigElement{
									FileName: "newrelic.ini",
									FilePath: "/app/",
								},
							}, ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "license_key",
									RawValue: "Schnauzer12",
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
						},
					},
					"Base/Env/CollectEnvVars": tasks.Result{
						Status: tasks.Success,
						Payload: map[string]string{
							"NEW_RELIC_LICENSE_KEY": "Banana",
						},
					},
				}
			})

			expectedResult := tasks.Result{
				Status:  tasks.Warning,
				Summary: "Multiple license keys detected:\n" + "     'Schnauzer12' from '/etc/newrelic.yml'\n     'Banana' from 'NEW_RELIC_LICENSE_KEY'\n\n     'Banana' will be used by New Relic APM Agents",
				Payload: []LicenseKey{
					{
						Value:  "Banana",
						Source: "/app/newrelic.ini",
					},
					{
						Value:  "Schnauzer12",
						Source: "/etc/newrelic.yml",
					}, {
						Value:  "Banana",
						Source: "NEW_RELIC_LICENSE_KEY",
					},
				},
			}

			It("Should return a warning status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("Should return a summary reporting unique license keys count along with each unique key", func() {
				Expect(result.Summary).To(ContainSubstring("Multiple license keys detected:"))
				Expect(result.Summary).To(ContainSubstring("'Schnauzer12' from '/etc/newrelic.yml'"))
				Expect(result.Summary).To(ContainSubstring("'Banana' from 'NEW_RELIC_LICENSE_KEY'"))
				Expect(result.Summary).To(ContainSubstring("'Banana' will be used by New Relic APM Agents"))
			})

			It("Should have a payload of found LicenseKeys", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})

		})

		Context("When multiple license key fields are found in the same file", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.Success,
						Payload: []ValidateElement{
							ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "root",
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob {
											Key:      "newrelic.license",
											RawValue: "Banana",
										},
										tasks.ValidateBlob {
											Key:      "newrelic.license",
											RawValue: "Banana",
										},
									},
								},
								Config: ConfigElement{
									FileName: "newrelic.ini",
									FilePath: "/app/",
								},
							}, ValidateElement{
								ParsedResult: tasks.ValidateBlob{
									Key:      "root",
									Children: []tasks.ValidateBlob{
										tasks.ValidateBlob {
											Key:      "license_key",
											RawValue: "Schnauzer12",
										},
										tasks.ValidateBlob {
											Key:      "license_key",
											RawValue: "Schnauzer13",
										},
									},
								},
								Config: ConfigElement{
									FileName: "newrelic.yml",
									FilePath: "/etc/",
								},
							},
						},
					},
				}
			})

			expectedResult := tasks.Result{
				Status:  tasks.Warning,
				Summary: "Multiple license keys detected:\n" + "     'Schnauzer12' from '/etc/newrelic.yml'\n      'Schnauzer13' from '/etc/newrelic.yml'\n     'Banana' from '/app/newrelic.ini'\n",
				Payload: []LicenseKey{
					{
						Value:  "Banana",
						Source: "/app/newrelic.ini",
					},
					{
						Value:  "Schnauzer12",
						Source: "/etc/newrelic.yml",
					}, 
					{
						Value:  "Schnauzer13",
						Source: "/etc/newrelic.yml",
					}, 
				},
			}

			It("Should return a warning status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("Should return a summary reporting unique license keys count along with each unique key", func() {
				Expect(result.Summary).To(ContainSubstring("Multiple license keys detected:"))
				Expect(result.Summary).To(ContainSubstring("'Schnauzer13' from '/etc/newrelic.yml'"))
				Expect(result.Summary).To(ContainSubstring("'Schnauzer12' from '/etc/newrelic.yml'"))
				Expect(result.Summary).To(ContainSubstring("'Banana' from '/app/newrelic.ini'"))
			})

			It("Should have a payload of found LicenseKeys", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})

		})

		Context("When no license keys are found", func() {
			BeforeEach(func() {
				options = tasks.Options{}
				upstream = map[string]tasks.Result{
					"Base/Config/Validate": tasks.Result{
						Status: tasks.None,
						Payload: []ValidateElement{},
					},
					"Base/Env/CollectEnvVars": tasks.Result{
						Status: tasks.Success,
						Payload: map[string]string{},
					},
				}
			})

			expectedResult := tasks.Result{
				Status:  tasks.Warning,
				Summary: "No New Relic licenses keys were found. Please ensure a license key is set in your New Relic agent configuration or environment.",
				Payload: []LicenseKey{},
			}

			It("Should return a warning status", func() {
				Expect(result.Status).To(Equal(expectedResult.Status))
			})

			It("Should return a summary reporting no license keys were found", func() {
				Expect(result.Summary).To(Equal(expectedResult.Summary))
			})

			It("Should have a payload of no license keys", func() {
				Expect(result.Payload).To(ConsistOf(expectedResult.Payload))
			})

		})

	})
})
