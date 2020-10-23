package registration

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/agent"
	dotnetConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/config"
	dotnetCustomInstrumentation "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/custominstrumentation"
	dotnetLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/log"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/profiler"
	netframeworkrequirements "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/requirements"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/w3wp"
	dotnetEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnet/env"
)

func init() {

	agent.RegisterWinWith(Register)
	profiler.RegisterWinWith(Register)
	w3wp.RegisterWinWith(Register)
	dotnetLog.RegisterWinWith(Register)
	env.RegisterWinWith(Register)
	dotnetConfig.RegisterWinWith(Register)
	dotnetCustomInstrumentation.RegisterWinWith(Register)
	netframeworkrequirements.RegisterWinWith(Register)
	dotnetEnv.RegisterWinWith(Register)

}
