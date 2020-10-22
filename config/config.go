package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//Verbosity is the current log level
type Verbosity int

const (
	//Info indicates normal logging level
	Info Verbosity = iota
	//Verbose indicates additional logging (for troubleshooting the app)
	Verbose
)

//userFlags is a struct containing the commandline arguments passed in at runtime
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
	UsageOptOut        bool
	Proxy              string
	ProxyUser          string
	ProxyPassword      string
	Tasks              string
	AttachmentKey      string
	ConfigFile         string
	Override           string
	OutputPath         string
	Filter             string
	FileUpload         string
	BrowserURL         string
	AttachmentEndpoint string
	Suites             string
}

type ConfigFlag struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

//MarshalJSON - custom JSON marshaling for this task, we'll strip out the passphrase to keep it only in memory, not on disk
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
		ProxySpecified   bool
		SkipVersionCheck bool
		Tasks            string
		AttachmentKey    string
		ConfigFile       string
		Override         string
		OutputPath       string
		Filter           string
		BrowserURL       string
		Suites           string
	}{
		Verbose:          f.Verbose,
		Quiet:            f.Quiet,
		VeryQuiet:        f.VeryQuiet,
		YesToAll:         f.YesToAll,
		ShowOverrideHelp: f.ShowOverrideHelp,
		ProxySpecified:   proxySpecified,
		SkipVersionCheck: f.SkipVersionCheck,
		Tasks:            f.Tasks,
		AttachmentKey:    f.AttachmentKey,
		ConfigFile:       f.ConfigFile,
		Override:         f.Override,
		OutputPath:       f.OutputPath,
		Filter:           f.Filter,
		BrowserURL:       f.BrowserURL,
		Suites:           f.Suites,
	})
}

//LogLevel is the current log level for output to the screen
var LogLevel Verbosity

//Flags story configuration information
var Flags = userFlags{}

// Version of the application
var Version string

// UsageEndpoint is the haberdasher endpoint to which usage statistics are sent
var UsageEndpoint string

// AttachmentEndpoint is the haberdasher endpoint to which attachments are sent
var AttachmentEndpoint string

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

	flag.BoolVar(&Flags.Version, "version", false, "Display current program version. Take precedence over -no-version-check")
	flag.BoolVar(&Flags.SkipVersionCheck, "skip-version-check", false, "Skips the automatic check for a newer version of the application.")

	flag.StringVar(&Flags.Tasks, "t", defaultString, "alias for -tasks")
	flag.StringVar(&Flags.Tasks, "tasks", defaultString, "Specific {name of task} - could be comma separated list and/or contain a wildcard (*)")

	flag.StringVar(&Flags.Suites, "s", defaultString, "alias for -suites")
	flag.StringVar(&Flags.Suites, "suites", defaultString, "Specific {name of task suite} - could be comma separated list. If you do '-h suites' it will list all diagnostic task suites that can be run.")

	flag.StringVar(&Flags.AttachmentKey, "a", defaultString, "alias for -attachment-key")
	flag.StringVar(&Flags.AttachmentKey, "attachment-key", defaultString, "Attachment key for automatic upload to a support ticket (get key from an existing ticket).")

	flag.BoolVar(&Flags.Help, "h", false, "alias for -help")
	flag.BoolVar(&Flags.Help, "help", false, "Displays full list of command line options. If you do '-h tasks' it will list all tasks that can be run.")

	flag.StringVar(&Flags.ConfigFile, "c", defaultString, "alias for -config-file")
	flag.StringVar(&Flags.ConfigFile, "config-file", defaultString, "Override default config file location. Can be used to specify either a folder to search in addition to the default folders or a specific config file")

	flag.StringVar(&Flags.Proxy, "p", defaultString, "alias for -proxy")
	flag.StringVar(&Flags.Proxy, "proxy", defaultString, "Proxy should be in the format http(s)://proxyIp:proxyPort Not necessary in most cases… will override config file if used)")

	flag.StringVar(&Flags.ProxyUser, "proxy-user", defaultString, "Proxy username, if necessary")
	flag.StringVar(&Flags.ProxyPassword, "proxy-pw", defaultString, "Proxy pasword, if necessary")

	flag.StringVar(&Flags.Override, "o", defaultString, "alias for -override")
	flag.StringVar(&Flags.Override, "override", defaultString, "Specify overrides for detected values. Format <Identifier>.<property>=<value> - example '-o Base/Config/Validate.agentLanguage=PHP'")

	flag.StringVar(&Flags.OutputPath, "output-path", filepath.FromSlash("./"), "Output directory for results. Files will be named 'nrdiag-output.json and nrdiag-output.zip.")

	flag.BoolVar(&Flags.YesToAll, "y", false, "alias for -yes")
	flag.BoolVar(&Flags.YesToAll, "yes", false, "Say 'yes' to any prompt that comes up while running.")

	flag.StringVar(&Flags.Filter, "filter", "success,warning,failure,error,info", "Filter results based on status. Accepted values: Success, Warning, Failure, Error, None or Info. Multiple values can be provided in commma separated list. e.g: \"Success,Warning,Failure\"")

	flag.BoolVar(&Flags.Quiet, "q", false, "Quiet ouput; only prints the high level results and not the explainatory output. Suppresses file addition warnings if '-y' is also used. Does not contradict '-v'")
	flag.BoolVar(&Flags.VeryQuiet, "qq", false, "Very quiet ouput; only prints a single summary line for output (implies '-q'). Suppresses file addition warnings if '-y' is also used. Does not contradict '-v'. Inclusion filters are ignored.")

	flag.StringVar(&Flags.FileUpload, "file-upload", defaultString, "File to upload to support ticket, requires running with '-a' option")

	flag.StringVar(&Flags.BrowserURL, "browser-url", defaultString, "Specify a URL to check for the presence of a New Relic Browser agent")

	flag.BoolVar(&Flags.UsageOptOut, "usage-opt-out", false, "Decline to send anonymous New Relic Diagnostic tool usage data to New Relic for this run")

	//if first arg looks like it was build with `go build`, then we are testing against Haberdasher staging or localhost endpoint
	if strings.Contains(os.Args[0], "NrDiag") {
		flag.StringVar(&Flags.AttachmentEndpoint, "attachment-endpoint", defaultString, "The endpoint to send attachments to. (NR ONLY)")
	}

	flag.Parse()

	// Bail early if bad length attachment key provided.
	if Flags.AttachmentKey != "" && len(Flags.AttachmentKey) < 32 {
		fmt.Printf("Invalid attachment key '%s' length: %d\n", Flags.AttachmentKey, len(Flags.AttachmentKey))
		fmt.Println("The 32 character NR Diagnostics Attachment Key can be found upper-right of your ticket on support.newrelic.com")
		os.Exit(1)
	}

	if Flags.VeryQuiet {
		Flags.Quiet = true

		//pseudo-filter to force everything to display
		Flags.Filter = ""
	}

	//This has to be in the config init otherwise you don't get logs as expected
	if Flags.Verbose {
		LogLevel = Verbose
	} else {
		LogLevel = Info
	}

	if Flags.BrowserURL != "" {
		Flags.Override = "Browser/Agent/GetSource.url=" + Flags.BrowserURL + "," + Flags.Override
		Flags.Tasks = "Browser/Agent/Detect," + Flags.Tasks
	}
}

