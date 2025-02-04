package requirements

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotnetRequirementsMessagingServicesCheck - This struct defines the .Net Messaging Services check
type DotnetRequirementsMessagingServicesCheck struct {
	findFiles             func([]string, []string) []string
	getWorkingDirectories tasks.GetWorkingDirectoriesFunc
	getFileVersion        tasks.GetFileVersionFunc
	versionIsCompatible   tasks.VersionIsCompatibleFunc
}

type osFunc func() (string, error)

type getListDllsFunc func() ([]string, error)

// This data type will store and track info on the different Messaging Services
// This allows for the logic to be agnostic to the requirements of the
// individual Messaging Service requirements, only the structs for a Messaging Service should need to change if requirements change
type messagingService struct {
	name        string
	version     string
	installed   bool
	versionGood bool
}

// Names of dlls to check.
var msmqName = "System.Messaging.dll"        //no version
var nServiceBusName = "NServiceBus.Core.dll" //get version
var nServiceBusVersion = []string{"5.0-5.*"}
var rabbitMqName = "RabbitMQ.Client.dll" //get version
var rabbitMqVersion = []string{"3.5+", "4+"}

// Identifier - This returns the Category, Subcategory and Name of this task
func (p DotnetRequirementsMessagingServicesCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/MessagingServicesCheck")
}

// Explain - Returns the help text for this task
func (p DotnetRequirementsMessagingServicesCheck) Explain() string {
	return "Check Messaging Services compatibility with New Relic .NET agent"
}

// Dependencies - Returns the dependencies for this task.
func (p DotnetRequirementsMessagingServicesCheck) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within this task
func (p DotnetRequirementsMessagingServicesCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#messaging"

	// abort if it isn't installed
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}

	localDirs := p.getWorkingDirectories()

	if len(localDirs) < 1 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Unable to determine the " + tasks.ThisProgramFullName + " working and executable directory paths.",
		}
	}

	dllNames := []string{msmqName, nServiceBusName, rabbitMqName}

	dllList := p.findFiles(dllNames, localDirs)
	log.Debug("DLL list", dllList)

	if len(dllList) < 1 {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Did not find any dlls associated with Messaging Services supported by the .Net Agent",
		}
	}

	var messageSystems []messagingService
	//process found DLLs to struct
	for _, dll := range dllList {
		if strings.Contains(dll, msmqName) {
			message, err := p.checkSystemMessaging(dll)
			if err != nil {
				return tasks.Result{
					Status:  tasks.Error,
					Summary: "Error parsing SystemMessaging DLL version. Error was: " + err.Error(),
				}

			}
			messageSystems = append(messageSystems, message)
		} else if strings.Contains(dll, nServiceBusName) {
			message, err := p.checkNServiceBus(dll)
			if err != nil {
				return tasks.Result{
					Status:  tasks.Error,
					Summary: "Error parsing NServiceBus DLL version. Error was: " + err.Error(),
				}

			}
			messageSystems = append(messageSystems, message)

		} else if strings.Contains(dll, rabbitMqName) {
			message, err := p.checkRabbitMq(dll)
			if err != nil {
				return tasks.Result{
					Status:  tasks.Error,
					Summary: "Error parsing RabbitMQ DLL version. Error was: " + err.Error(),
				}

			}
			messageSystems = append(messageSystems, message)
		}
	}

	status, summary, payload := processMessagingServices(messageSystems)

	return tasks.Result{
		Status:  status,
		Summary: summary,
		Payload: payload,
	}

}

func (p DotnetRequirementsMessagingServicesCheck) checkNServiceBus(path string) (messagingService, error) {

	version, _ := p.getFileVersion(path)
	log.Debug("NServiceBus version: ", version)
	compatible, err := p.versionIsCompatible(version, nServiceBusVersion)
	if err != nil {
		return messagingService{
			name:        path,
			version:     version,
			installed:   true,
			versionGood: false,
		}, err
	}

	return messagingService{
		name:        path,
		version:     version,
		installed:   true,
		versionGood: compatible,
	}, nil

}

func (p DotnetRequirementsMessagingServicesCheck) checkRabbitMq(path string) (messagingService, error) {
	version, _ := p.getFileVersion(path)
	log.Debug("RabbitMQ version: ", version)
	compatible, err := p.versionIsCompatible(version, rabbitMqVersion)
	if err != nil {
		return messagingService{
			name:        path,
			version:     version,
			installed:   true,
			versionGood: false,
		}, err
	}
	return messagingService{
		name:        path,
		version:     version,
		installed:   true,
		versionGood: compatible,
	}, nil
}

func (p DotnetRequirementsMessagingServicesCheck) checkSystemMessaging(path string) (messagingService, error) {
	version, err := p.getFileVersion(path)
	log.Debug("SystemMessaging version: ", version)
	if err != nil {
		return messagingService{
			name:        path,
			version:     version,
			installed:   true,
			versionGood: false,
		}, err
	}
	return messagingService{
		name:        path,
		version:     version,
		installed:   true,
		versionGood: true,
	}, nil
}

func processMessagingServices(installedMessagingServices []messagingService) (tasks.Status, string, []string) {

	cntNoVersion := 0
	cntBadVersion := 0
	var summarySlice []string
	var payload []string
	//Need to check the results and figure out what status to set
	for _, ms := range installedMessagingServices {
		if ms.version == "" {
			temp := "No version of " + ms.name + " detected. Found empty version"
			summarySlice = append(summarySlice, temp)
			payload = append(payload, "Found "+ms.name+" with no version")
			cntNoVersion++
		} else if !ms.versionGood {
			temp := "Incompatible version of " + ms.name + " detected. Found version " + ms.version
			summarySlice = append(summarySlice, temp)
			payload = append(payload, "Found "+ms.name+" with version "+ms.version)
			cntBadVersion++
		}
		payload = append(payload, "Found "+ms.name+" with version "+ms.version+" ")

	}

	if cntBadVersion == 0 && cntNoVersion == 0 {
		//success
		status := tasks.Success
		summary := "All Messaging Services detected as compatible, see output.json for more details."
		return status, summary, payload
	} else if cntNoVersion == 0 {
		//failure
		status := tasks.Failure
		summary := "Incompatible Messaging Services detected, see output.json for more details. Detected the following Messaging Services as incompatible: \n"
		for _, s := range summarySlice {
			summary = summary + s + "\n"
		}
		return status, summary, payload
	}

	//warning
	status := tasks.Warning
	summary := "Couldn't get version of some Messaging Services dlls: \n"
	for _, s := range summarySlice {
		summary = summary + s + "\n"
	}
	return status, summary, payload
}
