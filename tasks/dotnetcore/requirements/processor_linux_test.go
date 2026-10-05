package requirements

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestCheckLinuxArch(t *testing.T) {
	tests := []struct {
		name           string
		procType       string
		platform       string
		dotnetVersions []string
		wantStatus     tasks.Status
		wantSummary    string
	}{
		{name: "x86_64", procType: "x86_64", platform: "ubuntu", dotnetVersions: []string{"3.1.32"}, wantStatus: tasks.Success, wantSummary: "Processor detected as x64."},
		{name: "amd64 uppercase", procType: "AMD64", platform: "ubuntu", wantStatus: tasks.Success, wantSummary: "Processor detected as x64."},
		{name: "aarch64 on .NET 8", procType: "aarch64", platform: "ubuntu", dotnetVersions: []string{"8.0.1"}, wantStatus: tasks.Success, wantSummary: "Processor detected as ARM64."},
		{name: "arm64 without detected versions", procType: "arm64", platform: "debian", wantStatus: tasks.Success, wantSummary: "Processor detected as ARM64."},
		{
			name:           "aarch64 on .NET Core 3.1",
			procType:       "aarch64",
			platform:       "ubuntu",
			dotnetVersions: []string{"3.1.32", "8.0.1"},
			wantStatus:     tasks.Warning,
			wantSummary:    "Processor detected as ARM64. The .NET agent only supports ARM64 on .NET 5 or later. Unsupported .NET version(s) detected: 3.1.32",
		},
		{
			name:           "aarch64 on Alpine",
			procType:       "aarch64",
			platform:       "alpine",
			dotnetVersions: []string{"8.0.1"},
			wantStatus:     tasks.Warning,
			wantSummary:    "Processor detected as ARM64. The .NET agent does not support ARM64 on Alpine Linux.",
		},
		{name: "32-bit arm", procType: "armv7l", platform: "debian", wantStatus: tasks.Failure},
		{name: "i686", procType: "i686", platform: "centos", wantStatus: tasks.Failure},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			result := checkLinuxArch(test.procType, test.platform, test.dotnetVersions)

			// Assert
			if result.Status != test.wantStatus {
				t.Errorf("Status = %v, want %v (summary %q)", result.Status, test.wantStatus, result.Summary)
			}
			if test.wantSummary != "" && result.Summary != test.wantSummary {
				t.Errorf("Summary = %q, want %q", result.Summary, test.wantSummary)
			}
		})
	}
}