// UsagePayload gathers and sanitizes user command line input
// A map with string keys and interface values is returned
// The interface values will contain either a boolean or a string
func (f userFlags) UsagePayload() []ConfigFlag {
	return []ConfigFlag{
		ConfigFlag{Name: "verbose", Value: f.Verbose},
		ConfigFlag{Name: "interactive", Value: f.Interactive},
		ConfigFlag{Name: "quiet", Value: f.Quiet},
		ConfigFlag{Name: "veryQuiet", Value: f.VeryQuiet},
		ConfigFlag{Name: "help", Value: f.Help},
		ConfigFlag{Name: "version", Value: f.Version},
		ConfigFlag{Name: "yesToAll", Value: f.YesToAll},
		ConfigFlag{Name: "showOverrideHelp", Value: f.ShowOverrideHelp},
		ConfigFlag{Name: "proxy", Value: boolifyFlag(f.Proxy)},
		ConfigFlag{Name: "proxyUser", Value: boolifyFlag(f.ProxyUser)},
		ConfigFlag{Name: "proxyPassword", Value: boolifyFlag(f.ProxyPassword)},
		ConfigFlag{Name: "tasks", Value: f.Tasks},
		ConfigFlag{Name: "attachmentKey", Value: f.AttachmentKey},
		ConfigFlag{Name: "configFile", Value: boolifyFlag(f.ConfigFile)},
		ConfigFlag{Name: "override", Value: boolifyFlag(f.Override)},
		ConfigFlag{Name: "outputPath", Value: boolifyFlag(f.OutputPath)},
		ConfigFlag{Name: "filter", Value: f.Filter},
		ConfigFlag{Name: "fileUpload", Value: boolifyFlag(f.FileUpload)},
		ConfigFlag{Name: "browserURL", Value: boolifyFlag(f.BrowserURL)},
		ConfigFlag{Name: "attachmentEndpoint", Value: boolifyFlag(f.AttachmentEndpoint)},
		ConfigFlag{Name: "suites", Value: f.Suites},
	}
}

// boolifyFlag is a helper function for falsey/truthy conversion of UserFlag strings
func boolifyFlag(inputFlag string) bool {
	if inputFlag == "" {
		return false
	}
	return true
}

// IsForcedTask returns true if the supplied task (identifier) was supplied in the
// -t command line argument.
func (f userFlags) IsForcedTask(identifier string) bool {
	identifiers := strings.Split(f.Tasks, ",")
	for _, ident := range identifiers {
		trimmedIdentifer := strings.TrimSpace(ident)
		if strings.ToLower(identifier) == strings.ToLower(trimmedIdentifer) {
			return true
		}
	}
	return false
}
