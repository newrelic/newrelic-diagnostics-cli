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
		AttachmentKey      string
		ConfigFile         string
		Override           string
		OutputPath         string
		Filter             string
		BrowserURL         string
		AttachmentEndpoint string
		Suites             string
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
		AttachmentKey:      "string",
		ConfigFile:         "string",
		Override:           "",
		OutputPath:         "",
		Filter:             "string",
		BrowserURL:         "string",
		AttachmentEndpoint: "string",
		Suites:             "string",
	}

	samplePreparedConfig := []ConfigFlag{
		ConfigFlag{Name: "verbose", Value: true},
		ConfigFlag{Name: "interactive", Value: true},
		ConfigFlag{Name: "quiet", Value: true},
		ConfigFlag{Name: "veryQuiet", Value: false},
		ConfigFlag{Name: "help", Value: false},
		ConfigFlag{Name: "version", Value: true},
		ConfigFlag{Name: "yesToAll", Value: false},
		ConfigFlag{Name: "showOverrideHelp", Value: true},
		ConfigFlag{Name: "autoAttach", Value: true},
		ConfigFlag{Name: "proxy", Value: true},
		ConfigFlag{Name: "proxyUser", Value: true},
		ConfigFlag{Name: "proxyPassword", Value: false},
		ConfigFlag{Name: "tasks", Value: "string"},
		ConfigFlag{Name: "attachmentKey", Value: "string"},
		ConfigFlag{Name: "configFile", Value: true},
		ConfigFlag{Name: "override", Value: false},
		ConfigFlag{Name: "outputPath", Value: false},
		ConfigFlag{Name: "filter", Value: "string"},
		ConfigFlag{Name: "browserURL", Value: true},
		ConfigFlag{Name: "attachmentEndpoint", Value: true},
		ConfigFlag{Name: "suites", Value: "string"},
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
				AttachmentKey:      tt.fields.AttachmentKey,
				ConfigFile:         tt.fields.ConfigFile,
				Override:           tt.fields.Override,
				OutputPath:         tt.fields.OutputPath,
				Filter:             tt.fields.Filter,
				BrowserURL:         tt.fields.BrowserURL,
				AttachmentEndpoint: tt.fields.AttachmentEndpoint,
				Suites:             tt.fields.Suites,
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
