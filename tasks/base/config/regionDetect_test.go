package config

import (
	"reflect"
	"sort"
	"testing"
)

var fakeEUKey1 = "eu01xx000000000000000000000000000000NRAL"

func Test_detectRegions(t *testing.T) {
	type args struct {
		licenseKeyToSources map[string][]string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "it should parse US data center",
			args: args{
				licenseKeyToSources: map[string][]string{
					"0000000000000000000000000000000000000000": {"myapp/newrelic/newrelic.yml"},
				},
			},
			want: []string{"us01"},
		},
		{
			name: "it should parse other data centers",
			args: args{
				licenseKeyToSources: map[string][]string{
					"0000000000000000000000000000000000000000": {"Nowherce"},
					"eu01xx0000000000000000000000000000000000": {"Nowherce"},
				},
			},
			want: []string{"eu01", "us01"},
		},
		{
			name: "it should only parse first instance of regex pattern",
			args: args{
				licenseKeyToSources: map[string][]string{
					"eu01xx0000000000000000000000000000000000": {"Nowherce"},
				},
			},
			want: []string{"eu01"},
		},
		{
			name: "it should parse EU data center",
			args: args{
				licenseKeyToSources: map[string][]string{
					"eu01xx0000000000000000000000000000000000": {"Nowherce"},
				},
			},
			want: []string{"eu01"},
		},
		{
			name: "it should only return unique regions",
			args: args{
				licenseKeyToSources: map[string][]string{
					"eu01xx0000000000000000000000000000000000": {"Nowherce"},
					"eu01xx000000000000000000000000000000NRAL": {"somewhereelse"},
				},
			},
			want: []string{"eu01"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectRegions(tt.args.licenseKeyToSources)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("detectRegions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseRegion(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "it should parse EU data center",
			args: args{
				key: fakeEUKey1,
			},
			want: "eu01",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRegion(tt.args.key); got != tt.want {
				t.Errorf("parseRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}
