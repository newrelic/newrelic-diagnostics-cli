// +build linux darwin

package registration

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/java/jvm"
)

func init() {
	jvm.RegisterNixWith(Register)

}
