package agent

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func TestAndroidAgentDetect_Identifier(t *testing.T) {
	tests := []struct {
		name string
		p    AndroidAgentDetect
		want tasks.Identifier
	}{
		{
			name: "Should return success when Android/Agent/Detect is found as the task Identifier",
			p:    AndroidAgentDetect{},
			want: tasks.Identifier{
				Name:        "Detect",
				Category:    "Android",
				Subcategory: "Agent",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := AndroidAgentDetect{}
			if got := p.Identifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AndroidAgentDetect.Identifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAndroidAgentDetect_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		p    AndroidAgentDetect
		want []string
	}{
		{
			name: "Should return success when Base/Config/Collect is returned",
			p:    AndroidAgentDetect{},
			want: []string{"Base/Config/Collect"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := AndroidAgentDetect{}
			if got := p.Dependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AndroidAgentDetect.Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAndroidAgentDetect_Explain(t *testing.T) {
	tests := []struct {
		name string
		p    AndroidAgentDetect
		want string
	}{
		{name: "Should say -  Detects Android Environment",
			p:    AndroidAgentDetect{},
			want: "Detect if running in Android environment"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			if got := p.Explain(); got != tt.want {
				t.Errorf("AndroidAgentDetect.Explain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAndroidAgentDetect_Execute(t *testing.T) {
	// this struct reflects the two parameters for Execute
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}

	//mock successful upstream map
	successfulUpstream := make(map[string]tasks.Result)
	successfulUpstream["Base/Config/Collect"] = tasks.Result{
		Status: tasks.Success,
		Payload: []config.ConfigElement{
			{
				FileName: "AndroidManifest.xml",
				FilePath: "testpath",
			},
		},
	}

	successfulUpstreamNoConfig := make(map[string]tasks.Result)
	successfulUpstreamNoConfig["Base/Config/Collect"] = tasks.Result{
		Status:  tasks.Success,
		Payload: []config.ConfigElement{},
	}
	//mock failed upstream map
	failedUpstream := make(map[string]tasks.Result)
	failedUpstream["Base/Config/Collect"] = tasks.Result{
		Status: tasks.Failure,
		Payload: []config.ConfigElement{
			{
				FileName: "",
			},
		},
	}
	//wrong upstream type, causing error
	errorUpstream := make(map[string]tasks.Result)
	errorUpstream["Base/Config/Collect"] = tasks.Result{
		Status:  tasks.Success,
		Payload: []string{""},
	}
	tests := []struct {
		name string
		p    AndroidAgentDetect
		args args
		want tasks.Result
	}{
		{
			name: "Should return success when the AndroidManifest.xml is found ",
			p:    AndroidAgentDetect{},
			args: args{
				upstream: successfulUpstream,
			},
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: "Android environment detected",
			},
		},
		{
			name: "Should return tasks.none when upstream dependency failed ",
			p:    AndroidAgentDetect{},
			args: args{
				upstream: failedUpstream,
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Android environment not detected",
			},
		},
		{
			name: "Should return tasks.none  when  AndroidManifest.xml is not found ",
			p:    AndroidAgentDetect{},
			args: args{
				upstream: successfulUpstreamNoConfig,
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Android environment not detected",
			},
		},
		{
			name: "Should return tasks.error when error occurs checking the payload ",
			p:    AndroidAgentDetect{},
			args: args{
				upstream: errorUpstream,
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "Error occurred reading the Base/Config/Collect payload",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := AndroidAgentDetect{}
			if got := p.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AndroidAgentDetect.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
