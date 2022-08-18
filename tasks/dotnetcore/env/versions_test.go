package env

import (
	"log"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func getMockInfoOutput() []byte {
	mock, err := os.ReadFile("../../../mocks/dotnetcore-info.txt")
	if err != nil {
		log.Fatal("Unable to load mock")
	}
	return mock
}

func mCmdExecGood(string, ...string) ([]byte, error) {
	return getMockInfoOutput(), nil
}

func getMockVersionsFromInfo() []string {
	return []string{
		"2.1.818",
		"3.1.416",
		"5.0.404",
		"6.0.101",
		"2.1.30",
		"3.1.22",
		"5.0.13",
		"6.0.1",
	}
}

func mCmdExecEmptyInfo(cmd string, opts ...string) ([]byte, error) {
	if strings.Contains(opts[0], "--info") {
		return []byte{}, nil
	}
	return getMockVersionOutput(), nil
}

func getMockVersionOutput() []byte {
	return []byte("6.0.1")
}

func getMockVersionsFromVersion() []string {
	return []string{
		"6.0.1",
	}
}

func TestDotNetCoreEnvVersions_Execute(t *testing.T) {
	upstreamPayload := make(map[string]string)
	upstreamPayload["DOTNET_INSTALL_PATH"] = "/usr/share/dotnet"

	upstream := make(map[string]tasks.Result)
	upstream["Base/Env/CollectEnvVars"] = tasks.Result{
		Status:  tasks.Info,
		Payload: upstreamPayload,
	}
	type args struct {
		options  tasks.Options
		upstream map[string]tasks.Result
	}
	tests := []struct {
		name    string
		args    args
		cmdExec tasks.CmdExecFunc
		want    tasks.Result
	}{
		{
			name: "DotNetCoreVersions dotnet --info test with versions returned",
			args: args{
				options:  tasks.Options{},
				upstream: upstream,
			},
			cmdExec: mCmdExecGood,
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: strings.Join(getMockVersionsFromInfo(), ", "),
				Payload: getMockVersionsFromInfo(),
			},
		},
		{
			name: "DotNetCoreVersions dotnet --info test with nothing returned, response from dotnet --version",
			args: args{
				options:  tasks.Options{},
				upstream: upstream,
			},
			cmdExec: mCmdExecEmptyInfo,
			want: tasks.Result{
				Status:  tasks.Info,
				Summary: strings.Join(getMockVersionsFromVersion(), ", "),
				Payload: getMockVersionsFromVersion(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := DotNetCoreEnvVersions{
				cmdExec: tt.cmdExec,
			}
			if got := d.Execute(tt.args.options, tt.args.upstream); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DotNetCoreEnvVersions.Execute() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
