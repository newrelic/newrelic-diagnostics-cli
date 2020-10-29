package profiler

import (
	"strings"
	"fmt"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

var wasRegKeyPath = `SYSTEM\CurrentControlSet\Services\WAS`

type DotNetProfilerWasRegKey struct {
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetProfilerWasRegKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/WasRegKey")
}

// Explain - Returns the help text for each individual task
func (p DotNetProfilerWasRegKey) Explain() string {
	return "Validate WAS registry keys required for IIS-only .NET application profiling"
}

func (p DotNetProfilerWasRegKey) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

func (p DotNetProfilerWasRegKey) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		return tasks.Result{
			Status: tasks.None,
			Summary: "Did not detect .Net Agent as being installed, this check did not run",	
		}
	}

	return validateWasInstrumentationRegKeys()
}

func validateWasInstrumentationRegKeys() (result tasks.Result) {

	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, wasRegKeyPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug("WAS RegKey Check. Error opening WAS Reg Key. Error = ", string(err.Error()))
		return tasks.Result{
			Status: tasks.Error,
			Summary: "Could not open WAS Reg Key" + string(err.Error()),
		}
	}

	defer regKey.Close()

	regValues, _, regErr := regKey.GetStringsValue("Environment")

	if regErr != nil {
		log.Debug("WAS RegKey Check. Error opening Environment Sub Key. Error = ", string(regErr.Error()))
		return tasks.Result{
			Status: tasks.Warning,
			Summary: fmt.Sprintf("Unable to find WAS Registry keys needed for IIS hosted .NET app profiling set at: HKLM:\\%s\\Environment", wasRegKeyPath),
			URL: "https://docs.newrelic.com/docs/agents/net-agent/troubleshooting/profiler-conflicts",
		}
	}

	foundRegKeys := make(map[string]string)
	
	for _, regVal  := range regValues{
		kvPair := strings.Split(regVal, "=")
		if len(kvPair) != 2{
			continue
		}
		foundRegKeys[kvPair[0]] = kvPair[1]
	}

	regKeyErrors := []string{}
	for k, v := range expectedRegKeyWithVals {
		foundValue, ok := foundRegKeys[k]
		if !ok {
			err := fmt.Sprintf("%s was not set", k)
			regKeyErrors = append(regKeyErrors, err)
		} else if v != foundValue{
			err := fmt.Sprintf("%s was unexpectedly set to: '%s'. Expected: '%s'", k, foundValue, v)
			regKeyErrors = append(regKeyErrors, err)
		}
	}

	for _, k := range expectedRegKeyExists{
		foundValue, _ := foundRegKeys[k]
		if foundValue == "" {
			err := fmt.Sprintf("%s was not set", k)
			regKeyErrors = append(regKeyErrors, err)
		}
	}
	
	if len(regKeyErrors) > 0 {
		warningSummary := fmt.Sprintf("WAS registry keys needed for IIS hosted .NET app profiling are not correctly set. These should be located at: HKLM:\\%s\\Environment. Errors found:", wasRegKeyPath)
		for _, e := range regKeyErrors {
			warningSummary += "\n\t" + e
		}
		
		return tasks.Result{
			Status: tasks.Warning,
			Summary: warningSummary,
			URL: "https://docs.newrelic.com/docs/agents/net-agent/troubleshooting/profiler-conflicts#registry-keys",
			Payload: regValues,
		}
	}

	return tasks.Result{
		Status: tasks.Success,
		Summary: "WAS RegKeys needed for IIS hosted .Net App profiling are correctly set.",
		Payload: regValues,
	}
	
}
