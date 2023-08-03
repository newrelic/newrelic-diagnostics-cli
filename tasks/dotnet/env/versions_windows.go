package env

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"golang.org/x/sys/windows/registry"
)

const netVersionBaseLoc = `SOFTWARE\Microsoft\NET Framework Setup\NDP`
const netAbove4Loc = `SOFTWARE\Microsoft\NET Framework Setup\NDP\v4\Full\`

type DotNetEnvVersions struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t DotNetEnvVersions) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Env/Versions")
}

// Explain - Returns the help text for each individual task
func (t DotNetEnvVersions) Explain() string {
	return "Determine version(s) of .NET"
}

// Dependencies - Returns the dependencies for ech task.
func (t DotNetEnvVersions) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t DotNetEnvVersions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	versions := checkNetVersions()

	if versions == nil {

		return tasks.Result{
			Status:  tasks.Error,
			Summary: "Error opening registry keys or reading values",
		}
	}

	if len(versions) < 1 {
		return tasks.Result{
			Status:  tasks.Warning,
			Summary: "Opened registry keys, but did not find any versions of .NET installed",
			URL:     "https://docs.newrelic.com/docs/agents/net-agent/installation-configuration/install-net-agent",
		}
	}

	return tasks.Result{
		Status:  tasks.Info,
		Summary: strings.Join(versions, ", "),
		Payload: versions,
	}
}

//Queries the registry for .Net version and translates that to a more friendly form
func checkNetVersions() (versions []string) {
	v4OrAbove := false
	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, netVersionBaseLoc, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)
	if err != nil {
		//log.Debug("error opening .Net registry key")
		log.Debug(err)
		versions = nil
		return versions
	}

	versionsTemp, errSub := regKey.ReadSubKeyNames(0)

	if errSub != nil {
		log.Debug(errSub)

		versions = nil
		return versions
	}

	for _, ver := range versionsTemp {
		if ver == "v4" {
			v4OrAbove = true
		}

		if strings.Index(ver, "v") == 0 {
			versions = append(versions, strings.Replace(ver, "v", "", 1))
		}
	}

	if v4OrAbove {
		net40Plus := checkNetAbove4()
		if net40Plus != "" {
			versions = append(versions, net40Plus)
		}
	}
	return versions
}

//Queries the registry for .Net versions above 4.0 and translates that to a more friendly form
func checkNetAbove4() string {

	regKey, err := registry.OpenKey(registry.LOCAL_MACHINE, netAbove4Loc, registry.ENUMERATE_SUB_KEYS|registry.QUERY_VALUE)

	if err != nil {
		log.Debug(err)

		return ""
	}

	err = nil

	value, _, err := regKey.GetIntegerValue("Release")
	if err != nil {
		log.Debug(err)

		return ""
	}
	if value >= 460798 {
		return "4.7 or later"
	}
	if value >= 394802 {
		return "4.6.2"
	}
	if value >= 394254 {
		return "4.6.1"
	}
	if value >= 393295 {
		return "4.6"
	}
	if value >= 379893 {
		return "4.5.2"
	}
	if value >= 378675 {
		return "4.5.1"
	}
	if value >= 378389 {
		return "4.5"
	} else {
		return ""
	}

}
