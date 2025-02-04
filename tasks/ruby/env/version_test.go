package env

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func mockRubyVExecuteSuccess(name string, arg ...string) ([]byte, error) {
	return []byte("ruby 0.0.0p0 (fake version for testing)"), nil
}

func mockRubyVExecuteFailure(name string, arg ...string) ([]byte, error) {
	return []byte{}, errors.New("execution error")
}

func TestRubyEnvVersion_checkRubyVersion(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name        string
		fields      fields
		wantResult  tasks.Result
		cmdExecutor tasks.CmdExecFunc
	}{

		{name: "should parse output from a successfully executed ruby -v", wantResult: tasks.Result{
			Status:  tasks.Info,
			Summary: "ruby 0.0.0p0 (fake version for testing)",
			URL:     "",
			Payload: "ruby 0.0.0p0 (fake version for testing)",
		},
			cmdExecutor: mockRubyVExecuteSuccess,
		},
		{name: "should return an error if ruby -v failed", wantResult: tasks.Result{
			Status:  tasks.Error,
			Summary: "Unable to execute command: $ ruby -v. Error: execution error",
			URL:     "https://docs.newrelic.com/docs/agents/ruby-agent/getting-started/ruby-agent-requirements-supported-frameworks#ruby_versions",
		},
			cmdExecutor: mockRubyVExecuteFailure,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := RubyEnvVersion{
				cmdExecutor: tt.cmdExecutor,
			}
			if gotResult := p.checkRubyVersion(); !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("RubyEnvVersion.checkRubyVersion() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestRubyEnvVersion_Execute(t *testing.T) {
	type fields struct {
		cmdExecutor tasks.CmdExecFunc
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
	}{
		{name: "should return None result if upstream dependency failed",
			fields: fields{},
			args: args{
				options: tasks.Options{},
				upstream: map[string]tasks.Result{
					"Ruby/Config/Agent": {
						Status: tasks.Failure,
					},
				},
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "Ruby Agent not installed. This task didn't run.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := RubyEnvVersion{
				cmdExecutor: tt.fields.cmdExecutor,
			}
			if got := p.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RubyEnvVersion.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}
