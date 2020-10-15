package env

import (
	"reflect"
	"testing"

	"github.com/newrelic/NrDiag/tasks"
	"github.com/newrelic/NrDiag/tasks/base/config"
)

func Test_iOSEnvDetect_Identifier(t *testing.T) {
	tests := []struct {
		name string
		p    iOSEnvDetect
		want tasks.Identifier
	}{
		{
			name: "Should return success when iOSEnvDetect is returned as task Identifier",
			p:    iOSEnvDetect{},
			want: tasks.Identifier{
				Name:        "Detect",
				Category:    "iOS",
				Subcategory: "Env",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := iOSEnvDetect{}
			if got := p.Identifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("iOSEnvDetect.Identifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iOSEnvDetect_Explain(t *testing.T) {
	tests := []struct {
		name string
		p    iOSEnvDetect
		want string
	}{
		{
			name: "Should say - Detects iOS environment",
			p:    iOSEnvDetect{},
			want: "Detect if running in iOS environment",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := iOSEnvDetect{}
			if got := p.Explain(); got != tt.want {
				t.Errorf("iOSEnvDetect.Explain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iOSEnvDetect_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		p    iOSEnvDetect
		want []string
	}{
		{
			name: "Should return success when Base/Config/Collect is returned",
			p:    iOSEnvDetect{},
			want: []string{"Base/Config/Collect"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := iOSEnvDetect{}
			if got := p.Dependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("iOSEnvDetect.Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_iOSEnvDetect_Execute(t *testing.T) {
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	//mock successful upstream
	successfulUpstream := make(map[string]tasks.Result)
	successfulUpstream["Base/Config/Collect"] = tasks.Result{
		Payload: []config.ConfigElement{
			config.ConfigElement{
				FileName: "AppDelegate.swift",
			},
		},
	}

	//mock failed upstream
	failedUpstream := make(map[string]tasks.Result)
	failedUpstream["Base/Config/Collect"] = tasks.Result{
		Payload: []config.ConfigElement{
			config.ConfigElement{
				FileName: "",
			},
		},
	}

	//mock error when checking upstream

	errorUpstream := make(map[string]tasks.Result)
	errorUpstream["Base/Config/Collect"] = tasks.Result{
		Payload: []string{""},
	}

	tests := []struct {
		name string
		p    iOSEnvDetect
		args args
		want tasks.Result
	}{
		{
			name: "Should return success when the AppDelegate.swift is found",
			p:    iOSEnvDetect{},
			args: args{
				upstream: successfulUpstream,
			},
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: "iOS environment detected",
			},
		},
		{
			name: "Should return tasks.None  when the AppDelegate.swift is not found",
			p:    iOSEnvDetect{},
			args: args{
				upstream: failedUpstream,
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "iOS environment not detected",
			},
		},
		{
			name: "Should return tasks.Error  when there is an error checking the upstream payload",
			p:    iOSEnvDetect{},
			args: args{
				upstream: errorUpstream,
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Task did not meet requirements necessary to run: type assertion failure",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := iOSEnvDetect{}
			if got := p.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("iOSEnvDetect.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
