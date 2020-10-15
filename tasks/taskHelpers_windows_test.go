package tasks

import (
	"reflect"
	"testing"
)

func TestGetVersion(t *testing.T) {
	files := []struct {
		path string
		want string
	}{
		{path: `fixtures/Fake.dll`, want: ""},
		{path: `fixtures/NewRelic.Agent.Extensions.dll`, want: "6.17.387.0"},
		{path: `fixtures/versionedtester.exe`, want: "1.0.0.2"},
		{path: `fixtures/unversionedtester.exe`, want: ""},
	}

	for _, file := range files {
		t.Run(file.path, func(t *testing.T) {
			if version, _ := GetFileVersion(file.path); !reflect.DeepEqual(version, file.want) {
				t.Errorf("GetFileVersion() = %v, want %v", version, file.want)
			}
		})
	}
}
