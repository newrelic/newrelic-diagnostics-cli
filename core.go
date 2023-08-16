package main

import (
	"os"
	"sync"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/internal/haberdasher"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/output"
	"github.com/newrelic/newrelic-diagnostics-cli/output/color"
	"github.com/newrelic/newrelic-diagnostics-cli/usage"
	"github.com/newrelic/newrelic-diagnostics-cli/version"
)

func main() {
	runID := generateRunID()
	config.ParseFlags()
	log.Debug("---------------------------------------------------------------------------------------------")
	log.Debugf("Running nrdiag with version: %s and build timestamp %s\n", config.Version, config.BuildTimestamp)
	log.Debugf("Run ID: %s\n", runID)
	log.Debug("nrdiag was run with options", os.Args)

	//Error setting proxy and they specifically included one so let's break out of the program before we attempt any non-proxied calls.
	_, err := processHTTPProxy()
	if err != nil {
		log.Info("Proxy configuration found, but unable to use. \nError: " + err.Error() + "\nExiting program.")
		os.Exit(3)
	}

	options, overrides := processOverrides()

	// Setup Haberdasher client
	haberdasher.InitializeDefaultClient()
	haberdasher.DefaultClient.SetRunID(runID)
	haberdasher.DefaultClient.SetUserAgent("Nrdiag_/" + config.Version)

	if config.HaberdasherURL == "" {
		log.Info("No Haberdasher base URL set. Defaulting to localhost")
	} else {
		haberdasher.DefaultClient.SetBaseURL(config.HaberdasherURL)
	}

	if config.Flags.Version && config.Flags.Quiet {
		// Support for automated version check by newrelic-cli
		current := version.ProcessAutoVersionCheck()
		if !current {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	go processTasksToRun()

	// if statements for doing stuff with args
	if config.Flags.Help {
		processHelp()
	} else if config.Flags.Version {
		version.ProcessVersion(promptUser)
	} else if config.Flags.Interactive {
		// do interactive stuff
	} else {
		// the wait group is way of tracking open threads
		// anytime you spawn an async function, increment and pass it in
		// ... the called function is responsible for decrementing when done
		var wg sync.WaitGroup

		// zip file is passed around as a dependency for other functions
		zipfile := output.CreateZip()

		// create the filelist file, exit if it can't be created
		flErr := output.CreateFileList()
		if flErr != nil {
			log.Info("Error creating filelist", err)
			os.Exit(3)
		}

		// check if any files/directories should be included in the zip
		if config.Flags.Include != "" {
			output.HandleIncludeFlag(zipfile, config.Flags.Include)
		}

		wg.Add(1) // run the tasks in goroutine
		go processTasks(options, overrides, &wg)

		wg.Add(1) // collect files the tasks produce and add them to the zip file
		go output.ProcessFilesChannel(zipfile, &wg)

		// this is a synchronous function that reads from the results channel
		// does not need the wait group since it blocks
		outputResults := output.WriteLineResults()

		if !config.Flags.Quiet {
			// writes to the screen
			output.WriteSummary(outputResults)
		}

		// block on wait group so program does not exit prematurely
		log.Info(color.ColorString(color.White, "Creating nrdiag-output.zip"))
		wg.Wait()

		// creates the output file
		output.WriteOutputFile(outputResults)

		// copy our output file(s) to the zip file
		output.CopyOutputToZip(zipfile)

		// copy the file list to the zip file last to ensure it's up to date
		output.CopyFileListToZip(zipfile)

		// ...and close it out
		output.CloseZip(zipfile)

		// upload any files (zip and json)
		processUploads()

		// deal with haberdasher data
		if !config.Flags.UsageOptOut {
			usage.SendUsageData(outputResults, runID)
		}
		if !config.Flags.SkipVersionCheck {
			version.ProcessAutoVersionCheck()
		}

		if config.Flags.Suites == "" && config.Flags.Tasks == "" {
			var command, option string
			if config.Flags.InNewRelicCLI {
				command = "newrelic diagnose run"
				option = "--list-suites"
			} else {
				command = os.Args[0]
				option = "-h suites"
			}
			log.Infof("\n\nFor better results, run Diagnostics CLI with the 'suites' option to target a New Relic product. To learn how to use this option, run: '%s %s'\n\n", command, option)
		}
	}
}
