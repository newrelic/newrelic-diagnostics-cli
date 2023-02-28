//go:build windows
// +build windows

package profiler

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func Test_compareTLSRegKeys(t *testing.T) {
	type args struct {
		tlsRegKeys *entity.TLSRegKey
	}
	tests := []struct {
		name string
		args args
		want tasks.Result
	}{
		// TODO: Add test cases.
		{
			name: "Success Case",
			args: args{
				tlsRegKeys: &entity.TLSRegKey{
					Enabled:           1,
					DisabledByDefault: 0,
				},
			},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "You have TLS 1.2 enabled",
			},
		},
		{
			name: "Failure Case",
			args: args{
				tlsRegKeys: &entity.TLSRegKey{
					Enabled:           0,
					DisabledByDefault: 1,
				},
			},
			want: tasks.Result{
				Status:  tasks.Failure,
				Summary: "You have disabled TLS 1.2. Consult these docs to enable TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
			},
		},
		{
			name: "Warning Case 1",
			args: args{
				tlsRegKeys: &entity.TLSRegKey{
					Enabled:           1,
					DisabledByDefault: 1,
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Check your registry settings and consider enabling TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
			},
		},
		{
			name: "Warning Case 2",
			args: args{
				tlsRegKeys: &entity.TLSRegKey{
					Enabled:           0,
					DisabledByDefault: 0,
				},
			},
			want: tasks.Result{
				Status:  tasks.Warning,
				Summary: "Check your registry settings and consider enabling TLS 1.2 https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
			},
		},
	}
	p := DotNetTLSRegKey{
		name:         "test",
		validateKeys: new(mocks.MockDotNetTLSRegKey),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := p.compareTLSRegKeys(tt.args.tlsRegKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareTLSRegKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
