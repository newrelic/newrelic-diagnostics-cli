package registration

import (
	"github.com/newrelic/NrDiag/tasks/base/env"
	"github.com/newrelic/NrDiag/tasks/dotnet/agent"
	dotnetConfig "github.com/newrelic/NrDiag/tasks/dotnet/config"
	dotnetCustomInstrumentation "github.com/newrelic/NrDiag/tasks/dotnet/custominstrumentation"
	dotnetLog "github.com/newrelic/NrDiag/tasks/dotnet/log"
	"github.com/newrelic/NrDiag/tasks/dotnet/profiler"
	netframeworkrequirements "github.com/newrelic/NrDiag/tasks/dotnet/requirements"
	"github.com/newrelic/NrDiag/tasks/dotnet/w3wp"
	dotnetEnv "github.com/newrelic/NrDiag/tasks/dotnet/env"
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
