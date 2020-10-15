// +build linux

package registration

import (
	"github.com/newrelic/NrDiag/tasks/base/env"
)

func init() {
	env.RegisterLinuxWith(Register)

}
