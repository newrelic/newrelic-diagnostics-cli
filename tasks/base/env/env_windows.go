package env

import (
	"os"
	"os/user"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWinWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering Base/Env/*")

	registrationFunc(BaseEnvCheckWindowsAdmin{
		fileOpener:     os.Open,
		getEnvVars:     tasks.GetShellEnvVars,
		getCurrentUser: user.Current,
		isUserAdmin: tasks.IsUserAdmin,
	}, true)
	registrationFunc(BaseEnvIisCheck{}, true)
}
