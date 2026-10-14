package config

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

func TestCheckConfigsAcceptsEnvironmentSpecificAppSettings(t *testing.T) {
	// Arrange
	configFiles := []config.ValidateElement{
		{
			Config: config.ConfigElement{FileName: "appsettings.Production.json", FilePath: "/app/"},
			ParsedResult: tasks.ValidateBlob{
				Children: []tasks.ValidateBlob{{Key: "NewRelic.AppName", RawValue: "my-app"}},
			},
		},
	}

	// Act
	validated, failed := checkConfigs(configFiles)

	// Assert
	if failed || len(validated) != 1 {
		t.Errorf("expected appsettings.Production.json to validate, got %d files, failed=%v", len(validated), failed)
	}
}
