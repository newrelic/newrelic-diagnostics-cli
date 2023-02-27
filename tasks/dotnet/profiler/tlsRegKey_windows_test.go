//go:build windows
// +build windows

package profiler

import (
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func Test_validateTLSRegKeys(t *testing.T) {
	tests := []struct {
		name    string
		want    *TLSRegKey
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "First test",
			want:    &TLSRegKey{0, 1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateTLSRegKeys()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTLSRegKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateTLSRegKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_compareTLSRegKeys(t *testing.T) {
	type args struct {
		tlsRegKeys *TLSRegKey
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
				tlsRegKeys: &TLSRegKey{
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
				tlsRegKeys: &TLSRegKey{
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
				tlsRegKeys: &TLSRegKey{
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
				tlsRegKeys: &TLSRegKey{
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := compareTLSRegKeys(tt.args.tlsRegKeys); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("compareTLSRegKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateSchUseStrongCryptoKeys(t *testing.T) {
	on := 1
	//off := 0
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *int
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "First path",
			args: args{
				path: `SOFTWARE\WOW6432Node\Microsoft\.NETFramework\v4.0.30319`,
			},
			want: &on,
		},
		{
			name: "Second path",
			args: args{
				path: `SOFTWARE\Microsoft\.NETFramework\v4.0.30319`,
			},
			want: &on,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateSchUseStrongCryptoKeys(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSchUseStrongCryptoKeys() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("validateSchUseStrongCryptoKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
