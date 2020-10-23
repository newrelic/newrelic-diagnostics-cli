package log

import (
	"os"
	"regexp"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
	baseConfig "github.com/newrelic/newrelic-diagnostics-cli/tasks/base/config"
)

var logFilenamePatterns = []string{"newrelic_agent.*[.]log$",
	"newrelic-daemon[.]log$",
	"php_agent[.]log$",
	"newrelic-python-agent[.]log$",
	"NewRelic[.]Profiler.*[.]log$",
	"newrelic-infra.*[.]log$",
	"synthetics-minion[.]log$",
	"nrinstall-\\d{8}-\\d{6}-\\d{1,}[.]tar$",
}

var secureLogFilenamePatterns = []string{"docker[.]log$",
	"syslog$",
}

var logEnvVars = []string{
	"NRIA_LOG_FILE", // Infra agent
	"NEW_RELIC_LOG", //Java, Node and python agent paths
}

var keysInConfigFile = map[string][]string{
	"fullpaths": []string{
		"log_file",                //Python: "tmp/newrelic-python-agent.log"
		"newrelic.daemon.logfile", //PHP daemon default val: "/var/log/newrelic/newrelic-daemon.log"
		"newrelic.logfile",        //PHP: "/var/log/newrelic/newrelic-daemon.log",
		"logging.filepath",        // Node: "node/app/newrelic_agent.log"
	},
	"filenames": []string{
		"log_file_name", //Java and Ruby: "newrelic_agent.log"
		"-fileName",     //.NET: "FILENAME.log"
	},
	"directories": []string{
		"log_file_path", //Java, ruby: "/Users/shuayhuaca/Desktop/"
		"-directory",    //.NET "PATH\TO\LOG\DIRECTORY"
	},
	//"log.directory", //NET: how do we parse XML files?
}

var logPathsToIgnore = []string{
	"stdout",
	"stderr",
}

type LogElement struct {
	FileName string
	FilePath string
}

func collectFilePaths(envVars map[string]string, configElements []baseConfig.ValidateElement) ([]string, []string) {
	var paths []string
	localPath, err := os.Getwd()
	if err != nil {
		log.Info("Error reading local working directory")
	}
	paths = append(paths, localPath)

	if runtime.GOOS == "windows" {
		sysProgramFiles := envVars["ProgramFiles"]
		sysProgramData := envVars["ProgramData"]
		sysAppData := envVars["APPDATA"]

		paths = append(paths, sysProgramFiles+`\New Relic`)
		paths = append(paths, sysProgramData+`\New Relic\.NET Agent\Logs`)

		paths = append(paths, sysProgramFiles+`\New Relic\newrelic-infra\newrelic-infra.log`) //Windows, agent version 1.0.752 or lower
		paths = append(paths, sysProgramData+`\New Relic\newrelic-infra\newrelic-infra.log`)  //Windows, agent version 1.0.944 or higher

		//new infra logs (added if statment becasue I am unsure if the envvar will always point to the Roaming folder and not local or localLow )
		//Windows, agent version 1.0.775 to 1.0.944
		if strings.HasSuffix(sysAppData, "Roaming") {
			paths = append(paths, sysAppData+`\New Relic\newrelic-infra`)
		} else if strings.HasSuffix(sysAppData, "Local") {

			paths = append(paths, strings.TrimRight(sysAppData, "Local")+`Roaming\New Relic\newrelic-infra`)

		} else if strings.HasSuffix(sysAppData, "LocalLow") {
			paths = append(paths, strings.TrimRight(sysAppData, "LocalLow")+`Roaming\New Relic\newrelic-infra`)

		} else {
			paths = append(paths, sysAppData+`\Roaming\New Relic\newrelic-infra`)
		}

	} else {
		/*
			While the directories listed here will be walked, it is important to add any directory
			where a NR log file is expected, as only paths in this slice (no subdirectories)
			will be resolved from symbolic links. Matches will be deduped by tasks.FindFiles
		*/
		paths = append(paths, "/tmp")                                     //For Python Agent log and PHP installation log
		paths = append(paths, "/var/log")                                 //For Syn Minion and Infra
		paths = append(paths, "/var/log/newrelic")                        // For PHP agent and daemon log
		paths = append(paths, "/usr/local/newrelic-netcore20-agent/logs") // for dotnetcore
	}

	// Search for log files in standard locations
	//findFiles will return a full path that include filename
	fileLocations := tasks.FindFiles(logFilenamePatterns, paths)
	secureFileLocations := tasks.FindFiles(secureLogFilenamePatterns, paths)

	//logEnvVars contains OS-agnostic Environment variables of full log paths
	// These can be any filename and must be appended after FindFiles has constructed
	// the full URIs from known filenames and paths.

	// Any filepaths found in env vars will be automatically collected without prompting
	for _, logEnvVar := range logEnvVars {
		logPath := envVars[logEnvVar]

		if (len(logPath) > 0) &&
			!(tasks.PosString(fileLocations, logPath) > -1) && //make sure we are not adding a log file we already found
			!(tasks.PosString(secureFileLocations, logPath) > -1) {
			fileLocations = append(fileLocations, logPath)
		}
	}
	//this is our fallback because env vars and default paths should take precedence over config file
	if len(fileLocations) < 1 && len(secureFileLocations) < 1 {
		configFilePaths := getLogPathsFromConfigFile(configElements)
		fileLocations = append(fileLocations, configFilePaths...)
	}

	return fileLocations, secureFileLocations
}

