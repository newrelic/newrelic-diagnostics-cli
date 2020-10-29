// +build linux

package registration

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
)

func init() {
	env.RegisterLinuxWith(Register)

}
