// +build linux

package registration

import (
	baseEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	infraEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/env"
)

func init() {
	baseEnv.RegisterLinuxWith(Register)
	infraEnv.RegisterLinuxWith(Register)
}
