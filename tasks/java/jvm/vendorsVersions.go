package jvm

import (
	"fmt"
	"regexp"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	"github.com/shirou/gopsutil/process"
)

/*
	List of supported JRE distributions
	The keys to this map are used verbatim to generate
	a regular expression in `extractVendorFromJavaExecutable`
	They should exactly match how they appear in the output
	of `java -version.`

	Any vendors not found in this map will be flagged as
	unsupported. Known unsupported vendors can be called out
	explicitly by using an empty slice of compatibility
	requirements.

*/
var supportedVersions = map[string][]string{
	// supported vendors
	"OpenJDK":    []string{"1.7-1.9.*", "7-14.*"},
	"HotSpot":    []string{"1.7-1.9.*", "7-14.*"},
	"JRockit":    []string{"1-1.6.0.50"},
	"Coretto":    []string{"1.8-1.9.*", "8-11.*"},
	"Zulu":       []string{"1.8-1.9.*", "8-12.*"},
	"IBM":        []string{"1.7-1.8.*", "7-8.*"},
	"Oracle":     []string{"1.5.*", "5.0.*"},
	"Zing":       []string{"1.8-1.9.*", "8-11.*"},
	"OpenJ9":     []string{"1.8-1.9.*", "8-13.*"},
	"Dragonwell": []string{"1.8-1.9.*", "8-11.*"},
}

//Supported only with Java agent 4.3.x:
var supportedForJavaAgent4 = map[string][]string{
	"Apple":   []string{"1.6.*", "6.*"},
	"IBM":     []string{"1.6.*", "6.*"},
	"HotSpot": []string{"1.6.*", "6.*"},
}

type supportabilityStatus int

//Constants for use by the supportabilityStatus enum
const (
	NotSupported supportabilityStatus = iota

	LegacySupported

	FullySupported
)

/* PIDInfo keeps track of the vendor/version/supported status and cmdlines of a PID */
type PIDInfo struct {
	PID       int32
	Vendor    string
	Version   string
	Supported supportabilityStatus
}

type JavaJVMVendorsVersions struct {
	findProcessByName func(string) ([]process.Process, error)
	cmdExec           func(name string, arg ...string) ([]byte, error)
	runtimeGOOS       string
	getCmdLineArgs    func(process.Process) (string, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p JavaJVMVendorsVersions) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Java/JVM/VendorsVersions")
}

// Explain - Returns the help text for this task
func (p JavaJVMVendorsVersions) Explain() string {
	return "Check Java process JVM compatibility with New Relic Java agent"
}

// Dependencies - this task requires no assumptions about the environment; therefore, depends on no prior task
func (p JavaJVMVendorsVersions) Dependencies() []string {
	return []string{}
}

