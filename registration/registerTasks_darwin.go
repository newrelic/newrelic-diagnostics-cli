package registration

import (
	baseEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/jvm"
)

func init() {
	baseEnv.RegisterDarwinWith(Register)
	jvm.RegisterDarwinWith(Register)
}
