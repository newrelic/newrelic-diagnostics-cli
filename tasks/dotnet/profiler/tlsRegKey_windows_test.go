package profiler

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/domain/entity"
	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/stretchr/testify/mock"
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

func TestDotNetTLSRegKey_Execute(t *testing.T) {
	validateKeysMock := new(mocks.MockDotNetTLSRegKey)
	dotNetTLSRegKey := DotNetTLSRegKey{
		name:         "Reg key",
		validateKeys: validateKeysMock,
	}
	wantedErrors := []string{"Sch1", "Sch2", "TLS"}
	valid := 1
	invalid := 0
	type args struct {
		op       tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name    string
		p       DotNetTLSRegKey
		args    args
		want    tasks.Result
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "DotNet/Agent/Installed not successful",
			p:    dotNetTLSRegKey,
			args: args{
				op: tasks.Options{},
				upstream: map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status:  tasks.Success,
						Summary: "Regular Summary",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: wantedErrors[0],
			},
			wantErr: true,
		},
		{
			name: "DotNet/Agent/Installed not successful sch 2",
			p:    dotNetTLSRegKey,
			args: args{
				op: tasks.Options{},
				upstream: map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status:  tasks.Success,
						Summary: "Regular Summary",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: wantedErrors[1],
			},
			wantErr: true,
		},
		{
			name: "DotNet/Agent/Installed not successful TLS err",
			p:    dotNetTLSRegKey,
			args: args{
				op: tasks.Options{},
				upstream: map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status:  tasks.Success,
						Summary: "Regular Summary",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: wantedErrors[2],
			},
			wantErr: true,
		},
		{
			name: "DotNet/Agent/Installed not successful SCH not enabled",
			p:    dotNetTLSRegKey,
			args: args{
				op: tasks.Options{},
				upstream: map[string]tasks.Result{
					"DotNet/Agent/Installed": {
						Status:  tasks.Success,
						Summary: "Regular Summary",
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.Failure,
				Summary: "SchUseStrongCrypto must be enabled.  See more in these docs https://docs.newrelic.com/docs/apm/agents/net-agent/troubleshooting/no-data-appears-after-disabling-tls-10/#windows-registry",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				switch tt.want.Summary {
				case wantedErrors[0]:
					validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(nil, errors.New(wantedErrors[0])).Once()
				case wantedErrors[1]:
					validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(&valid, nil).Once()
					validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(nil, errors.New(wantedErrors[1])).Once()
				default:
					validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(&valid, nil).Once()
					validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(&valid, nil).Once()
					validateKeysMock.On("ValidateTLSRegKeys", mock.Anything).Return(nil, errors.New(wantedErrors[2])).Once()

				}
			} else {
				validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(&valid, nil).Once()
				validateKeysMock.On("ValidateSchUseStrongCryptoKeys", mock.Anything).Return(&invalid, nil).Once()
			}

			if got := tt.p.Execute(tt.args.op, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DotNetTLSRegKey.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