/* Execute - iterates through list of running java processes and returns their respective supportability */
func (p JavaJVMVendorsVersions) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	/* obtain currently running Java procs as a slice of PIDInfo structs */
	log.Debug(p.Identifier(), "- Checking for running Java processes on this host")
	javaProcesses, err := p.findProcessByName("java")
	if err != nil {
		log.Debug(p.Identifier(), "- Error getting list of running Java processes. Error is: ", err.Error())

		return tasks.Result{
			Summary: fmt.Sprintf("The task %s encountered an error while detecting all running Java processes.", p.Identifier()),
			Status:  tasks.Error,
		}
	}

	/* there are not running java processes on this system */
	if len(javaProcesses) == 0 {
		return tasks.Result{
			Summary: "This task did not detect any running Java processes on this system.",
			Status:  tasks.None,
		}
	}

	var javaPIDInfos []PIDInfo

	for _, proc := range javaProcesses {
		unsupportedBasePID := PIDInfo{
			PID:       proc.Pid,
			Vendor:    "unknown",
			Version:   "unknown",
			Supported: NotSupported,
		}
		//Attempt to get cmd line arguments from java process
		cmdLineArgs, err := p.getCmdLineArgs(proc)

		if err != nil {
			javaPIDInfos = append(javaPIDInfos, unsupportedBasePID)

			//We entirely depend on cmd line args to do further vendor/version identification
			//bail if we don't have em
			continue
		}

		//Parses executable from cmdLineArgs
		javaExecutable := parseJavaExecutable(cmdLineArgs)
		var vendor, version string
		var ok bool

		if javaExecutable == "" {
			//fallback: get vendor details by parsing cmd line args
			vendor, version, ok = parseVendorDetailsByArgs(cmdLineArgs)

			if !ok {
				javaPIDInfos = append(javaPIDInfos, unsupportedBasePID)
				continue
			}
		} else {

			//Runs executable -version
			vendor, version, ok = p.parseVendorDetailsByExe(javaExecutable)

			if !ok {
				//fallback: get vendor details by parsing cmd line args
				vendor, version, ok = parseVendorDetailsByArgs(cmdLineArgs)

				if !ok {
					javaPIDInfos = append(javaPIDInfos, unsupportedBasePID)
					continue
				}
			}
		}

		javaPIDInfos = append(javaPIDInfos, PIDInfo{
			PID:       proc.Pid,
			Vendor:    vendor,
			Version:   version,
			Supported: p.determineSupportability(vendor, version),
		})

	}

	supportabilityCounts := getSupportabilityCounts(javaPIDInfos)

	taskStatus, taskSummary := determineSummaryStatus(supportabilityCounts)

	return tasks.Result{
		Status:  taskStatus,
		Summary: taskSummary,
		Payload: javaPIDInfos,
		URL:     "https://docs.newrelic.com/docs/agents/java-agent/getting-started/compatibility-requirements-java-agent#jvm",
	}

}

//getSupportabilityCounts takes a slice of []PIDInfos and returns the total counts present
// in the slice for each possible SupportabilityStatus
func getSupportabilityCounts(javaPIDInfos []PIDInfo) map[supportabilityStatus]int {
	counts := map[supportabilityStatus]int{
		NotSupported:    0,
		LegacySupported: 0,
		FullySupported:  0,
	}

	for _, pid := range javaPIDInfos {
		counts[pid.Supported]++
	}

	return counts
}

//determineSummaryStatusResult determines the task status and corresponding summary by
//examining the detected JVM vendor/versions. If any JVMs are found that
// are not supported by a modern or legacy agent, it will return a tasks.Failure.
// If all JVMs are supported, but a least one JVM requiring a legacy agent is found,
// it will return a tasks.Warning status.

func determineSummaryStatus(counts map[supportabilityStatus]int) (tasks.Status, string) {

	total := counts[NotSupported] + counts[LegacySupported] + counts[FullySupported]

	if total == 0 {
		return tasks.Error, "Java processes were found, but an error occurred determining supportability."
	}

	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("There are %d Java process(es) running on this server. We detected:", total))

	if counts[FullySupported] > 0 {
		summary.WriteString(fmt.Sprintf("\n%d process(es) with a vendor/version fully supported by the latest version of the Java Agent.", counts[FullySupported]))
	}

	if counts[LegacySupported] > 0 {
		summary.WriteString(fmt.Sprintf("\n%d process(es) with a vendor/version requiring a legacy version of the Java Agent.", counts[LegacySupported]))
	}

	if counts[NotSupported] > 0 {
		summary.WriteString(fmt.Sprintf("\n%d process(es) with an unsupported vendor/version combination.", counts[NotSupported]))
	}

	//Early return failure if any unsupported JVMs found
	if counts[NotSupported] > 0 {
		summary.WriteString("\nPlease see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.")
		return tasks.Failure, summary.String()
	}

	//Early return warning if any legacy supported JVMs found
	if counts[LegacySupported] > 0 {
		summary.WriteString("\nPlease see nrdiag-output.json results for the Java/JVM/VendorsVersions task for more details.")
		return tasks.Warning, summary.String()
	}

	return tasks.Success, summary.String()
}

