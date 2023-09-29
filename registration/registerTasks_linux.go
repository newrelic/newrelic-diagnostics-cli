//go:build linux
// +build linux

package registration

import (
	baseEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	infraEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/env"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/jvm"
)

func init() {
	baseEnv.RegisterLinuxWith(Register)
	infraEnv.RegisterLinuxWith(Register)
	jvm.RegisterLinuxWith(Register)
}
