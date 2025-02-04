package requirements

import (
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// RegisterWinWith - will register any plugins in this package
func RegisterWinWith(registrationFunc func(tasks.Task, bool)) {
	log.Debug("Registering DotNet/Requirements/*")

	registrationFunc(DotnetRequirementsProcessorType{
		getProcessorArch: tasks.GetProcessorArch,
	}, true)
	registrationFunc(DotnetRequirementsOS{}, true)
	registrationFunc(DotnetRequirementsDatastores{
		findFiles:             tasks.FindFiles,
		getWorkingDirectories: tasks.GetWorkingDirectories,
		getFileVersion:        tasks.GetFileVersion,
	}, true)
	registrationFunc(DotnetRequirementsNetTargetAgentVerValidate{}, true)

	registrationFunc(DotnetRequirementsOwinCheck{
		findFiles:             tasks.FindFiles,
		getWorkingDirectories: tasks.GetWorkingDirectories,
		getFileVersion:        tasks.GetFileVersion,
	}, true)
	registrationFunc(DotnetRequirementsRequirementCheck{}, true)
	registrationFunc(DotnetRequirementsMessagingServicesCheck{
		findFiles:             tasks.FindFiles,
		getWorkingDirectories: tasks.GetWorkingDirectories,
		getFileVersion:        tasks.GetFileVersion,
		versionIsCompatible:   tasks.VersionIsCompatible,
	}, true)

}
