package config

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
)

// Verbosity is the current log level
type Verbosity int

const (
	//Info indicates normal logging level
	Info Verbosity = iota
	//Verbose indicates additional logging (for troubleshooting the app)
	Verbose
)

type Region string

const (
	USRegion Region = "us"
	EURegion Region = "eu"
	NoRegion Region = "none"
)

// userFlags is a struct containing the commandline arguments passed in at runtime
type userFlags struct {
	Verbose            bool
	Interactive        bool
	Quiet              bool
	VeryQuiet          bool
	Help               bool
	Version            bool
	SkipVersionCheck   bool
	YesToAll           bool
	ShowOverrideHelp   bool
	AutoAttach         bool
	UsageOptOut        bool
	Run                bool
	ListScripts        bool
	Proxy              string
	ProxyUser          string
	ProxyPassword      string
	Tasks              string
	ConfigFile         string
	Override           string
	OutputPath         string
	Filter             string
	BrowserURL         string
	AttachmentEndpoint string
	Suites             string
	Include            string
	APIKey             string
	Region             string
	Script             string
	ScriptFlags        string
	K8sNamespace       string
	InNewRelicCLI      bool
}

type ConfigFlag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// MarshalJSON - custom JSON marshaling for this task, we'll strip out the passphrase to keep it only in memory, not on disk
func (f userFlags) MarshalJSON() ([]byte, error) {
	proxySpecified := false
	if f.Proxy != "" || f.ProxyPassword != "" || f.ProxyUser != "" {
		proxySpecified = true
	}

	return json.Marshal(&struct {
		Verbose          bool
		Quiet            bool
		VeryQuiet        bool
		YesToAll         bool
		ShowOverrideHelp bool
		AutoAttach       bool
		ProxySpecified   bool
		SkipVersionCheck bool
		Run              bool
		ListScripts      bool
		Tasks            string
		ConfigFile       string
		Override         string
		OutputPath       string
		Filter           string
		BrowserURL       string
		Suites           string
		APIKey           string
		Include          string
		Region           string
		Script           string
		ScriptFlags      string
		K8sNamespace     string
	}{
		Verbose:          f.Verbose,
		Quiet:            f.Quiet,
		VeryQuiet:        f.VeryQuiet,
		YesToAll:         f.YesToAll,
		ShowOverrideHelp: f.ShowOverrideHelp,
		AutoAttach:       f.AutoAttach,
		ProxySpecified:   proxySpecified,
		SkipVersionCheck: f.SkipVersionCheck,
		Run:              f.Run,
		ListScripts:      f.ListScripts,
		Tasks:            f.Tasks,
		ConfigFile:       f.ConfigFile,
		Override:         f.Override,
		OutputPath:       f.OutputPath,
		Filter:           f.Filter,
		BrowserURL:       f.BrowserURL,
		Suites:           f.Suites,
		Include:          f.Include,
		APIKey:           f.APIKey,
		Region:           f.Region,
		Script:           f.Script,
		ScriptFlags:      f.ScriptFlags,
		K8sNamespace:     f.K8sNamespace,
	})
}

// LogLevel is the current log level for output to the screen
var LogLevel Verbosity

// Flags story configuration information
var Flags = userFlags{}

// Version of the application
var Version string

// USUsageEndpoint is the haberdasher endpoint to which usage statistics are sent for US accounts
var USUsageEndpoint string

// EUUsageEndpoint is the haberdasher endpoint to which usage statistics are sent for EU accounts
var EUUsageEndpoint string

// EUUsageEndpoint is the haberdasher endpoint to which usage statistics are sent
var UsageEndpoint string

// USAttachmentEndpoint is the haberdasher endpoint to which attachments are sent for US accounts
var USAttachmentEndpoint string

// EUAttachmentEndpoint is the haberdasher endpoint to which attachments are sent for EU accounts
var EUAttachmentEndpoint string

// AttachmentEndpoint is the haberdasher endpoint to which attachments are sent
var AttachmentEndpoint string

