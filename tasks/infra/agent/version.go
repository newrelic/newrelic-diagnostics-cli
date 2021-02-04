package agent

import (
	"fmt"
	"regexp"
	"errors"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

// InfraAgentVersion - This struct defines the Infrastructure agent version task
type InfraAgentVersion struct {
	runtimeOS   string
	cmdExecutor func(name string, arg ...string) ([]byte, error)
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p InfraAgentVersion) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Infra/Agent/Version")
}

// Explain - Returns the help text for each individual task
func (p InfraAgentVersion) Explain() string {
	return "Determine version of New Relic Infrastructure agent"
}

// Dependencies - Returns the dependencies for ech task. When executed by name each dependency will be executed as well and the results from that dependency passed in to the downstream task
func (p InfraAgentVersion) Dependencies() []string {
	return []string{
		"Infra/Config/Agent",
		"Base/Env/CollectEnvVars",
	}
}

// Execute - The core work within each task
func (p InfraAgentVersion) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result { //By default this task is commented out. To see it run go to the tasks/registerTasks.go file and uncomment the w.Register for this task
	if upstream["Base/Env/CollectEnvVars"].Status != tasks.Info {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: Upstream dependency failed",
		}
	}

	envVars, ok := upstream["Base/Env/CollectEnvVars"].Payload.(map[string]string)
	if !ok {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: tasks.AssertionErrorSummary,
		}
	}

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: Infrastructure Agent not detected on system",
		}
	}

	binaryPath, err := p.getBinaryPath(envVars)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to determine New Relic Infrastructure binary path: %s", err.Error()),
		}
	}
	
	log.Debug("Binary Path found was ", binaryPath)

	rawVersionOutput, err := p.getInfraVersion(binaryPath)
	if err != nil {
		return tasks.Result{
			Status:  tasks.Failure,
			Summary: fmt.Sprintf("New Relic Infrastructure Agent version could not be determined: %s", err.Error()),
			URL:     "https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/update-infrastructure-agent",
		}
	}

	//$ newrelic-infra -version
	//New Relic Infrastructure Agent version: 1.5.40
	versionRegex := regexp.MustCompile(": ([0-9.]+)") //This pulls the numeric version from the string returned
	matches := versionRegex.FindStringSubmatch(rawVersionOutput)

	if len(matches) < 2 {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to parse New Relic Infrastructure Agent version from: %s", rawVersionOutput),
		}
	}

	ver, err := tasks.ParseVersion(matches[1])
	if err != nil {
		return tasks.Result{
			Status:  tasks.Error,
			Summary: fmt.Sprintf("Unable to parse New Relic Infrastructure Agent version from: %s", rawVersionOutput),
		}
	}

	return tasks.Result{
		Status:  tasks.Info,
		Summary: matches[1],
		Payload: ver,
	}
}

func (p InfraAgentVersion) getInfraVersion(binaryPath string) (string, error) {

	version, cmdBuildErr := p.cmdExecutor(binaryPath, "-version")
	if cmdBuildErr != nil {
		log.Debug("Error running ", binaryPath, "-version:", cmdBuildErr)
		log.Debug("Output was ", string(version))
		return string(version), cmdBuildErr
	}
	return string(version), nil
}

func (p InfraAgentVersion) getBinaryPath(envVars map[string]string) (string, error) {
	var binaryPath string

	if p.runtimeOS == "windows" {
		sysProgramFiles, ok := envVars["ProgramFiles"]
		if !ok {
			return "", errors.New("ProgramFiles environment variable not set")
		}
		binaryPath = sysProgramFiles + `\New Relic\newrelic-infra\newrelic-infra.exe`
	} else {
		binaryPath = "newrelic-infra"
	}

	return binaryPath, nil
}