func sanitizeCmdLineArg(cmdLineArg string) string {
	//Remove whitespace
	whitespaceTrimmed := strings.TrimSpace(cmdLineArg)
	//Remove comma at end of line
	commaTrimmed := strings.Replace(whitespaceTrimmed, ",", "", 1)
	//Remove single/double quotes
	return trimQuotes(commaTrimmed)
}
func trimQuotes(src string) string {
	var quoted = regexp.MustCompile("^['`\"](.*)['`\"]$")

	if quoted.Match([]byte(src)) {
		return quoted.ReplaceAllString(src, "$1")
	}
	return src
}

//parseJavaExecutable takes in command line arguments and returns first argument that is determined to be
// the java executable.
func parseJavaExecutable(cmdLineArgs string) string {

	//first pass, splitting on ' -'
	argsSplitByDash := strings.Split(cmdLineArgs, " -")
	for _, cmdLineArg := range argsSplitByDash {
		sanitizedCmdLineArg := sanitizeCmdLineArg(cmdLineArg)

		/* does the arg string end in java or java.exe(windows) */
		match, _ := regexp.MatchString(".*java.exe$|.*java$", sanitizedCmdLineArg)
		if match {
			return sanitizedCmdLineArg
		}
	}

	//second pass, splitting on ' '
	argsSplitBySpace := strings.Split(cmdLineArgs, " ")
	for _, cmdLineArg := range argsSplitBySpace {
		match, _ := regexp.MatchString(".*java$", cmdLineArg)
		if match {
			return cmdLineArg
		}
	}
	log.Debugf("Unable to parse Java executable from %s", cmdLineArgs)
	return ""
}

//getCmdLineArgs is a wrapper for dependency injecting proc.Cmdline in testing
func getCmdLineArgs(proc process.Process) (string, error) {
	return proc.Cmdline()
}

//parseVendorDetailsByArgs parses vendor and version directly from java process arguments list
//used as a fallback if java -version fails.
func parseVendorDetailsByArgs(cmdLineArgs string) (string, string, bool) {

	vendor := extractVendorFromArgs(cmdLineArgs)
	if vendor == "" {
		return "", "", false
	}

	version := extractVersionFromArgs(cmdLineArgs)
	if version == "" {
		return "", "", false
	}

	return vendor, version, true

}

//parseVendorDetailsByExe takes in a java executable path (e.g. /foo/bar/bin/java) and
//attempts to parse the vendor and version by running: /foo/bar/bin/java -version
func (p JavaJVMVendorsVersions) parseVendorDetailsByExe(javaExecutable string) (string, string, bool) {
	execOutputRaw, err := p.cmdExec(javaExecutable, "-version")
	if err != nil {
		log.Debug("Error running", javaExecutable, "-version:", err.Error())
		//Plan B. If the cmd `myfullpath/to/my/javabinary -version` fails, we can try the simple cmd: `java -version`
		cmdOutput, cmdErr := p.cmdExec("java", "-version")
		if cmdErr != nil {
			//We don't do any meaningful error interpretation downstream, so
			//just swallow the error, nod to it in the logs and return a !ok bool
			//parseVendorDetailsFromArgs
			log.Debug("Error running java -version:", cmdErr.Error())
			return "", "", false
		}
		execOutputRaw = cmdOutput
	}

	execOutputStr := string(execOutputRaw)
	log.Debug(javaExecutable, "-version:", execOutputStr)

	vendor := extractVendorFromJavaExecutable(execOutputStr)
	if vendor == "" {
		return "", "", false
	}

	version := extractVersionFromJavaExecutable(execOutputStr)
	if version == "" {
		return "", "", false
	}

	return vendor, version, true
}