// USHaberdasherURL is the base url for the Haberdasher service in the US
var USHaberdasherURL string

// EUHaberdasherURL is the base url for the Haberdasher service in the EU
var EUHaberdasherURL string

// HaberdasherURL is the base url for the Haberdasher service
var HaberdasherURL string

// BuildTimestamp stores when the build was done
var BuildTimestamp string

func ParseFlags() {
	// declaring the cmd arg Flags
	//
	// Define short option with no description, then long option with description
	// FOR EXAMPLE:
	// 		"v", false, ""
	// 		"verbose", false, "Display verbose logging during check execution. Off by default"

	defaultString := ""

	flag.BoolVar(&Flags.Verbose, "v", false, "alias for -verbose")
	flag.BoolVar(&Flags.Verbose, "verbose", false, "Display verbose logging during check execution. Off by default")

	flag.BoolVar(&Flags.Version, "version", false, "Display current program version. Take precedence over -skip-version-check")
	flag.BoolVar(&Flags.SkipVersionCheck, "skip-version-check", false, "Skips the automatic check for a newer version of the application.")

	flag.StringVar(&Flags.Tasks, "t", defaultString, "alias for -tasks")
	flag.StringVar(&Flags.Tasks, "tasks", defaultString, "Specific {name of task} - could be comma separated list and/or contain a wildcard (*)")

	flag.StringVar(&Flags.Suites, "s", defaultString, "alias for -suites")
	flag.StringVar(&Flags.Suites, "suites", defaultString, "Specific {name of task suite} - could be comma separated list. If you do '-h suites' it will list all diagnostic task suites that can be run.")

	flag.BoolVar(&Flags.AutoAttach, "a", false, "alias for -attach")
	flag.BoolVar(&Flags.AutoAttach, "attach", false, "Attach for automatic upload to New Relic account")

	flag.StringVar(&Flags.APIKey, "api-key", defaultString, "API Key from New Relic for upload to New Relic account")

	flag.BoolVar(&Flags.Help, "h", false, "alias for -help")
	flag.BoolVar(&Flags.Help, "help", false, "Displays full list of command line options. If you do '-h tasks' it will list all tasks that can be run.")

	flag.StringVar(&Flags.ConfigFile, "c", defaultString, "alias for -config-file")
	flag.StringVar(&Flags.ConfigFile, "config-file", defaultString, "Override default config file location. Can be used to specify either a folder to search in addition to the default folders or a specific config file")

	flag.StringVar(&Flags.Proxy, "p", defaultString, "alias for -proxy")
	flag.StringVar(&Flags.Proxy, "proxy", defaultString, "Proxy should be in the format http(s)://proxyIp:proxyPort Not necessary in most cases… will override config file if used)")

	flag.StringVar(&Flags.ProxyUser, "proxy-user", defaultString, "Proxy username, if necessary")
	flag.StringVar(&Flags.ProxyPassword, "proxy-pw", defaultString, "Proxy password, if necessary")

	flag.StringVar(&Flags.Override, "o", defaultString, "alias for -override")
	flag.StringVar(&Flags.Override, "override", defaultString, "Specify overrides for detected values. Format <Identifier>.<property>=<value> - example '-o Base/Config/Validate.agentLanguage=PHP'")

	flag.StringVar(&Flags.OutputPath, "output-path", filepath.FromSlash("./"), "Output directory for results. Files will be named 'nrdiag-output.json and nrdiag-output.zip.")

	flag.BoolVar(&Flags.YesToAll, "y", false, "alias for -yes")
	flag.BoolVar(&Flags.YesToAll, "yes", false, "Say 'yes' to any prompt that comes up while running.")

	flag.StringVar(&Flags.Filter, "filter", "success,warning,failure,error,info", "Filter results based on status. Accepted values: Success, Warning, Failure, Error, None or Info. Multiple values can be provided in comma separated list. e.g: \"Success,Warning,Failure\"")

	flag.BoolVar(&Flags.Quiet, "q", false, "Quiet output; only prints the high level results and not the explanatory output. Suppresses file addition warnings if '-y' is also used. Does not contradict '-v'")
	flag.BoolVar(&Flags.VeryQuiet, "qq", false, "Very quiet output; only prints a single summary line for output (implies '-q'). Suppresses file addition warnings if '-y' is also used. Does not contradict '-v'. Inclusion filters are ignored.")

	flag.StringVar(&Flags.BrowserURL, "browser-url", defaultString, "Specify a URL to check for the presence of a New Relic Browser agent")

	flag.StringVar(&Flags.K8sNamespace, "k8s-namespace", defaultString, "Specify a namespace to use when executing the kubectl command")

	flag.BoolVar(&Flags.UsageOptOut, "usage-opt-out", false, "Decline to send anonymous New Relic Diagnostic tool usage data to New Relic for this run")

	flag.StringVar(&Flags.Include, "include", defaultString, "Include a file or directory (including subdirectories) in the nrdiag-output.zip. Limit 4GB. To upload the results to New Relic also use the '-a' flag.")

	flag.StringVar(&Flags.Region, "r", defaultString, "alias for -region")
	flag.StringVar(&Flags.Region, "region", defaultString, "The region your New Relic account is in. Accepted values: EU or US. Case insensitive. (Default: US)")

	flag.BoolVar(&Flags.ListScripts, "list-scripts", false, "List available scripts")

	flag.StringVar(&Flags.Script, "script", defaultString, "View or run a script")

	flag.StringVar(&Flags.ScriptFlags, "script-flags", defaultString, "Use with -run -script to pass command line flags to the script")

	flag.BoolVar(&Flags.Run, "run", false, "Use with -script to run the script")

	//if first arg looks like it was build with `go build`, then we are testing against Haberdasher staging or localhost endpoint
	if strings.Contains(os.Args[0], "newrelic-diagnostics-cli") {
		flag.StringVar(&Flags.AttachmentEndpoint, "attachment-endpoint", defaultString, "The endpoint to send attachments to. (NR ONLY)")
	}

	flag.Parse()

	if Flags.VeryQuiet {
		Flags.Quiet = true

		//pseudo-filter to force everything to display
		Flags.Filter = ""
	}

	// This has to be in the config init otherwise you don't get logs as expected
	if Flags.Verbose {
		LogLevel = Verbose
	} else {
		LogLevel = Info
	}

	if Flags.BrowserURL != "" {
		Flags.Override = "Browser/Agent/GetSource.url=" + Flags.BrowserURL + "," + Flags.Override
		Flags.Tasks = "Browser/Agent/Detect," + Flags.Tasks
	}

	if Flags.K8sNamespace != "" {
		Flags.Override = "K8s/Infra/Config.namespace=" + Flags.K8sNamespace + "," + Flags.Override
		Flags.Override = "K8s/Infra/Daemonset.namespace=" + Flags.K8sNamespace + "," + Flags.Override
		Flags.Override = "K8s/Infra/Deploy.namespace=" + Flags.K8sNamespace + "," + Flags.Override
		Flags.Override = "K8s/Infra/Logs.namespace=" + Flags.K8sNamespace + "," + Flags.Override
		Flags.Override = "K8s/Infra/Pods.namespace=" + Flags.K8sNamespace + "," + Flags.Override
	}

	// Set the endpoints based on region
	switch parseRegionFlagAndEnv(Flags.Region, os.Getenv("NEW_RELIC_REGION")) {
	case EURegion:
		UsageEndpoint = EUUsageEndpoint
		// Only set AttachmentEndpoint if the `-attachment-endpoint` flag was not used
		if Flags.AttachmentEndpoint == "" {
			AttachmentEndpoint = EUAttachmentEndpoint
		}
		HaberdasherURL = EUHaberdasherURL
	default:
		UsageEndpoint = USUsageEndpoint
		if Flags.AttachmentEndpoint == "" {
			AttachmentEndpoint = USAttachmentEndpoint
		}
		HaberdasherURL = USHaberdasherURL
	}

	Flags.InNewRelicCLI = (os.Getenv("NEWRELIC_CLI_SUBPROCESS") != "")
}

