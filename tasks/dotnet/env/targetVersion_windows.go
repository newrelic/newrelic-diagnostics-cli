package env

import (
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type DotNetEnvTargetVersion struct {
	osGetwd            tasks.OsFunc
	findFiles          func([]string, []string) []string
	returnStringInFile tasks.ReturnStringInFileFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p DotNetEnvTargetVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("DotNet/Env/TargetVersion")
}

func (p DotNetEnvTargetVersion) Explain() string {
	return "Determine framework version of .NET application"
}

// Dependencies - Returns the dependencies for each task.
func (p DotNetEnvTargetVersion) Dependencies() []string {
	return []string{
		"DotNet/Agent/Installed",
	}
}

// Execute - The core work within each task
func (p DotNetEnvTargetVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	var result tasks.Result
	// abort if it isn't installed
	if upstream["DotNet/Agent/Installed"].Status != tasks.Success {
		if upstream["DotNet/Agent/Installed"].Summary == tasks.NoAgentDetectedSummary {
			return tasks.Result{
				Status:  tasks.None,
				Summary: tasks.NoAgentUpstreamSummary + "DotNet/Agent/Installed",
			}
		}
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.UpstreamFailedSummary + "DotNet/Agent/Installed",
		}
	}

	workingDir, err := p.osGetwd()
	if err != nil {
		log.Debug("Error getting current working directory. ", err)
		result.Status = tasks.Error
		result.Summary = "Error getting current working directory."
		return result
	}

	configFiles := p.getNetConfigFiles(workingDir)
	if len(configFiles) < 1 {
		result.Status = tasks.Warning
		result.Summary = "Unable to find app config file. Are you running the " + tasks.ThisProgramFullName + " from your application's parent directory?"
		result.URL = "https://docs.newrelic.com/docs/agents/manage-apm-agents/troubleshooting/new-relic-diagnostics"
		return result
	}

	netVersion, countErrors := p.getNetVersionFromFiles(configFiles)

	if len(netVersion) < 1 {
		if countErrors < 1 {
			result.Status = tasks.None
			result.Summary = "Unable to find .NET version."
			return result
		} else {
			result.Status = tasks.Error
			result.Summary = "Error finding targetFramework"
			result.URL = "https://docs.newrelic.com/docs/agents/net-agent/getting-started/compatibility-requirements-net-framework-agent"
			return result
		}

	}
	result.Status = tasks.Info
	result.Summary = strings.Join(netVersion, ", ")
	result.Payload = netVersion
	return result
}

func (p DotNetEnvTargetVersion) getNetConfigFiles(workingDir string) []string {
	patterns := []string{`(.+).csproj$`, "^(?i)(web|app)[.]config$",
		"(?i).+[.]exe[.]config$", //  app.config files are almost always app-me.exe.config. filter NewRelicStatusMonitor.exe.config later
	}
	configs := p.findFiles(patterns, []string{workingDir})

	var configPaths []string
	for _, path := range configs {
		splitPaths := strings.SplitAfterN(path, "found file ", 1)
		if len(splitPaths) < 1 {
			continue
		}

		configPaths = append(configPaths, splitPaths[0])

	}
	return configPaths
}

func (p DotNetEnvTargetVersion) getNetVersionFromFiles(configPaths []string) ([]string, int) {
	countErrors := 0
	netVersion := []string{}
	var tmpNetVersion []string
	var err error
	for _, configFile := range configPaths {
		//Sample from Web.config: <httpRuntime targetFramework="4.7" />
		tmpNetVersion, err = p.returnStringInFile("httpRuntime targetFramework=\"([0-9.]+)", configFile)
		if err != nil {
			log.Debug("Error finding targetFramework", err)
			countErrors++
			continue
		}

		if len(tmpNetVersion) == 0 {
			//Sample from App.config: <supportedRuntime version="v4.0" sku=".NETFramework,Version=v4.6.1" />
			tmpNetVersion, err = p.returnStringInFile(".NETFramework,Version=v([0-9.]+)", configFile)
			if err != nil {
				log.Debug("Error finding targetFramework", err)
				countErrors++
				continue
			}
		}
		if len(tmpNetVersion) > 1 {
			ver := tmpNetVersion[1] //version from the capture group
			if len(ver) > 1 {
				netVersion = append(netVersion, ver)
			} else {
				log.Debug("Error parsing targetFramework version from value:", tmpNetVersion[0])
				countErrors++
			}
		} else {
			//Sample from a PRS.API.csproj: <TargetFramework>net5.0</TargetFramework>
			tmpNetVersion, err = p.returnStringInFile(`<TargetFramework>net([0-9.]+)<\/TargetFramework>`, configFile)
			if err != nil {
				log.Debug("Error finding targetFramework", err)
				countErrors++
				continue
			}
			if len(tmpNetVersion) > 1 {
				ver := tmpNetVersion[1] //version from the capture group 1
				if len(ver) > 1 {
					//remove word
					netVersion = append(netVersion, ver)
				} else {
					log.Debug("Error parsing targetFramework version from value:", tmpNetVersion[0])
					countErrors++
				}
			}
		}

	}
	return netVersion, countErrors

}
