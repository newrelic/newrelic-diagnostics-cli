package custominstrumentation

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("<extension/>"), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestResolveNewRelicHome(t *testing.T) {
	tests := []struct {
		name        string
		goos        string
		installPath string
		envVars     map[string]string
		want        string
	}{
		{
			name:        "CORECLR_NEW_RELIC_HOME wins",
			goos:        "windows",
			installPath: `C:\Program Files\New Relic\.NET Agent\netcore\`,
			envVars:     map[string]string{"CORECLR_NEW_RELIC_HOME": `D:\agent`, "CORECLR_NEWRELIC_HOME": `E:\legacy`},
			want:        `D:\agent`,
		},
		{
			name:        "legacy CORECLR_NEWRELIC_HOME is used when the new name is unset",
			goos:        "linux",
			installPath: "/usr/local/newrelic-dotnet-agent",
			envVars:     map[string]string{"CORECLR_NEWRELIC_HOME": "/opt/nr"},
			want:        "/opt/nr",
		},
		{
			name:        "Windows falls back to the ProgramData home the MSI installs",
			goos:        "windows",
			installPath: `C:\Program Files\New Relic\.NET Agent\netcore\`,
			envVars:     map[string]string{"ProgramData": `C:\ProgramData`},
			want:        filepath.Join(`C:\ProgramData`, "New Relic", ".NET Agent"),
		},
		{
			name:        "Windows defaults ProgramData when it isn't in the environment",
			goos:        "windows",
			installPath: `C:\Program Files\New Relic\.NET Agent\netcore\`,
			envVars:     map[string]string{},
			want:        filepath.Join(`C:\ProgramData`, "New Relic", ".NET Agent"),
		},
		{
			name:        "Linux falls back to the install path",
			goos:        "linux",
			installPath: "/usr/local/newrelic-dotnet-agent",
			envVars:     map[string]string{},
			want:        "/usr/local/newrelic-dotnet-agent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			got := resolveNewRelicHome(tt.goos, tt.installPath, tt.envVars)

			// Assert
			if got != tt.want {
				t.Errorf("resolveNewRelicHome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExecuteCollectsFromNewRelicHomeExtensions(t *testing.T) {
	// Arrange
	home := t.TempDir()
	installPath := t.TempDir()
	extensions := filepath.Join(home, "extensions")
	writeFile(t, filepath.Join(extensions, "Custom.xml"))
	writeFile(t, filepath.Join(extensions, "netcore", "CoreOnly.xml"))
	writeFile(t, filepath.Join(extensions, "netframework", "FrameworkOnly.xml"))
	writeFile(t, filepath.Join(extensions, "NewRelic.Providers.Wrapper.Sql.Instrumentation.xml"))
	writeFile(t, filepath.Join(installPath, "extensions", "NotLoaded.xml"))

	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.Success, Payload: installPath},
		"Base/Env/CollectEnvVars":    {Status: tasks.Info, Payload: map[string]string{"CORECLR_NEW_RELIC_HOME": home}},
	}

	// Act
	result := DotNetCoreCustomInstrumentationCollect{}.Execute(tasks.Options{}, upstream)

	// Assert
	elements, _ := result.Payload.([]CustomInstrumentationElement)
	var got []string
	for _, e := range elements {
		got = append(got, e.FileName)
	}
	sort.Strings(got)
	want := []string{"CoreOnly.xml", "Custom.xml"}
	if result.Status != tasks.Success || len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Errorf("expected %v, got status %v and files %v", want, result.Status, got)
	}
}
