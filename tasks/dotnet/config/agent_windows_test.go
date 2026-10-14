package config

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestDotNetConfigAgentNoConfigFilesCollected(t *testing.T) {
	// Arrange
	p := DotNetConfigAgent{}
	upstream := map[string]tasks.Result{
		"DotNet/Agent/Installed": {Status: tasks.Success, Summary: "Found dlls required by the .NET Agent, but no newrelic.config file was found"},
		"Base/Config/Validate":   {Status: tasks.None},
	}

	// Act
	result := p.Execute(tasks.Options{}, upstream)

	// Assert
	if result.Status != tasks.Warning {
		t.Errorf("expected Warning, got %v: %s", result.Status, result.Summary)
	}
}
