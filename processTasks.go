package main

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/output/color"

	"github.com/newrelic/newrelic-diagnostics-cli/attach"
	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/output"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/suites"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func processTasksToRun() {
	log.Debugf("There are %d tasks in this queue\n", len(registration.Work.WorkQueue))

	if config.Flags.Tasks != "" {
		taskIdentifiers := processFlagsTasks(config.Flags.Tasks)
		registration.AddTasksByIdentifiers(taskIdentifiers)
	} else if config.Flags.Suites != "" {
		matchedSuites, err := processFlagsSuites(config.Flags.Suites, os.Args)
		if err != nil {
			log.Infof("\nError:\n%s", err.Error())
			os.Exit(1)
		}

		var suiteNameList []string
		for _, suite := range matchedSuites {
			suiteNameList = append(suiteNameList, suite.DisplayName)
		}
		log.Infof("%s %s\n", color.ColorString(color.White, "\nExecuting following diagnostic task suites:"), strings.Join(suiteNameList, ", "))

		taskIdentifiers := suites.DefaultSuiteManager.FindTasksBySuites(matchedSuites)
		registration.AddTasksByIdentifiers(taskIdentifiers)
	} else {
		registration.AddAllToQueue()
	}
	log.Debugf("There are %d tasks in this queue\n", len(registration.Work.WorkQueue))
	registration.CompleteTaskRegistration()
}

func processTasks(options tasks.Options, overrides []override, wg *sync.WaitGroup) {
	log.Debugf("work queue has %d items\n", len(registration.Work.WorkQueue))
	taskCount := 0
	for task := range registration.Work.WorkQueue {
		taskCount++
		if taskCount == 1 && !config.Flags.VeryQuiet {
			// writes to the screen
			output.WriteOutputHeader()
		}
		var taskOptions = make(map[string]string)
		// Loop through incoming options to assign out to the named task Options to avoid carrying in the wrong options
		for key, value := range options.Options {
			taskOptions[key] = value
		}
		namedTaskOptions := tasks.Options{Options: taskOptions}

		log.Debug("Running :", task.Identifier())
		log.Debug("Incoming options are", options)

		// Check for dependancies on the task and include results if dependent
		dependentResults := make(map[string]tasks.Result)
		for _, depIdent := range task.Dependencies() {
			log.Debug("dependency for processing: ", depIdent)
			dependentResults[depIdent] = registration.Work.Results[depIdent].Result
		}

		//Parse overrides to detect which task we are running
		for _, value := range overrides {
			// Initialize the taskOptions object
			log.Debugf("override %s: %s", value.Identifier, value.value)
			if strings.EqualFold(value.Identifier.String(), task.Identifier().String()) {
				log.Debug("Adding override to task namedTaskOptions", value.key, ":", value.value)
				namedTaskOptions.Options[value.key] = value.value
			}
		}

		log.Debug("Starting", task.Identifier(), "with options", namedTaskOptions)
		var result tasks.Result
		// Check for an option key to map to Status or Payload and if so, bypass task execution
		overrideEnabled := false
		if _, ok := namedTaskOptions.Options["Status"]; ok {
			log.Debug("Override Status passed in for ", task.Identifier(), "Value of ", namedTaskOptions.Options["Status"])

			switch status := strings.ToLower(namedTaskOptions.Options["Status"]); status {
			case "success":
				result.Status = tasks.Success
			case "warning":
				result.Status = tasks.Warning
			case "failure":
				result.Status = tasks.Failure
			case "info":
				result.Status = tasks.Info
			case "error":
				result.Status = tasks.Error
			case "none":
				result.Status = tasks.None
			default:
				log.Info("Attempted to set status override to invalid status", namedTaskOptions.Options["Status"])
			}

			result.Summary += "Status set by override to " + namedTaskOptions.Options["Status"] + "\n"
			overrideEnabled = true
		}

		if _, ok := namedTaskOptions.Options["Payload"]; ok {
			log.Debug("Override Payload passed in for ", task.Identifier())
			result.Payload = namedTaskOptions.Options["Payload"]
			result.Summary += "Payload set by override\n"
			overrideEnabled = true
		}

		if !overrideEnabled {
			result = task.Execute(namedTaskOptions, dependentResults)
		}

		taskResult := registration.TaskResult{
			Task:        task,
			Result:      result,
			WasOverride: overrideEnabled,
		}

		registration.Work.Results[task.Identifier().String()] = taskResult //This should be done in output.go but due to async causes issues
		registration.Work.ResultsChannel <- taskResult

		if len(result.FilesToCopy) > 0 {
			log.Debug(" - writing result to file channel")
			registration.Work.FilesChannel <- taskResult
		}

	} // end of loop to run tasks in

	log.Debug("Closing task channel")
	close(registration.Work.ResultsChannel)
	close(registration.Work.FilesChannel)

	log.Debug("Decrementing wait group in processTasks.")
	wg.Done()
}

