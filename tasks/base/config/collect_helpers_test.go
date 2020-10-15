package config

import (
	"fmt"
	"testing"
)

func Test_getNETAgentConfigPathFromFile(t *testing.T) {

	tests := []struct {
		filepath  string
		want_path string
		want_err  bool
	}{
		{
			filepath:  "./fixtures/App_with_valid_path.config",
			want_path: "./fixtures/validate_basexml.config",
			want_err:  false,
		},
		{
			filepath:  "./fixtures/App_with_no_path.config",
			want_path: "",
			want_err:  false,
		},
		{
			filepath:  "./fixtures/App_with_invalid_path_noexist.config",
			want_path: "",
			want_err:  true,
		},
		{
			filepath:  "./fixtures/App_with_invalid_path_isdir.config",
			want_path: "",
			want_err:  true,
		},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %v", i), func(t *testing.T) {
			got, err := getNETAgentConfigPathFromFile(tt.filepath)
			gotErr := (err != nil)
			if got != tt.want_path {
				t.Errorf("getNETAgentConfigPathFromFile() got = '%v', want %v", got, tt.want_path)
			}
			if gotErr != tt.want_err {
				t.Errorf("getNETAgentConfigPathFromFile() gotErr = %v, want %v", gotErr, tt.want_err)
			}
		})
	}
}

func Test_isFileWithNETAgentConfigPath(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{filename: "app-me.exe.config", want: true},
		{filename: "web.config", want: true},
		{filename: "app.config", want: true},
		{filename: "bob.config", want: false},
		{filename: "app.cfg", want: false},
		{filename: "web.cfg", want: false},
		{filename: "config.ini", want: false},
		{filename: "newrelic.js", want: false},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprintf("Test #%v", i), func(t *testing.T) {
			if got := isFileWithNETAgentConfigPath(tt.filename); got != tt.want {
				t.Errorf("isFileWithNETAgentConfigPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
