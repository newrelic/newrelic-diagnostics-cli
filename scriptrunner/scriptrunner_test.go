package scriptrunner

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/mocks"
)

func TestRunner_addUUIDToFilename(t *testing.T) {
	type fields struct {
		Deps IRunnerDependencies
	}
	type args struct {
		savepath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "Add UUID to filename",
			fields: fields{
				Deps: &mocks.MockScriptRunner{},
			},
			args: args{
				savepath: "test.sh",
			},
			want: "test-1234-1234-1234-1234.sh",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sr := &Runner{
				Deps: tt.fields.Deps,
			}
			if got := sr.addUUIDToFilename(tt.args.savepath); got != tt.want {
				t.Errorf("Runner.addUUIDToFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