func processFlagsTasks(flagValue string) []string {
	var validatedIdentifiers []string
	identifiers := strings.Split(flagValue, ",")
	for _, ident := range identifiers {
		if len(ident) > 0 { //This removes the adding of a blank identifier
			validatedIdentifiers = append(validatedIdentifiers, ident)
		}
	}
	return validatedIdentifiers
}

func getLicenseKey(thisResult tasks.Result) ([]string, error) {
	licenseKeyToSources, ok := thisResult.Payload.(map[string][]string)
	if !ok {
		return nil, fmt.Errorf("unable to retrieve license Key")
	}
	log.Debug("Valid License Key(s) provided")
	validLicenseKeys := []string{}
	for lk := range licenseKeyToSources {
		validLicenseKeys = append(validLicenseKeys, lk)
	}
	return validLicenseKeys, nil

}

func processFlagsSuites(flagValue string, args []string) ([]suites.Suite, error) {
	suiteIdentifiers := sanitizeAndParseFlagValue(flagValue)
	sanitizedArgs := sanitizeOSArgs(args)

	matchedSuites, unMatchedSuites := suites.DefaultSuiteManager.FindSuitesByIdentifiers(suiteIdentifiers)
	suites.DefaultSuiteManager.AddSelectedSuites(matchedSuites)
	//arguments passed that were intended to be suites but where not include due to misformat
	unknownArgs := suites.DefaultSuiteManager.CaptureOutOfPlaceArgs(sanitizedArgs, suiteIdentifiers)

	var errorMsg string

	if len(unMatchedSuites) > 0 {
		unMatchedSuitesList := "  \"" + strings.Join(unMatchedSuites, "\"\n  \"") + "\"\n"
		errorMsg = fmt.Sprintf("\n- Could not find the following task suites:\n\n%s", unMatchedSuitesList)
	}

	if len(unknownArgs) > 0 {
		unknownArgsList := "  \"" + strings.Join(unknownArgs, "\"\n  \"") + "\"\n"
		errorMsg += fmt.Sprintf("\n- You may have attempted to pass these arguments as suites:\n\n%v", unknownArgsList)
	}

	if len(errorMsg) > 1 {
		errorMsg += "\nPlease use the `--help suites` to check for available suites and proper formatting \n"
		return matchedSuites, fmt.Errorf(errorMsg)
	}

	return matchedSuites, nil

}

func processUploads() {
	log.Debug("processing uploads")

	//get timestamp to use attachment
	timestamp := time.Now().UTC().Format(time.RFC3339)

	if config.Flags.YesToAll {
		checkAttachmentFlags(timestamp)
		return
	}

	question := "We've created nrdiag-output.zip and nrdiag-output.json\n" +
		"Do you want to upload these to your New Relic account?"
	if promptUser(question) {
		checkAttachmentFlags(timestamp)
	}

}

func checkAttachmentFlags(timestamp string) {
	var ValidLicenseKeys []string
	attachDeps := new(attach.AttachDeps)

	if config.Flags.APIKey != "" {
		//hit DAS
		apiKey := config.Flags.APIKey
		attach.Upload("upload_api", apiKey, timestamp, attachDeps)
	} else if config.Flags.AutoAttach { //check for validated license keys and upload with those keys
		for _, taskResult := range registration.Work.Results {
			if taskResult.Task.Identifier().String() == "Base/Config/ValidateLicenseKey" && taskResult.Result.Status == tasks.Success {
				LicenseKeys, err := getLicenseKey(taskResult.Result)
				if err != nil {
					log.Debug("Could not retrieve a license key, automatic attachment will not be possible")
					return
				} else {
					ValidLicenseKeys = LicenseKeys
				}

			} else if taskResult.Task.Identifier().String() == "Base/Config/ValidateLicenseKey" && taskResult.Result.Status != tasks.Success {
				log.Info("No valid license keys specified, upload to New Relic Account cannot be completed")
				return
			}
		}
		for _, licenseKey := range ValidLicenseKeys {
			log.Info("Uploading files by Account ID...")
			if config.Flags.LegacyAttach {
				attach.UploadLegacy(licenseKey, timestamp)
			} else {
				attachDeps := new(attach.AttachDeps)
				attach.Upload("upload_s3", licenseKey, timestamp, attachDeps)
			}
		}
	}
}

func sanitizeOSArgs(osArgs []string) []string {
	var sanitizedArgs []string
	for _, osArg := range osArgs {
		trimmedArg := strings.TrimSpace(osArg)

		if trimmedArg == "" {
			continue
		}

		if strings.Contains(trimmedArg, ",") {
			splitArgs := strings.Split(trimmedArg, ",")
			sanitizedArgs = append(sanitizedArgs, splitArgs...)
			continue
		}

		sanitizedArgs = append(sanitizedArgs, trimmedArg)
	}
	return sanitizedArgs
}

func sanitizeAndParseFlagValue(flagValue string) []string {
	var sanitizedArgs []string
	parsedArgs := strings.Split(flagValue, ",")

	for _, arg := range parsedArgs {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg == "" {
			continue
		}
		sanitizedArgs = append(sanitizedArgs, trimmedArg)
	}

	return sanitizedArgs
}
