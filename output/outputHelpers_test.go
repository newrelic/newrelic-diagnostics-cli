package output

import (
	"archive/zip"
	"os"
	"testing"

	"github.com/newrelic/NrDiag/tasks"
)

func Test_copyFilesToZip(t *testing.T) {
	zipFile := CreateZip()
	defer os.Remove("output.zip")
	type args struct {
		dst        *zip.Writer
		filesToZip []tasks.FileCopyEnvelope
	}
	tests := []struct {
		name string
		args args
	}{
		{"addFiles", args{zipFile, []tasks.FileCopyEnvelope{tasks.FileCopyEnvelope{Path: "tasks/fixtures/java/newrelic/newrelic.yml"}}}},
	}

	for _, tt := range tests {
		copyFilesToZip(tt.args.dst, tt.args.filesToZip)
	}

}