// UsagePayload gathers and sanitizes user command line input
// A map with string keys and interface values is returned
// The interface values will contain either a boolean or a string
func (f userFlags) UsagePayload() []ConfigFlag {
	return []ConfigFlag{
		{Name: "verbose", Value: f.Verbose},
		{Name: "interactive", Value: f.Interactive},
		{Name: "quiet", Value: f.Quiet},
		{Name: "veryQuiet", Value: f.VeryQuiet},
		{Name: "help", Value: f.Help},
		{Name: "version", Value: f.Version},
		{Name: "yesToAll", Value: f.YesToAll},
		{Name: "showOverrideHelp", Value: f.ShowOverrideHelp},
		{Name: "autoAttach", Value: f.AutoAttach},
		{Name: "proxy", Value: boolifyFlag(f.Proxy)},
		{Name: "proxyUser", Value: boolifyFlag(f.ProxyUser)},
		{Name: "proxyPassword", Value: boolifyFlag(f.ProxyPassword)},
		{Name: "tasks", Value: f.Tasks},
		{Name: "configFile", Value: boolifyFlag(f.ConfigFile)},
		{Name: "override", Value: boolifyFlag(f.Override)},
		{Name: "outputPath", Value: boolifyFlag(f.OutputPath)},
		{Name: "filter", Value: f.Filter},
		{Name: "browserURL", Value: boolifyFlag(f.BrowserURL)},
		{Name: "attachmentEndpoint", Value: boolifyFlag(f.AttachmentEndpoint)},
		{Name: "suites", Value: f.Suites},
		{Name: "include", Value: f.Include},
		{Name: "region", Value: f.Region},
		{Name: "script", Value: f.Script},
		{Name: "k8sNamespace", Value: f.K8sNamespace},
	}
}

