package agent

import (
	"errors"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/new-relic-infrastructure-agent-170-0
var InfraCtlVerRequirement tasks.Ver = tasks.Ver{
	Major: 1,
	Minor: 7,
	Patch: 0,
	Build: 0,
}

func GetInfraCtlCmd(upstream map[string]tasks.Result) (string, error) {
	if upstream["Base/Env/CollectEnvVars"].Status != tasks.Info {
		return "", errors.New("upstream dependency failure")
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		return "", errors.New(tasks.AssertionErrorSummary)
	}

	programFilesPath, ok := envVars["ProgramFiles"]
	if !ok {
		return "", errors.New("environment variable 'ProgramFiles' not available")
	}

	return programFilesPath + `\New Relic\newrelic-infra\newrelic-infra-ctl.exe`, nil
}