func getLogPathsFromConfigFile(configElements []baseConfig.ValidateElement) []string {

	var filePaths []string

	for _, configFile := range configElements {
		var fullpath, filename, dir string
		//search for nr log keys that contain fullpath to the logs
		for _, v := range keysInConfigFile["fullpaths"] {
			foundKeys := configFile.ParsedResult.FindKey(v)
			for _, key := range foundKeys {
				val := key.Value()
				//"myapp/newrelic_agent.log"
				if len(val) > 0 {
					fullpath = val
					break //we should only grab one log path per config file
				}
			}
		}

		for _, v := range keysInConfigFile["filenames"] {
			foundKeys := configFile.ParsedResult.FindKey(v)
			for _, key := range foundKeys {
				val := key.Value()
				if len(val) > 0 {
					filename = val
					break //we should only grab one log path per config file
				}
			}
		}

		for _, v := range keysInConfigFile["directories"] {
			foundKeys := configFile.ParsedResult.FindKey(v)
			for _, key := range foundKeys {
				val := key.Value()
				if len(val) > 0 {
					dir = val
					break //we should only grab one log path per config file
				}
			}
		}

		if len(fullpath) > 0 {
			filePaths = append(filePaths, fullpath)
		}
		if len(dir) > 0 {
			if len(filename) > 0 {
				filePaths = append(filePaths, dir+filename)
			} else {
				logFiles := findLogFiles(logFilenamePatterns, dir)
				filePaths = append(filePaths, logFiles...)
			}
		}
	}

	return filePaths
}

func findLogFiles(patterns []string, dir string) []string {
	pathsToFiles := getFilesFromDir(dir)

	// map to automatically dedupe file matches
	foundLogFiles := make(map[string]struct{})
	for _, file := range pathsToFiles {
		// loop through pattern list and add files that match to our string array
		for _, pattern := range patterns {
			regex := regexp.MustCompile(pattern)
			if regex.MatchString(file) {
				foundLogFiles[file] = struct{}{} // empty struct is smallest memory footprint
			}
		}
	}
	var uniqueFoundFiles []string
	for fileLocation := range foundLogFiles {
		uniqueFoundFiles = append(uniqueFoundFiles, fileLocation)
	}
	return uniqueFoundFiles
}

func getFilesFromDir(dir string) []string {
	var potentialLogFiles []string
	f, err := os.Open(dir)
	if err != nil {
		log.Debug(err)
	}
	files, err := f.Readdir(-1)
	f.Close()
	if err != nil {
		log.Debug(err)
	}

	for _, file := range files {
		if !file.IsDir() && !(strings.HasPrefix(file.Name(), ".")) {
			potentialLogFiles = append(potentialLogFiles, dir+file.Name())
		}
	}
	return potentialLogFiles
}
