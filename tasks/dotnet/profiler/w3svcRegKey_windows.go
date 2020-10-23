// +build windows

package profiler

import (
	"strings"
	"fmt"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

var w3svcRegPath = `SYSTEM\CurrentControlSet\Services\W3SVC\`

type DotNetProfilerW3svcRegKey struct {
	name string
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetProfilerW3svcRegKey) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Profiler/W3svcRegKey")
}

// Explain - Returns the help text for each individual task
func (p DotNetProfilerW3svcRegKey) Explain() string {
	return "Validate W3SVC registry keys required for IIS-only .NET application profiling"
}

// Dependencies - This task has no dependencies
func (p DotNetProfilerW3svcRegKey) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

func (p DotNetProfilerW3svcRegKey) Execute(op tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["DotNet/Agent/Installed"].Status != tasks.Success{
		return tasks.Result{
			Status: tasks.None,
			Summary: "Did not detect .Net Agent as being installed, this check did not run",
		}
	}

	return validateW3svcInstrumentationRegKeys()

}

func validateW3svcInstrumentationRegKeys() (result tasks.Result) {
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, w3svcRegPath, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug("W3svc RegKey Check. Error opening W3SVC Reg Key. Error = ", err.Error())
		return tasks.Result{
			Status: tasks.Error,
			Summary: "Could not open W3SVC Reg Key" + err.Error(),
		}
	}

	defer regKey.Close()
	regValues, _, regErr := regKey.GetStringsValue("Environment")
	
	if regErr != nil {
		return tasks.Result{
			Status: tasks.Warning,
			Summary: fmt.Sprintf("Unable to find W3SVC Registry keys needed for IIS hosted .NET app profiling set at: HKLM:\\%s\\Environment", w3svcRegPath),
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
			err := fmt.Sprintf("%s was unexpectedly set to: '%s'. Expected: '%s'", k, foundRegKeys[k], v)
			regKeyErrors = append(regKeyErrors, err)
		}
	}

	for _, k := range expectedRegKeyExists{
		val, _ := foundRegKeys[k]
		if val == "" {
			err := fmt.Sprintf("%s was not set", k)
			regKeyErrors = append(regKeyErrors, err)
		}
	}

	if len(regKeyErrors) > 0 {
		warningSummary := fmt.Sprintf("W3SVC registry keys needed for IIS hosted .NET app profiling are not correctly set. These should be located at: HKLM:\\%s\\Environment. Errors found:", w3svcRegPath)
		for _, e := range regKeyErrors{
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
		Summary: "W3SVC RegKeys needed for IIS hosted .Net App profiling are correctly set.",
		Payload: regValues,
	}


}
