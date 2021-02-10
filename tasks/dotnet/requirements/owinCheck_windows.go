package requirements

import (
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// DotnetRequirementsOwinCheck - This struct defines the OWIN check struct
type DotnetRequirementsOwinCheck struct {
	findFiles             func([]string, []string) []string
	getWorkingDirectories tasks.GetWorkingDirectoriesFunc
	getFileVersion        tasks.GetFileVersionFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotnetRequirementsOwinCheck) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Requirements/OwinCheck")
}

// Explain - Returns the help text for each individual task
func (p DotnetRequirementsOwinCheck) Explain() string {
	return "Check application's OWIN version compatibility with New Relic .NET agent"
}

// Dependencies - Returns the dependencies for ech task.
func (p DotnetRequirementsOwinCheck) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within each task
func (p DotnetRequirementsOwinCheck) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
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
	owinHosted, owinLoc := p.checkForOwin()
	if !owinHosted {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Didn't detect Owin dlls",
		}
	}

	owinVerString, err := p.getFileVersion(owinLoc)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Warning,
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#app-web-servers",
			Summary: "OWIN dlls detected but unable to confirm OWIN version. See debug logs for more information on error. Version returned " + err.Error(),
		}
	}

	supported, supportedErr := p.validateOwinVersion(owinVerString)

	if supportedErr != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error validating OWIN version, see debug logs for more details",
		}
	}
	if supported {
		return tasks.Result{
			Status:  tasks.Success,
			Summary: "Detected as OWIN hosted with version " + owinVerString,
		}
	}

	return tasks.Result{
		Status:  tasks.Failure,
		URL:     "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent#app-web-servers",
		Summary: "Detected OWIN dlls but version not detected as v3 or higher. Detected version is " + owinVerString,
	}

}

func (p DotnetRequirementsOwinCheck) checkForOwin() (bool, string) {
	fileKey := []string{"Microsoft.Owin.dll"}
	localDirs := p.getWorkingDirectories()
	owinFiles := p.findFiles(fileKey, localDirs)

	if len(owinFiles) > 0 {
		return true, owinFiles[0]
	}
	return false, ""
}

func (p DotnetRequirementsOwinCheck) validateOwinVersion(version string) (bool, error) {

	compatibleVersion := []string{"3+"}
	supported, err := tasks.VersionIsCompatible(version, compatibleVersion)
	if err != nil {
		return false, err
	}
	return supported, nil
}
