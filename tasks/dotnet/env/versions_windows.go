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

// Queries the registry for .Net version and translates that to a more friendly form
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

// Queries the registry for .Net versions above 4.0 and translates that to a more friendly form
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
	return releaseKeyToVersion(value)
}

// Minimum Release key for each .NET Framework 4.5+ version, highest first.
// https://learn.microsoft.com/en-us/dotnet/framework/install/how-to-determine-which-versions-are-installed#minimum-version
var frameworkReleaseKeys = []struct {
	minRelease uint64
	version    string
}{
	{533320, "4.8.1"},
	{528040, "4.8"},
	{461808, "4.7.2"},
	{461308, "4.7.1"},
	{460798, "4.7"},
	{394802, "4.6.2"},
	{394254, "4.6.1"},
	{393295, "4.6"},
	{379893, "4.5.2"},
	{378675, "4.5.1"},
	{378389, "4.5"},
}

// highestKnownReleaseKey - the highest Release value shipped with 4.8.1; anything above it is a newer, unknown version
const highestKnownReleaseKey = 533325

// releaseKeyToVersion translates the v4\Full Release registry value into a .NET Framework version
func releaseKeyToVersion(release uint64) string {
	if release > highestKnownReleaseKey {
		return "4.8.1 or later"
	}
	for _, key := range frameworkReleaseKeys {
		if release >= key.minRelease {
			return key.version
		}
	}
	return ""
}
