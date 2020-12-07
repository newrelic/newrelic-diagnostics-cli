package log

import (
	"fmt"
	"os"
	"path/filepath"
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

var logSysProp = "-Dnewrelic.logfile" //EX: Dnewrelic.logfile=/opt/newrelic/java/logs/newrelic/somenewnameformylogs.log

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

type LogElement struct {
	FileName           string
	FilePath           string
	Source             string
	Value              string
	IsSecureLocation   bool
	CanCollect         bool
	ReasonToNotCollect string
}

var (
	defaultSource    = "log found by looking at New Relic default paths"
	envVarSource     = "log found through New Relic environment variable"
	sysPropSource    = "log found by looking at JVM arguments"
	configFileSource = "log found through values set in New Relic config file"
)

func collectFilePaths(envVars map[string]string, configElements []baseConfig.ValidateElement, foundSysPropPath string) []LogElement {
	var paths []string
	currentPath, err := os.Getwd()
	if err != nil {
		log.Info("Error reading local working directory")
	}
	paths = append(paths, currentPath)

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

	var logFilesFound []LogElement

	// Search for log files in standard locations
	//findFiles will return a full path that include filename
	fileLocations := tasks.FindFiles(logFilenamePatterns, paths)
	if len(fileLocations) > 0 {
		for _, fileLocation := range fileLocations {
			dir, fileName := filepath.Split(fileLocation)
			logFilesFound = append(logFilesFound, LogElement{
				FileName:         fileName,
				FilePath:         dir,
				Source:           defaultSource,
				Value:            fileLocation,
				IsSecureLocation: false,
				CanCollect:       true,
			})
		}
	}
	secureFileLocations := tasks.FindFiles(secureLogFilenamePatterns, paths)
	if len(secureFileLocations) > 0 {
		for _, fileLocation := range secureFileLocations {
			dir, fileName := filepath.Split(fileLocation)
			logFilesFound = append(logFilesFound, LogElement{
				FileName:         fileName,
				FilePath:         dir,
				Source:           defaultSource,
				Value:            fileLocation,
				IsSecureLocation: true,
				CanCollect:       true,
			})
		}
	}

	//logEnvVars contains OS-agnostic Environment variables (full log paths or only log filename) Any filepaths found in env vars will be automatically collected without prompting
	for _, logEnvVar := range logEnvVars {
		logPath, isPresent := envVars[logEnvVar]
		if isPresent &&
			!(tasks.PosString(fileLocations, logPath) > -1) && //make sure we are not adding a log file we already found
			!(tasks.PosString(secureFileLocations, logPath) > -1) {

			if logPath == "stdout" || logPath == "stderr" {
				logFilesFound = append(logFilesFound, LogElement{
					Source:             logEnvVar,
					Value:              logPath,
					IsSecureLocation:   false,
					CanCollect:         false,
					ReasonToNotCollect: tasks.ThisProgramFullName + " cannot collect logs that have been set to STDOUT OR STDERR",
				})
				continue
			}

			dir, fileName := filepath.Split(logPath)
			if len(dir) > 0 { //path is a directory or a fullpath that includes filename
				logFilesFound = append(logFilesFound, LogElement{
					FileName:         fileName,
					FilePath:         dir,
					Source:           logEnvVar,
					Value:            logPath,
					IsSecureLocation: false,
					CanCollect:       true,
				})
				continue
			}
			//path is only a filename, let's attempt to find it in the current directory (and not in all paths because we could find a file with that name for another NR product and falsely say that we collect the right log)
			logsFullPath := tasks.FindFiles([]string{fileName}, []string{currentPath})
			if len(logsFullPath) > 0 {
				for _, logFullPath := range logsFullPath {
					dir, logName := filepath.Split(logFullPath)

					logFilesFound = append(logFilesFound, LogElement{
						FileName:         logName,
						FilePath:         dir,
						Source:           logEnvVar,
						Value:            logFullPath,
						IsSecureLocation: false,
						CanCollect:       true,
					})
				}
				continue
			}
			//unable to find it
			logFilesFound = append(logFilesFound, LogElement{
				FileName:           fileName,
				FilePath:           currentPath,
				Source:             logEnvVar,
				Value:              logPath,
				IsSecureLocation:   false,
				CanCollect:         false,
				ReasonToNotCollect: fmt.Sprintf(tasks.ThisProgramFullName+" is unable to find a file named %s under the current directory where it's being run.", logPath),
			})
		}
	}

	//check for system properties
	if len(foundSysPropPath) > 0 {
		//make sure we are not adding a log file we already found
		if !(tasks.PosString(fileLocations, foundSysPropPath) > -1) && !(tasks.PosString(secureFileLocations, foundSysPropPath) > -1) {
			dir, fileName := filepath.Split(foundSysPropPath)

			logFilesFound = append(logFilesFound, LogElement{
				FileName:         fileName,
				FilePath:         dir,
				Source:           logSysProp,
				Value:            foundSysPropPath,
				IsSecureLocation: false,
				CanCollect:       true,
			})
		}

	}

	configFilePaths := getLogPathsFromConfigFile(configElements)

	if len(configFilePaths) > 0 {
		for _, configFilePath := range configFilePaths {
			dir, fileName := filepath.Split(configFilePath)

			logFilesFound = append(logFilesFound, LogElement{
				FileName:         fileName,
				FilePath:         dir,
				Source:           configFileSource,
				Value:            configFilePath,
				IsSecureLocation: false,
				CanCollect:       true,
			})
		}
	}
	return logFilesFound
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
