package config

import (
	"reflect"
	"testing"
)

func Test_userFlags_UsagePayload(t *testing.T) {

	type fields struct {
		Verbose            bool
		Interactive        bool
		Quiet              bool
		VeryQuiet          bool
		Help               bool
		Version            bool
		YesToAll           bool
		ShowOverrideHelp   bool
		AutoAttach         bool
		UsageOptOut        bool
		Proxy              string
		ProxyUser          string
		ProxyPassword      string
		Tasks              string
		ConfigFile         string
		Override           string
		OutputPath         string
		Filter             string
		BrowserURL         string
		AttachmentEndpoint string
		Suites             string
		Include            string
		APIKey             string
		Region             string
		Script             string
	}

	sampleFlags := fields{
		Verbose:            true,
		Interactive:        true,
		Quiet:              true,
		VeryQuiet:          false,
		Help:               false,
		Version:            true,
		YesToAll:           false,
		ShowOverrideHelp:   true,
		AutoAttach:         true,
		Proxy:              "string",
		ProxyUser:          "string",
		ProxyPassword:      "",
		Tasks:              "string",
		ConfigFile:         "string",
		Override:           "",
		OutputPath:         "",
		Filter:             "string",
		BrowserURL:         "string",
		AttachmentEndpoint: "string",
		Suites:             "string",
		Include:            "string",
		APIKey:             "string",
		Region:             "string",
		Script:             "string",
	}

	samplePreparedConfig := []ConfigFlag{
		{Name: "verbose", Value: true},
		{Name: "interactive", Value: true},
		{Name: "quiet", Value: true},
		{Name: "veryQuiet", Value: false},
		{Name: "help", Value: false},
		{Name: "version", Value: true},
		{Name: "yesToAll", Value: false},
		{Name: "showOverrideHelp", Value: true},
		{Name: "autoAttach", Value: true},
		{Name: "proxy", Value: true},
		{Name: "proxyUser", Value: true},
		{Name: "proxyPassword", Value: false},
		{Name: "tasks", Value: "string"},
		{Name: "configFile", Value: true},
		{Name: "override", Value: false},
		{Name: "outputPath", Value: false},
		{Name: "filter", Value: "string"},
		{Name: "browserURL", Value: true},
		{Name: "attachmentEndpoint", Value: true},
		{Name: "suites", Value: "string"},
		{Name: "include", Value: "string"},
		{Name: "region", Value: "string"},
		{Name: "script", Value: "string"},
	}

	tests := []struct {
		name   string
		fields fields
		want   []ConfigFlag
	}{
		{
			name:   "it should correctly parse receiver's flags to slice of ConfigFlag structs",
			fields: sampleFlags,
			want:   samplePreparedConfig,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := userFlags{
				Verbose:            tt.fields.Verbose,
				Interactive:        tt.fields.Interactive,
				Quiet:              tt.fields.Quiet,
				VeryQuiet:          tt.fields.VeryQuiet,
				Help:               tt.fields.Help,
				Version:            tt.fields.Version,
				YesToAll:           tt.fields.YesToAll,
				ShowOverrideHelp:   tt.fields.ShowOverrideHelp,
				AutoAttach:         tt.fields.AutoAttach,
				UsageOptOut:        tt.fields.UsageOptOut,
				Proxy:              tt.fields.Proxy,
				ProxyUser:          tt.fields.ProxyUser,
				ProxyPassword:      tt.fields.ProxyPassword,
				Tasks:              tt.fields.Tasks,
				ConfigFile:         tt.fields.ConfigFile,
				Override:           tt.fields.Override,
				OutputPath:         tt.fields.OutputPath,
				Filter:             tt.fields.Filter,
				BrowserURL:         tt.fields.BrowserURL,
				AttachmentEndpoint: tt.fields.AttachmentEndpoint,
				Suites:             tt.fields.Suites,
				Include:            tt.fields.Include,
				APIKey:             tt.fields.APIKey,
				Region:             tt.fields.Region,
				Script:             tt.fields.Script,
			}
			if got := f.UsagePayload(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("userFlags.UsagePayload() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_boolifyFlag(t *testing.T) {
	type args struct {
		inputFlag string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "It should return \"false\" for empty string",
			args: args{inputFlag: ""},
			want: false,
		},
		{
			name: "It should return \"true\" for any string",
			args: args{inputFlag: "this is a string"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := boolifyFlag(tt.args.inputFlag); got != tt.want {
				t.Errorf("boolifyFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_stringToRegion(t *testing.T) {
	type args struct {
		region string
	}
	tests := []struct {
		name string
		args args
		want Region
	}{
		{
			name: "lower case US string to region type",
			args: args{
				region: "us",
			},
			want: USRegion,
		},
		{
			name: "upper case US string to region type",
			args: args{
				region: "US",
			},
			want: USRegion,
		},
		{
			name: "lower case EU string to region type",
			args: args{
				region: "eu",
			},
			want: EURegion,
		},
		{
			name: "upper case EU string to region type",
			args: args{
				region: "EU",
			},
			want: EURegion,
		},
		{
			name: "blank region string to region type",
			args: args{
				region: "",
			},
			want: NoRegion,
		},
		{
			name: "unknown region string to region type",
			args: args{
				region: "unknown",
			},
			want: NoRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringToRegion(tt.args.region); got != tt.want {
				t.Errorf("stringToRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRegionFlagAndEnv(t *testing.T) {
	type args struct {
		regionFromFlag string
		regionFromEnv  string
	}
	tests := []struct {
		name string
		args args
		want Region
	}{
		{
			name: "Both provided - Flag gets priority",
			args: args{
				regionFromFlag: "us",
				regionFromEnv:  "eu",
			},
			want: USRegion,
		},
		{
			name: "Only env provided",
			args: args{
				regionFromFlag: "",
				regionFromEnv:  "eu",
			},
			want: EURegion,
		},
		{
			name: "Only flag provided",
			args: args{
				regionFromFlag: "eu",
				regionFromEnv:  "",
			},
			want: EURegion,
		},
		{
			name: "Nothing provided - default to US",
			args: args{
				regionFromFlag: "",
				regionFromEnv:  "",
			},
			want: USRegion,
		},
		{
			name: "Invalid env provided - use flag",
			args: args{
				regionFromFlag: "us",
				regionFromEnv:  "invalid",
			},
			want: USRegion,
		},
		{
			name: "Invalid flag provided - use env",
			args: args{
				regionFromFlag: "invalid",
				regionFromEnv:  "eu",
			},
			want: EURegion,
		},
		{
			name: "Invalid flag and env provided - use default US",
			args: args{
				regionFromFlag: "invalid",
				regionFromEnv:  "invalid",
			},
			want: USRegion,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRegionFlagAndEnv(tt.args.regionFromFlag, tt.args.regionFromEnv); got != tt.want {
				t.Errorf("parseRegionFlagAndEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