// boolifyFlag is a helper function for falsey/truthy conversion of UserFlag strings
func boolifyFlag(inputFlag string) bool {
	return inputFlag != ""
}

// IsForcedTask returns true if the supplied task (identifier) was supplied in the
// -t command line argument.
func (f userFlags) IsForcedTask(identifier string) bool {
	identifiers := strings.Split(f.Tasks, ",")
	for _, ident := range identifiers {
		trimmedIdentifier := strings.TrimSpace(ident)
		if strings.EqualFold(identifier, trimmedIdentifier) {
			return true
		}
	}
	return false
}

// parseRegionFlagAndEnv - Parse region flag and region env variable, determine which to use.
// Prioritize in this order:
// - Use the command line flag if that is provided
// - If flag is not provided, use env variable NEW_RELIC_REGION, if present
// - If neither the flag or env variable are present, default to US
func parseRegionFlagAndEnv(regionFromFlag string, regionFromEnv string) Region {
	regionFlag := stringToRegion(regionFromFlag)
	regionEnv := stringToRegion(regionFromEnv)

	// No region provided, use default US
	if regionFlag == NoRegion && regionEnv == NoRegion {
		return USRegion
	}
	// Both the command line flag and the env variable were provided, priority is flag
	if regionFlag != NoRegion && regionEnv != NoRegion {
		return regionFlag
	}
	// Region was provided via env
	if regionFlag == NoRegion && regionEnv != NoRegion {
		return regionEnv
	}
	// Region was provided via flag
	if regionFlag != NoRegion && regionEnv == NoRegion {
		return regionFlag
	}
	// if we somehow got here, go with default US
	return USRegion
}

// stringToRegion - converts region string to Region type
func stringToRegion(region string) Region {
	r := strings.TrimSpace(strings.ToLower(region))
	if r == "eu" {
		return EURegion
	}
	if r == "us" {
		return USRegion
	}
	return NoRegion
}