/* check 1st and 2nd lines of the java -version output for version information */
func extractVersionFromJavaExecutable(execOutput string) (version string) {
	splitString := strings.Split(execOutput, "\n")
	if len(splitString) < 3 {
		log.Debugf("Unexpected number of elements found when splitting: %s", execOutput)
		return
	}
	matchVersionString := regexp.MustCompile(".*(version|build) [\"]{0,1}([0-9_.]*).*")
	/* check first line for version
	e.g. java version "1.8.0_144" */
	match := matchVersionString.FindStringSubmatch(splitString[0])
	if match != nil {
		version = match[2]
		log.Debugf("Detected version %s examining %s", version, splitString[0])
		return
	}
	log.Debugf("Unable to determine version from %s", splitString[0])
	log.Debug("Checking second line for JRE build version")
	/* fallback: check second line for JRE build version
	e.g. Java(TM) SE Runtime Environment (build 1.8.0_144-b01) */
	match = matchVersionString.FindStringSubmatch(splitString[1])

	if match != nil {
		version = match[2]
		log.Debugf("Detected version %s examining %s", version, splitString[1])
		return
	}

	log.Debugf("Unable to determine version examining %s", splitString[1])

	return
}

/* check 3rd line of the java -version output for vendor information */
func extractVendorFromJavaExecutable(execOutput string) (vendor string) {
	splitString := strings.Split(execOutput, "\n")
	if len(splitString) < 3 {
		log.Debugf("Unexpected number of elements found when splitting: %s", execOutput)
		return
	}

	uniqueVendors := map[string]interface{}{}
	for v := range supportedVersions {
		uniqueVendors[v] = struct{}{}
	}

	for v := range supportedForJavaAgent4 {
		uniqueVendors[v] = struct{}{}
	}

	var vendorList []string
	for v := range uniqueVendors {
		vendorList = append(vendorList, v)
	}

	//Construct vendor regex e.g.: .*(HotSpot|JRockit|OpenJDK|Zulu|OpenJ9|Zing).*
	vendorRegex := ".*(" + strings.Join(vendorList, "|") + ").*"
	matchVendorString := regexp.MustCompile(vendorRegex)

	/* check 3rd line for mention of vendor
	e.g. Java HotSpot(TM) 64-Bit Server VM (build 25.144-b01, mixed mode) */
	match := matchVendorString.FindStringSubmatch(splitString[2])
	if match != nil {
		vendor = match[1]
		return
	}

	log.Debugf("Unable to determine vendor examining %s", splitString[2])
	return
}

/* check a java proc's command line arguments for JRE version */
func extractVersionFromArgs(cmdLineArgs string) (version string) {
	/* split a single proc's command line args into a slice of strings */
	sliceCmdLineArgs := strings.Split(cmdLineArgs, " -")
	for _, cmdLineArg := range sliceCmdLineArgs {
		switch {
		/* e.g. -Djava.version=1.5.0_07 */
		case strings.Contains(cmdLineArg, "java.version"):
			matchVendorString := regexp.MustCompile(".*=([0-9_.]*)")
			version = matchVendorString.FindStringSubmatch(cmdLineArg)[1]
			return
		/* e.g. -Djava.runtime.version=1.5.0_07-164 */
		case strings.Contains(cmdLineArg, "java.runtime.version"):
			/* AIX */
			/* -Djava.runtime.version=pap6460sr16fp1ifx-20140908_01 (SR16 FP1) */
			if strings.Contains(cmdLineArg, "pap6460") || strings.Contains(cmdLineArg, "pap3260") {
				log.Debug("Detected AIX server with a process running Java 6 JVM")
				return "1.6"
			}
			if strings.Contains(cmdLineArg, "pap6470") || strings.Contains(cmdLineArg, "pap3270") {
				log.Debug("Detected AIX server with a process running Java 7 JVM")
				return "1.7"
			}
			if strings.Contains(cmdLineArg, "pap6480") || strings.Contains(cmdLineArg, "pap3280") {
				log.Debug("Detected AIX server with a process running Java 8 JVM")
				return "1.8"
			}
			/* End AIX checks */
			matchVendorString := regexp.MustCompile(".*=([0-9_.]*)")
			version = matchVendorString.FindStringSubmatch(cmdLineArg)[1]
			return
		/* e.g. -Djava.vm.version=Oracle JRockit(R) (R28.0.0-617-125986-1.6.0_17-20091215-2120-windows-x86_64, compiled mode) */
		case strings.Contains(cmdLineArg, "java.vm.version"):
			matchVendorString := regexp.MustCompile(".*([0-9]{1}\\.[0-9]{1}\\.[0-9]{1}[_0-9]{0,3}).*") //nolint
			version = matchVendorString.FindStringSubmatch(cmdLineArg)[1]
			return
		}
	}
	return
}

