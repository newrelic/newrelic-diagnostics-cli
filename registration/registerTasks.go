package registration

import (
	"os"
	"strings"

	"github.com/newrelic/NrDiag/tasks"
	androidAgent "github.com/newrelic/NrDiag/tasks/android/agent"
	androidConfig "github.com/newrelic/NrDiag/tasks/android/config"
	androidLog "github.com/newrelic/NrDiag/tasks/android/log"
	baseAgent "github.com/newrelic/NrDiag/tasks/base/agent"
	"github.com/newrelic/NrDiag/tasks/base/collector"
	"github.com/newrelic/NrDiag/tasks/base/config"
	containers "github.com/newrelic/NrDiag/tasks/base/containers"
	"github.com/newrelic/NrDiag/tasks/base/env"
	logTasks "github.com/newrelic/NrDiag/tasks/base/log"
	browserAgent "github.com/newrelic/NrDiag/tasks/browser/agent"
	dotnetCoreAgent "github.com/newrelic/NrDiag/tasks/dotnetcore/agent"
	dotnetCoreConfig "github.com/newrelic/NrDiag/tasks/dotnetcore/config"
	dotnetCoreCustInst "github.com/newrelic/NrDiag/tasks/dotnetcore/custominstrumentation"
	dotnetCoreEnv "github.com/newrelic/NrDiag/tasks/dotnetcore/env"
	dotnetCoreLog "github.com/newrelic/NrDiag/tasks/dotnetcore/log"
	dotnetCoreRequirements "github.com/newrelic/NrDiag/tasks/dotnetcore/requirements"
	"github.com/newrelic/NrDiag/tasks/example/template"
	goAgent "github.com/newrelic/NrDiag/tasks/go/agent"
	goEnv "github.com/newrelic/NrDiag/tasks/go/env"
	goLog "github.com/newrelic/NrDiag/tasks/go/log"
	iOSAgent "github.com/newrelic/NrDiag/tasks/iOS/agent"
	iOSConfig "github.com/newrelic/NrDiag/tasks/iOS/config"
	iOSEnv "github.com/newrelic/NrDiag/tasks/iOS/env"
	iOSLog "github.com/newrelic/NrDiag/tasks/iOS/log"
	infraAgent "github.com/newrelic/NrDiag/tasks/infra/agent"
	infraConfig "github.com/newrelic/NrDiag/tasks/infra/config"
	infraEnv "github.com/newrelic/NrDiag/tasks/infra/env"
	infraLog "github.com/newrelic/NrDiag/tasks/infra/log"
	javaAgent "github.com/newrelic/NrDiag/tasks/java/agent"
	javaAppserver "github.com/newrelic/NrDiag/tasks/java/appserver"
	javaConfig "github.com/newrelic/NrDiag/tasks/java/config"
	javaEnv "github.com/newrelic/NrDiag/tasks/java/env"
	javaJvm "github.com/newrelic/NrDiag/tasks/java/jvm"
	javaLog "github.com/newrelic/NrDiag/tasks/java/log"
	nodeAgent "github.com/newrelic/NrDiag/tasks/node/agent"
	nodeRequirements "github.com/newrelic/NrDiag/tasks/node/requirements"
	nodeConfig "github.com/newrelic/NrDiag/tasks/node/config"
	nodeEnv "github.com/newrelic/NrDiag/tasks/node/env"
	nodeLog "github.com/newrelic/NrDiag/tasks/node/log"
	phpAgent "github.com/newrelic/NrDiag/tasks/php/agent"
	phpConfig "github.com/newrelic/NrDiag/tasks/php/config"
	phpDaemon "github.com/newrelic/NrDiag/tasks/php/daemon"
	phpEnv "github.com/newrelic/NrDiag/tasks/php/env"
	phpLog "github.com/newrelic/NrDiag/tasks/php/log"
	pythonAgent "github.com/newrelic/NrDiag/tasks/python/agent"
	pythonConfig "github.com/newrelic/NrDiag/tasks/python/config"
	pythonEnv "github.com/newrelic/NrDiag/tasks/python/env"
	pythonLog "github.com/newrelic/NrDiag/tasks/python/log"
	pythonRequirements "github.com/newrelic/NrDiag/tasks/python/requirements"
	rubyAgent "github.com/newrelic/NrDiag/tasks/ruby/agent"
	rubyConfig "github.com/newrelic/NrDiag/tasks/ruby/config"
	rubyEnv "github.com/newrelic/NrDiag/tasks/ruby/env"
	rubyLog "github.com/newrelic/NrDiag/tasks/ruby/log"
	rubyRequirements "github.com/newrelic/NrDiag/tasks/ruby/requirements"
	syntheticsMinion "github.com/newrelic/NrDiag/tasks/synthetics/minion"
)

func init() {

	env.RegisterWith(Register)
	config.RegisterWith(Register)
	collector.RegisterWith(Register)
	logTasks.RegisterWith(Register)
	containers.RegisterWith(Register)
	javaJvm.RegisterWith(Register)
	phpDaemon.RegisterWith(Register)
	syntheticsMinion.RegisterWith(Register)
	javaConfig.RegisterWith(Register)
	javaAgent.RegisterWith(Register)
	javaLog.RegisterWith(Register)
	javaAppserver.RegisterWith(Register)
	goAgent.RegisterWith(Register)
	goLog.RegisterWith(Register)
	nodeConfig.RegisterWith(Register)
	nodeAgent.RegisterWith(Register)
	nodeLog.RegisterWith(Register)
	phpConfig.RegisterWith(Register)
	phpAgent.RegisterWith(Register)
	phpLog.RegisterWith(Register)
	pythonConfig.RegisterWith(Register)
	pythonAgent.RegisterWith(Register)
	pythonLog.RegisterWith(Register)
	pythonEnv.RegisterWith(Register)
	pythonRequirements.RegisterWith(Register)
	rubyConfig.RegisterWith(Register)
	rubyAgent.RegisterWith(Register)
	rubyLog.RegisterWith(Register)
	infraConfig.RegisterWith(Register)
	infraAgent.RegisterWith(Register)
	infraLog.RegisterWith(Register)
	infraEnv.RegisterWith(Register)
	androidConfig.RegisterWith(Register)
	androidAgent.RegisterWith(Register)
	androidLog.RegisterWith(Register)
	iOSConfig.RegisterWith(Register)
	iOSAgent.RegisterWith(Register)
	iOSLog.RegisterWith(Register)
	iOSEnv.RegisterWith(Register)
	dotnetCoreAgent.RegisterWith(Register)
	browserAgent.RegisterWith(Register)
	dotnetCoreConfig.RegisterWith(Register)
	dotnetCoreLog.RegisterWith(Register)
	rubyEnv.RegisterWith(Register)
	goEnv.RegisterWith(Register)
	nodeEnv.RegisterWith(Register)
	phpEnv.RegisterWith(Register)
	pythonEnv.RegisterWith(Register)
	dotnetCoreCustInst.RegisterWith(Register)
	javaEnv.RegisterWith(Register)
	dotnetCoreRequirements.RegisterWith(Register)
	javaEnv.RegisterWith(Register)
	dotnetCoreEnv.RegisterWith(Register)
	baseAgent.RegisterWith(Register)
	nodeRequirements.RegisterWith(Register)
	rubyRequirements.RegisterWith(Register)

	//example stuff, doesn't need to "ship" because binary gets name after directory with `go build` cmd
	if strings.Contains(os.Args[0], "NrDiag") {
		template.RegisterWith(Register)
	}

	Work.WorkQueue = make(chan tasks.Task)
}
