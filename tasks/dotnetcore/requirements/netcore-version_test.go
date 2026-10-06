package requirements

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestNetCoreVersionExecute(t *testing.T) {
	tests := []struct {
		name        string
		upstream    map[string]tasks.Result
		wantStatus  tasks.Status
		wantSummary string
	}{
		{
			name: "agent not installed",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.None},
			},
			wantStatus:  tasks.None,
			wantSummary: "Did not detect .Net Core Agent as being installed, skipping this task.",
		},
		{
			name: "no .NET versions found",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Error},
			},
			wantStatus:  tasks.None,
			wantSummary: "Unable to determine versions of .NET Core installed, skipping this task.",
		},
		{
			name: "supported versions and compatible agent",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"8.0.100", "8.0.1", "10.0.0"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Info, Payload: "10.45.0.0"},
			},
			wantStatus:  tasks.Success,
			wantSummary: "Your .NET agent version 10.45.0.0 is compatible with the detected .NET version(s): 8.0.100, 8.0.1, 10.0.0",
		},
		{
			name: "supported versions without an agent version",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"6.0.25"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Error},
			},
			wantStatus:  tasks.Success,
			wantSummary: "Supported .NET version(s) detected: 6.0.25",
		},
		{
			name: "agent too old for the runtime",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"6.0.25", "8.0.1"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Info, Payload: "9.9.0.0"},
			},
			wantStatus:  tasks.Warning,
			wantSummary: "Your .NET agent version 9.9.0.0 is too old to support the following .NET version(s): 8.0.1",
		},
		{
			name: "unsupported runtime",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"1.1.2", "8.0.1"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Info, Payload: "10.45.0.0"},
			},
			wantStatus:  tasks.Warning,
			wantSummary: "One or more .NET versions are not supported by the New Relic .NET agent: 1.1.2",
		},
		{
			name: "unsupported runtime and agent too old",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"1.1.2", "9.0.0"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Info, Payload: "9.9.0.0"},
			},
			wantStatus: tasks.Warning,
			wantSummary: "One or more .NET versions are not supported by the New Relic .NET agent: 1.1.2\n" +
				"Your .NET agent version 9.9.0.0 is too old to support the following .NET version(s): 9.0.0",
		},
		{
			name: "no parseable versions",
			upstream: map[string]tasks.Result{
				"DotNetCore/Agent/Installed": {Status: tasks.Success},
				"DotNetCore/Env/Versions":    {Status: tasks.Info, Payload: []string{"not-a-version"}},
				"DotNetCore/Agent/Version":   {Status: tasks.Info, Payload: "10.45.0.0"},
			},
			wantStatus: tasks.Error,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Act
			result := DotNetCoreRequirementsNetCoreVersion{}.Execute(tasks.Options{}, test.upstream)

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
