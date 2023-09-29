//go:build linux || darwin
// +build linux darwin

package agent

import "github.com/newrelic/newrelic-diagnostics-cli/tasks"

// https://docs.newrelic.com/docs/release-notes/infrastructure-release-notes/infrastructure-agent-release-notes/new-relic-infrastructure-agent-140
var InfraCtlVerRequirement tasks.Ver = tasks.Ver{
	Major: 1,
	Minor: 4,
	Patch: 0,
	Build: 0,
}

func GetInfraCtlCmd(upstream map[string]tasks.Result) (string, error) {
	return "newrelic-infra-ctl", nil
}