/* check a java proc's command line arguments for the JRE vendor and/or distribution */
func extractVendorFromArgs(cmdLineArgs string) (vendor string) {

	//splitting on just space here would break when: Djava.vm.name=Java HotSpot(TM)

	sliceCmdLineArgs := strings.Split(cmdLineArgs, " -")
	for _, cmdLineArg := range sliceCmdLineArgs {
		if strings.Contains(cmdLineArg, "java.vm.name") {
			/* IBM J9 */
			if strings.Contains(cmdLineArg, "IBM") {
				log.Debug("Detected IBM JVM")
				vendor = "IBM"
				return
			}
			/* IBM Classic VM */
			if strings.Contains(cmdLineArg, "Classic") {
				log.Debug("Detected IBM Classic JVM")
				vendor = "IBM"
				return
			}

			if strings.Contains(cmdLineArg, "HotSpot") {
				log.Debug("Detected Hotspot JVM")
				vendor = "HotSpot"
				return
			}

			if strings.Contains(cmdLineArg, "JRockit") {
				log.Debug("Detected JRockit JVM")
				vendor = "JRockit"
				return
			}
		}
		if strings.Contains(cmdLineArg, "java.vm.vendor") {
			//cmlineArg would look something like this: -Djava.vm.vendor=Oracle Corporation

			if strings.Contains(cmdLineArg, "Apple") {
				log.Debug("Detected Apple JVM")
				vendor = "Apple"
				return
			}
			if strings.Contains(cmdLineArg, "IBM") {
				log.Debug("Detected IBM JVM")
				vendor = "IBM"
			}
			if (strings.Contains(cmdLineArg, "Oracle")) && vendor != "HotSpot" {
				log.Debug("Detected Oracle JVM")
				vendor = "Oracle"
			}
		}
	}
	return
}

func (p JavaJVMVendorsVersions) determineSupportability(vendor string, version string) supportabilityStatus {
	if p.isFullySupported(vendor, version) {
		return FullySupported
	}

	if p.isLegacySupported(vendor, version) {
		return LegacySupported
	}

	return NotSupported
}

func isItCompatible(version string, supportedVersion []string) (supported bool) {
	version = strings.Replace(version, "_", ".", -1)

	isCompatible, err := tasks.VersionIsCompatible(version, supportedVersion)

	if err != nil {
		log.Debug("We could not parse the version: ", version)
		return false
	}

	if isCompatible {
		return true
	}

	return false
}

/* This should be kept up-to-date with our public-facing documentation */
func (p JavaJVMVendorsVersions) isFullySupported(vendor string, version string) bool {

	// We only support IBM in Linux environment
	if (vendor == "IBM") && (p.runtimeGOOS != "linux") {
		return false
	}

	requirements := supportedVersions[vendor]

	if requirements != nil {
		return isItCompatible(version, requirements)
	}

	return false

}

func (p JavaJVMVendorsVersions) isLegacySupported(vendor string, version string) bool {

	//We only support Apple hotspot in OS X environment
	if (vendor == "Apple") && (p.runtimeGOOS != "darwin") {
		return false
	}

	//We only support IBM in Linux environment
	if (vendor == "IBM") && (p.runtimeGOOS != "linux") {
		return false
	}

	requirements := supportedForJavaAgent4[vendor]

	if requirements != nil {
		return isItCompatible(version, requirements)
	}

	return false
}
