package config

import (
	"regexp"
	"strconv"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// When no region can be parsed using the agent spec provided regex, fall back
// to US region-- this is expected for legacy APM license keys.
const defaultRegion = "us01"

var regionLicenseRegex = regexp.MustCompile(`^([a-z]{2,3}[0-9]{2})x{1,2}`)

// BaseConfigRegionDetect - receiver struct for task definition
type BaseConfigRegionDetect struct {
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseConfigRegionDetect) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Config/RegionDetect")
}

// Explain - Returns the help text for each individual task
func (t BaseConfigRegionDetect) Explain() string {
	return "Determine New Relic region"
}

// Dependencies - Returns the dependencies for ech task.
func (t BaseConfigRegionDetect) Dependencies() []string {
	return []string{"Base/Config/ValidateLicenseKey"}
}

// Execute - Returns all datacenter regions detected from upstream license keys.
func (t BaseConfigRegionDetect) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	result := tasks.Result{
		Status:  tasks.None,
		Summary: "No New Relic license keys found -- unable to detect datacenter region.",
	}

	licenseKeyToSources, ok := upstream["Base/Config/ValidateLicenseKey"].Payload.(map[string][]string)

	if !ok {
		return result
	}

	if len(licenseKeyToSources) == 0 {
		return result
	}

	regions := detectRegions(licenseKeyToSources)

	result.Status = tasks.Info
	result.Summary = strconv.Itoa(len(regions)) + " unique New Relic region(s) detected from config."
	result.Payload = regions

	return result
}

func detectRegions(licenseKeyToSources map[string][]string) []string {
	detectedRegions := []string{}

	for lk := range licenseKeyToSources {
		detectedRegions = append(detectedRegions, parseRegion(lk))
	}

	detectedRegions = tasks.DedupeStringSlice(detectedRegions)
	return detectedRegions
}

func parseRegion(licenseKey string) string {
	parsedRegion := defaultRegion

	m := regionLicenseRegex.FindStringSubmatch(licenseKey)
	if len(m) > 1 {
		parsedRegion = m[1]
	}
	return parsedRegion
}
