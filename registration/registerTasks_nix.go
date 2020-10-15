// +build linux darwin

package registration

import (
	"github.com/newrelic/NrDiag/tasks/java/jvm"
)

func init() {
	jvm.RegisterNixWith(Register)

}
