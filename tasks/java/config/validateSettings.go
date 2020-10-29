package config

import (
	"fmt"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/output/color"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

// JavaConfigValidateSettings - This struct defined the sample plugin which can be used as a starting point
type JavaConfigValidateSettings struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p JavaConfigValidateSettings) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/Config/ValidateSettings")
}

// Explain - Returns the help text for each individual task
func (p JavaConfigValidateSettings) Explain() string {
	return "This task validates the types of Java agent config values."
}

// Dependencies - Returns the dependencies for each task.
func (p JavaConfigValidateSettings) Dependencies() []string {
	return []string{
		"Java/Config/Agent",
	}
}

// Execute - The core work within each task
func (p JavaConfigValidateSettings) Execute(_ tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	result := tasks.Result{
		Status:  tasks.Success,
		Summary: "Validated all config files",
	}
	configs, ok := upstream["Java/Config/Agent"].Payload.([]config.ValidateElement)

	if !ok || len(configs) == 0 {
		result.Status = tasks.None
		result.Summary = "No config files found"
		return result
	}
	config := configs[0].ParsedResult.AsMap()
	spec := LoadSpec()
	validationProblems := make([]ValidationResult, 0)
	for key, val := range config {
		kind, present := spec[key]
		if !present {
			validationProblems = append(validationProblems, ValidationResult{Key: key, Value: val, Status: Unknown})
		} else {
			strKind := kind.(string)
			vr := ValidateSetting(val, strKind)
			if vr.Status == Invalid {
				vr.Kind = strKind
				vr.Key = key
				validationProblems = append(validationProblems, vr)
			}
		}
	}
	result.Payload = validationProblems
	for _, vr := range validationProblems {
		if vr.Status == Invalid {
			result.Status = tasks.Warning
			result.Summary = Summarize(validationProblems)
			result.URL = "https://docs.newrelic.com/docs/agents/java-agent/configuration/java-agent-configuration-config-file"
			break
		}
	}
	return result
}

func Summarize(vrs []ValidationResult) string {
	var lines []string

	for _, vr := range vrs {
		if vr.Status != Invalid {
			continue
		}
		key := color.ColorString(color.White, vr.Key)
		value := color.ColorString(color.LightRed, fmt.Sprintf("%v", vr.Value))
		message := color.ColorString(color.Yellow, vr.Message)
		lines = append(lines, fmt.Sprintf("    Problem with key %s with value %v:\n        %s", key, value, message))
	}
	return strings.Join(lines, "\n")

}
