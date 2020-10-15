// +build windows

package w3wp
import (
	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

// RegisterWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/w3wp/*")

	registrationFunc(DotNetW3wpCollect{}, true)
	registrationFunc(DotNetW3wpValidate{}, true)
}
