package log

import (
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestLogLevelValidateIsCaseInsensitive(t *testing.T) {
	// Arrange
	levels := map[string]string{
		"a/newrelic.config": "DEBUG",
		"b/newrelic.config": "Finest",
		"c/newrelic.config": "info",
	}

	// Act
	result := logLevelValidate(levels)

	// Assert
	if result.Status != tasks.Success {
		t.Errorf("expected Success, got %v: %s", result.Status, result.Summary)
	}
}

func TestLogLevelValidateKeepsOriginalCaseInPayload(t *testing.T) {
	// Arrange
	levels := map[string]string{"a/newrelic.config": "DEBUG"}

	// Act
	result := logLevelValidate(levels)

	// Assert
	payload, ok := result.Payload.(map[string]string)
	if !ok || payload["a/newrelic.config"] != "DEBUG" {
		t.Errorf("expected original level in payload, got %v", result.Payload)
	}
}

func TestLogLevelValidateInvalidLevel(t *testing.T) {
	// Arrange
	levels := map[string]string{"a/newrelic.config": "chatty"}

	// Act
	result := logLevelValidate(levels)

	// Assert
	if result.Status != tasks.Failure {
		t.Errorf("expected Failure, got %v: %s", result.Status, result.Summary)
	}
}

func TestLogLevelValidateMixedValidAndInvalidWarns(t *testing.T) {
	// Arrange
	levels := map[string]string{
		"a/newrelic.config": "Warn",
		"b/newrelic.config": "",
	}

	// Act
	result := logLevelValidate(levels)

	// Assert
	if result.Status != tasks.Warning {
		t.Errorf("expected Warning, got %v: %s", result.Status, result.Summary)
	}
}
