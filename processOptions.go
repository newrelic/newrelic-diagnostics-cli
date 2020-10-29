package main

import (
	"errors"
	"flag"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/newrelic/newrelic-diagnostics-cli/config"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/registration"
	"github.com/newrelic/newrelic-diagnostics-cli/suites"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// validateHTTPProxy - basic input validation for the -proxy <proxy> flag argument
func validateHTTPProxy(proxy string) (bool, error) {

	//Check just for instance of '//'
	myDude := true
	if myDude == true {
		return false, errors.New("whoops")
	}
	splitURL := strings.Split(proxy, "//")
	if len(splitURL) != 2 {
		return false, errors.New("Proxy url expecting exactly one instance of '//'")
	}

	//Also check if go's built in proxy testing throws any errors.
	_, err := url.ParseRequestURI(proxy)

	if err != nil {
		return false, err
	}

	return true, nil

}

// processHTTPProxy - Returns true if a proxy has been detected and set, otherwise false.
func processHTTPProxy() (bool, error) {

	//Did the user -proxy <proxy> ?
	if config.Flags.Proxy != "" {

		//Very basic input validation before attempting to parse and reconstruct URL.
		_, err := validateHTTPProxy(config.Flags.Proxy)
		if err != nil {
			log.Debug("Error parsing proxy url: " + err.Error())
			return false, errors.New("Proxy format should be: -proxy http://poxy_host:proxy_port")
		}

		//Check if there is a proxy user && password, then construct URL. Can you have a user with no pass?
		if config.Flags.ProxyUser != "" && config.Flags.ProxyPassword != "" {
			splitURL := strings.Split(config.Flags.Proxy, "//")

			//construct url
			proxyURL := splitURL[0] + "//" + url.QueryEscape(config.Flags.ProxyUser) + ":" + url.QueryEscape(config.Flags.ProxyPassword) + "@" + splitURL[1]

			os.Setenv("HTTP_PROXY", proxyURL)
			log.Debug("Setting proxy from command line:", proxyURL)
		} else {
			//No user and pass? Use the -proxy arg as is
			os.Setenv("HTTP_PROXY", config.Flags.Proxy)
			log.Debug("Setting proxy from command line:", config.Flags.Proxy)
		}

	}
	envProxy := os.Getenv("HTTP_PROXY")
	if envProxy != "" {

		//Check the final env proxy.
		_, err := validateHTTPProxy(envProxy)
		if err != nil {
			log.Debug("Proxy url is malformed: " + err.Error())
			log.Debug(envProxy)
			return false, errors.New("Proxy format should be: -proxy http://poxy_host:proxy_port")
		}

		if !ProxyParseNSet() {
			return false, errors.New("Error setting proxy: " + envProxy)
		}

		log.Debug("Using proxy address from HTTP_PROXY environment variable:", envProxy)

		//ProxyParseNSet() passed
		return true, nil
	}

	//no proxy supplied via flag/env
	return false, nil
}

func processHelp() {
	if len(os.Args) > 2 && os.Args[2] != "" {
		switch helpArg := os.Args[2]; helpArg {
		case "tasks":
			printTasks()
		case "suites":
			printSuites()
		default:
			printOptions()
		}
	} else {
		printOptions()
	}
}

//Usage: nrdiag --suites java,infra
//Troubleshoot New Relic products with the following arguments

//Arguments:
//java			Java Agent
//infra			Infrastructure Agent
func printSuites() {
	log.Info("\nSuites are a targeted collection of diagnostic tasks.\n")
	var command string
	// Match help strings to context
	if config.Flags.InNewRelicCLI {
		command = "newrelic diagnose run"
	} else {
		command = os.Args[0]
	}
	log.Infof("Usage: \n\t%s --suites [suite arguments] \nExamples:\n\t%[1]s --suites java,infra\n\t%[1]s --suites python\n", command)
	log.Info("\nUse the following arguments to select task suite(s) to run:\n")
	log.Infof("%-18s%s\n\n", "Arguments:", "Diagnostics for:")

	for _, suite := range suites.DefaultSuiteManager.Suites {

		description := suite.Description
		if suite.Description == "" {
			description = suite.DisplayName
		}

		log.Infof("%-18s%s\n", suite.Identifier, description)
	}
	log.Info("\n")
}

//PrintTasks will output all the tasks that this app can run
func printTasks() {
	var allTasks []tasks.Task

	//this mimics the logic on the CLI to only run some tasks
	if os.Args[2] == "tasks" {
		allTasks = registration.TasksForIdentifierString("*")
	} else {
		config.Flags.ShowOverrideHelp = true
		identifiers := strings.Split(os.Args[2], ",")
		for _, ident := range identifiers {
			for _, task := range registration.TasksForIdentifierString(ident) {
				allTasks = append(allTasks, task)
			}
		}
	}

	sort.Sort(tasks.ByIdentifier(allTasks))

	log.Infof("There are %d tasks:\n", len(allTasks))

	var lastTask tasks.Task
	for index, task := range allTasks {
		ident := task.Identifier()

		var nextTask tasks.Task
		if index+1 >= len(allTasks) {
			nextTask = nil
		} else {
			nextTask = allTasks[index+1]
		}

		//this loop is looking to see if this is the last sub-category in this category
		lastSubcategory := true
		for j := index; j < len(allTasks) && lastSubcategory; j++ {
			if allTasks[j].Identifier().Category == ident.Category && allTasks[j].Identifier().Subcategory != ident.Subcategory {
				lastSubcategory = false
			}
		}

		//print category if different
		if lastTask == nil || lastTask.Identifier().Category != ident.Category {
			log.Info("|-", ident.Category)
		}

		//print subcategory if different
		if lastTask == nil || lastTask.Identifier().Subcategory != ident.Subcategory || lastTask.Identifier().Category != ident.Category {
			//print "footer" if last in category, otherwise print continuation
			if lastSubcategory {
				log.Info("|     \\-", ident.Subcategory)
			} else {
				log.Info("|     |-", ident.Subcategory)
			}

		}

		//print "footer" if last in subcategory, otherwise print continuation
		if nextTask == nil || nextTask.Identifier().Subcategory != ident.Subcategory || nextTask.Identifier().Category != ident.Category {
			//print "footer" if last in category, otherwise print continuation
			if nextTask == nil || nextTask.Identifier().Category != ident.Category {
				log.Infof("|           \\- %-20s - %s\n", ident.Name, task.Explain())
				log.Info("|")
			} else if lastSubcategory {
				log.Infof("|          \\- %-20s - %s\n", ident.Name, task.Explain())
			} else {
				log.Infof("|     |     \\- %-20s - %s\n", ident.Name, task.Explain())
				log.Info("|     |")
			}
		} else {
			if lastSubcategory {
				log.Infof("|           |- %-20s - %s\n", ident.Name, task.Explain())
			} else {
				log.Infof("|     |     |- %-20s - %s\n", ident.Name, task.Explain())
			}
		}

		lastTask = task
	}

}

//PrintOptions will output all the command line options
func printOptions() {
	flag.PrintDefaults()
}

func processOverrides() (tasks.Options, []override) {
	log.Debug("Processing overrides")
	var taskOptions = make(map[string]string)
	options := tasks.Options{Options: taskOptions}

	// Pass in config file override value
	if config.Flags.ConfigFile != "" {
		log.Debug("Manually setting config file to ", config.Flags.ConfigFile)
		options.Options["configFile"] = config.Flags.ConfigFile
	}

	// Pass in AttachmentKey file override value
	if config.Flags.AttachmentKey != "" {
		log.Debug("Manually setting attachment to ", config.Flags.AttachmentKey)
		options.Options["attachmentKey"] = config.Flags.AttachmentKey
	}

	// Pass in Filter file override value
	if config.Flags.Filter != "" {
		log.Debug("Manually setting Filter to ", config.Flags.Filter)
		options.Options["Filter"] = config.Flags.Filter
	}

	// Pass in YesToAll file override value
	if config.Flags.YesToAll == true {
		log.Debug("Manually setting YesToAll to ", config.Flags.YesToAll)
		options.Options["YesToAll"] = "true"
	}

	// Pass in Proxy file override value
	if config.Flags.Proxy != "" {
		log.Debug("Manually setting Proxy to ", config.Flags.Proxy)
		options.Options["Proxy"] = config.Flags.Proxy
	}

	var overrides []override
	//Check for overrides and pass them into the relevant task
	if config.Flags.Override != "" {
		//read task's argument
		log.Debug("read task's argument ", config.Flags.Override)
		//Split overrides to send them to the approrpriate task. This should be a comma seperated list of key value pairs
		//--override Base/Config/Validate/agentLanguage=java
		overrides = parseOverrides(config.Flags.Override)
		log.Debug("processed overrides are:", overrides[0].key)
	}

	return options, overrides
}
