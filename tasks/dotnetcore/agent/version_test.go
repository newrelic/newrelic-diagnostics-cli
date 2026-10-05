package agent

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestDotNetCoreAgentVersionUpstreamNotSuccessful(t *testing.T) {
	// Arrange
	p := DotNetCoreAgentVersion{getFileVersion: func(string) (string, error) { return "10.45.0.0", nil }}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.None, Summary: "Unable to locate the .NET Core agent's installation files"},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.None || result.Summary != tasks.UpstreamFailedSummary+"DotNetCore/Agent/Installed" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDotNetCoreAgentVersionNoAgentDetected(t *testing.T) {
	// Arrange
	p := DotNetCoreAgentVersion{getFileVersion: func(string) (string, error) { return "10.45.0.0", nil }}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.None, Summary: tasks.NoAgentDetectedSummary},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.None || result.Summary != tasks.NoAgentUpstreamSummary+"DotNetCore/Agent/Installed" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDotNetCoreAgentVersionBadPayload(t *testing.T) {
	// Arrange
	p := DotNetCoreAgentVersion{getFileVersion: func(string) (string, error) { return "10.45.0.0", nil }}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.Success, Payload: 42},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.Error || result.Summary != tasks.AssertionErrorSummary {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDotNetCoreAgentVersionFileVersionError(t *testing.T) {
	// Arrange
	p := DotNetCoreAgentVersion{getFileVersion: func(string) (string, error) { return "", errors.New("no version information found") }}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.Success, Payload: "/usr/local/newrelic-dotnet-agent/"},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.Error || result.Summary != "Error finding .NET Core agent version" {
		t.Errorf("unexpected result: %+v", result)
	}
}

func TestDotNetCoreAgentVersionSuccess(t *testing.T) {
	// Arrange
	agentDir := filepath.Join("usr", "local", "newrelic-dotnet-agent")
	var requestedPath string
	p := DotNetCoreAgentVersion{getFileVersion: func(path string) (string, error) {
		requestedPath = path
		return "10.45.0.0", nil
	}}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.Success, Payload: agentDir},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.Info || result.Summary != "10.45.0.0" || result.Payload != "10.45.0.0" {
		t.Errorf("unexpected result: %+v", result)
	}
	if requestedPath != filepath.Join(agentDir, "NewRelic.Agent.Core.dll") {
		t.Errorf("read version from %q", requestedPath)
	}
}

func TestDotNetCoreAgentVersionReadsRealDll(t *testing.T) {
	// Arrange - the fixture stands in for NewRelic.Agent.Core.dll
	p := DotNetCoreAgentVersion{getFileVersion: func(string) (string, error) {
		return tasks.GetPEFileVersion(filepath.Join("..", "..", "fixtures", "NewRelic.Agent.Extensions.dll"))
	}}
	upstream := map[string]tasks.Result{
		"DotNetCore/Agent/Installed": {Status: tasks.Success, Payload: "unused"},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.Info || result.Payload != "6.17.387.0" {
		t.Errorf("unexpected result: %+v", result)
	}
}
