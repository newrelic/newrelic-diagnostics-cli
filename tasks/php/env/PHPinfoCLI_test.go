package env

import (
	"errors"
	"reflect"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestPHPEnvPHPinfoCLI_Identifier(t *testing.T) {
	tests := []struct {
		name string
		p    PHPEnvPHPinfoCLI
		want tasks.Identifier
	}{
		{
			name: "Should return correct identifier",
			p:    PHPEnvPHPinfoCLI{},
			want: tasks.Identifier{
				Name:        "PHPinfoCLI",
				Category:    "PHP",
				Subcategory: "Env",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PHPEnvPHPinfoCLI{}
			if got := p.Identifier(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PHPEnvPHPinfoCLI.Identifier() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPHPEnvPHPinfoCLI_Explain(t *testing.T) {
	tests := []struct {
		name string
		p    PHPEnvPHPinfoCLI
		want string
	}{
		{
			name: "Should return correct explanation",
			p:    PHPEnvPHPinfoCLI{},
			want: "Collect PHP CLI configuration",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PHPEnvPHPinfoCLI{}
			if got := p.Explain(); got != tt.want {
				t.Errorf("PHPEnvPHPinfoCLI.Explain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPHPEnvPHPinfoCLI_Dependencies(t *testing.T) {
	tests := []struct {
		name string
		p    PHPEnvPHPinfoCLI
		want []string
	}{
		{
			name: "Should return correct list of dependencies",
			p:    PHPEnvPHPinfoCLI{},
			want: []string{"PHP/Config/Agent"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PHPEnvPHPinfoCLI{}
			if got := p.Dependencies(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PHPEnvPHPinfoCLI.Dependencies() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPHPEnvPHPinfoCLI_Execute(t *testing.T) {
	type fields struct {
		cmdExec tasks.CmdExecFunc
	}
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	emptyUpstream := map[string]tasks.Result{
		"PHP/Config/Agent": {
			Status: tasks.None,
		},
	}
	successfulUpstream := map[string]tasks.Result{
		"PHP/Config/Agent": {
			Status: tasks.Success,
		},
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   tasks.Result
	}{
		{
			name:   "Should return none result if upstream tasks.Result Status is not success",
			fields: fields{},
			args: args{
				upstream: emptyUpstream,
			},
			want: tasks.Result{
				Status:  tasks.None,
				Summary: "PHP Agent was not detected on this host. Skipping PHP info check.",
			},
		},
		{
			name: "Should return payload of string when PHP info is successful",
			fields: fields{
				cmdExec: mockPHPInfo,
			},
			args: args{
				upstream: successfulUpstream,
			},
			want: tasks.Result{
				Status:  tasks.Success,
				Summary: "PHP info has been gathered",
				Payload: "mockPHPInfo string",
			},
		},
		{
			name: "Should return tasks.Error if PHP info encounters an error",
			fields: fields{
				cmdExec: mockPHPError,
			},
			args: args{
				upstream: successfulUpstream,
			},
			want: tasks.Result{
				Status:  tasks.Error,
				Summary: "error executing PHP -i",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PHPEnvPHPinfoCLI{
				cmdExec: tt.fields.cmdExec,
			}
			if got := p.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PHPEnvPHPinfoCLI.Execute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func mockPHPInfo(name string, arg ...string) ([]byte, error) {
	return []byte("mockPHPInfo string"), nil
}

func mockPHPError(name string, arg ...string) ([]byte, error) {
	return []byte(""), errors.New("php info error")
}

func TestPHPEnvPHPinfoCLI_gatherPHPInfoCLI(t *testing.T) {
	type fields struct {
		cmdExec tasks.CmdExecFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Should return PHP info output as string",
			fields: fields{
				cmdExec: mockPHPInfo,
			},
			want: "mockPHPInfo string",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PHPEnvPHPinfoCLI{
				cmdExec: tt.fields.cmdExec,
			}
			if got, _ := p.gatherPHPInfoCLI(); got != tt.want {
				t.Errorf("PHPEnvPHPinfoCLI.gatherPHPInfoCLI() = %v, want %v", got, tt.want)
			}
		})
	}
}
