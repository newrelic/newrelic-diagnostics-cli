package output

import (
	"archive/zip"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type MockFileInfo struct {
	FileName    string
	IsDirectory bool
}

func (mfi MockFileInfo) Name() string       { return mfi.FileName }
func (mfi MockFileInfo) Size() int64        { return int64(8) }
func (mfi MockFileInfo) Mode() os.FileMode  { return os.ModePerm }
func (mfi MockFileInfo) ModTime() time.Time { return time.Now() }
func (mfi MockFileInfo) IsDir() bool        { return mfi.IsDirectory }
func (mfi MockFileInfo) Sys() interface{}   { return nil }

func MockWriteFunction(path string, info os.FileInfo, zipfile *zip.Writer) error {
	if path == "/test/test.txt" {
		return nil
	}
	return errors.New("writeFileToZip error")
}

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
		{"addFiles", args{zipFile, []tasks.FileCopyEnvelope{{Path: "tasks/fixtures/java/newrelic/newrelic.yml"}}}},
	}

	for _, tt := range tests {
		copyFilesToZip(tt.args.dst, tt.args.filesToZip)
	}

}

func TestWalkCopyFunction(t *testing.T) {
	zipFile := CreateZip()
	defer os.Remove("output.zip")
	// var fileInfo MockFileInfo
	fileInfo := MockFileInfo{
		FileName:    "test.txt",
		IsDirectory: false,
	}
	exeInfo := MockFileInfo{
		FileName:    "text.exe",
		IsDirectory: false,
	}
	dirInfo := MockFileInfo{
		FileName:    "test",
		IsDirectory: true,
	}
	type args struct {
		path    string
		info    os.FileInfo
		err     error
		zipfile *zip.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "Initial Error", args: args{path: "/test", info: fileInfo, err: errors.New("Test Error"), zipfile: zipFile}, wantErr: true},
		{name: "Path is a Dir", args: args{path: "/test", info: dirInfo, err: nil, zipfile: zipFile}, wantErr: false},
		{name: "Path is a Exe", args: args{path: "/test/text.exe", info: exeInfo, err: nil, zipfile: zipFile}, wantErr: true},
		{name: "WriteToZip Error", args: args{path: "/test/text.txt", info: fileInfo, err: nil, zipfile: zipFile}, wantErr: true},
		{name: "Full pass", args: args{path: "/test/test.txt", info: fileInfo, err: nil, zipfile: zipFile}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := WalkCopyFunction(tt.args.path, tt.args.info, tt.args.err, tt.args.zipfile, MockWriteFunction); (err != nil) != tt.wantErr {
				t.Errorf("WalkCopyFunction() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWalkSizeFunction(t *testing.T) {
	fileInfo := MockFileInfo{
		FileName:    "test.txt",
		IsDirectory: false,
	}
	type args struct {
		info os.FileInfo
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{name: "Test Size", args: args{info: fileInfo}, want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WalkSizeFunction(tt.args.info); got != tt.want {
				t.Errorf("WalkSizeFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getTruncatedScriptOutputString(t *testing.T) {
	type args struct {
		output []byte
	}
	tests := []struct {
		name   string
		args   args
		maxLen int
		want   string
		want1  bool
	}{
		{
			name: "No truncation",
			args: args{
				output: []byte("test"),
			},
			maxLen: 5,
			want:   "test",
			want1:  false,
		},
		{
			name: "Yes truncation",
			args: args{
				output: []byte("test"),
			},
			maxLen: 2,
			want:   "te",
			want1:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			OrigMaxScriptOutputLen := MaxScriptOutputLen
			MaxScriptOutputLen = tt.maxLen
			got, got1 := getTruncatedScriptOutputString(tt.args.output)
			if got != tt.want {
				t.Errorf("getTruncatedScriptOutputString() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getTruncatedScriptOutputString() got1 = %v, want %v", got1, tt.want1)
			}
			MaxScriptOutputLen = OrigMaxScriptOutputLen
		})
	}
}
