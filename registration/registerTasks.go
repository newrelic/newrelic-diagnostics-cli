package registration

import (
	"os"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	androidAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/android/agent"
	androidConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/android/config"
	androidLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/android/log"
	baseAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/agent"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/collector"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
	containers "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/containers"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/base/env"
	logTasks "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/log"
	browserAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/browser/agent"
	dotnetCoreAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/agent"
	dotnetCoreConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/config"
	dotnetCoreCustInst "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/custominstrumentation"
	dotnetCoreEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/env"
	dotnetCoreLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/log"
	dotnetCoreRequirements "github.com/newrelic/newrelic-diagnostics-cli/tasks/dotnetcore/requirements"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks/example/template"
	goAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/go/agent"
	goEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/go/env"
	goLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/go/log"
	iOSAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/iOS/agent"
	iOSConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/iOS/config"
	iOSEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/iOS/env"
	iOSLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/iOS/log"
	infraAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/agent"
	infraConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/config"
	infraEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/env"
	infraLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/infra/log"
	javaAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/agent"
	javaAppserver "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/appserver"
	javaConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/config"
	javaEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/env"
	javaJvm "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/jvm"
	javaLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/java/log"
	nodeAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/agent"
	nodeConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/config"
	nodeEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/env"
	nodeLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/log"
	nodeRequirements "github.com/newrelic/newrelic-diagnostics-cli/tasks/node/requirements"
	phpAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/php/agent"
	phpConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/php/config"
	phpDaemon "github.com/newrelic/newrelic-diagnostics-cli/tasks/php/daemon"
	phpEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/php/env"
	phpLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/php/log"
	pythonAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/python/agent"
	pythonConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/python/config"
	pythonEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/python/env"
	pythonLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/python/log"
	pythonRequirements "github.com/newrelic/newrelic-diagnostics-cli/tasks/python/requirements"
	rubyAgent "github.com/newrelic/newrelic-diagnostics-cli/tasks/ruby/agent"
	rubyConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/ruby/config"
	rubyEnv "github.com/newrelic/newrelic-diagnostics-cli/tasks/ruby/env"
	rubyLog "github.com/newrelic/newrelic-diagnostics-cli/tasks/ruby/log"
	rubyRequirements "github.com/newrelic/newrelic-diagnostics-cli/tasks/ruby/requirements"
	syntheticsMinion "github.com/newrelic/newrelic-diagnostics-cli/tasks/synthetics/minion"
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
	if strings.Contains(os.Args[0], "newrelic-diagnostics-cli") {
		template.RegisterWith(Register)
	}

	Work.WorkQueue = make(chan tasks.Task)
}
