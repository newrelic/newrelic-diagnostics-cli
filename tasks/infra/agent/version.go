package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/helpers/httpHelper"
	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

var (
	githubAPIReleaseURL   = "https://api.github.com/repos/newrelic/infrastructure-agent/releases/tags/"
	oldestGithubRelease   = tasks.Ver{Major: 1, Minor: 12, Patch: 0, Build: 0}
	errUnsupportedVersion = fmt.Errorf("New Relic Infrastructure Agent version unsupported, installed agent was released more than two years ago")
	errRecommendedUpgrade = fmt.Errorf("New Relic Infrastructure Agent upgrade is recommended, installed agent was released more than one year ago")
)

type githubReleaseData struct {
	PublishedAt string `json:"published_at"`
}

// InfraAgentVersion - This struct defines the Infrastructure agent version task
type InfraAgentVersion struct {
	runtimeOS   string
	httpGetter  requestFunc
	now         func() time.Time
	cmdExecutor func(name string, arg ...string) ([]byte, error)
}

// yearsBetween returns the years between two given dates
func yearsBetween(firstDate time.Time, secondDate time.Time) int64 {
	return int64((firstDate.Sub(secondDate).Hours()) / 24 / 365)
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

	if upstream["Infra/Config/Agent"].Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "Task did not meet requirements necessary to run: Infrastructure Agent not detected on system",
		}
	}

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
			Summary: fmt.Sprintf("New Relic Infrastructure Agent version could not be determined because Diagnostics CLI encountered this issue when running the command 'newrelic-infra -version': %s", err.Error()),
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

	err = p.validatePublishDate(ver)
	if err != nil {
		urlUpdateTask := tasks.Result{
			URL:     "https://docs.newrelic.com/docs/infrastructure/new-relic-infrastructure/installation/update-infrastructure-agent",
			Summary: err.Error(),
			Status:  tasks.Error,
		}

		// change task status depending on error
		if errors.Is(err, errUnsupportedVersion) {
			urlUpdateTask.Status = tasks.Failure
		} else if errors.Is(err, errRecommendedUpgrade) {
			urlUpdateTask.Status = tasks.Warning
		}

		return urlUpdateTask
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
	//binary path in Windows: C:\Program Files\New Relic\newrelic-infra\newrelic-infra.exe
	if p.runtimeOS == "windows" {
		sysProgramFiles, ok := envVars["ProgramFiles"]
		if !ok {
			return "", errors.New("environment variable not set: ProgramFiles")
		}
		binaryPath = sysProgramFiles + `\New Relic\newrelic-infra\newrelic-infra.exe`
	} else {
		binaryPath = "newrelic-infra"
	}

	return binaryPath, nil
}

// validatePublishDate returns an error if the version was released more than one year ago
func (p InfraAgentVersion) validatePublishDate(version tasks.Ver) error {
	// Github releases started on version 1.12.0
	if version.IsLessThanEq(oldestGithubRelease) {
		return errUnsupportedVersion
	}

	publishData, err := p.getGithubPublishDate(fmt.Sprintf("%d.%d.%d", version.Major, version.Minor, version.Patch))
	if err != nil {
		return fmt.Errorf("Unable to get New Relic Infrastructure Agent release date: %w", err)
	}

	years := yearsBetween(p.now(), publishData)
	if years >= 2 {
		return errUnsupportedVersion
	} else if years == 1 {
		return errRecommendedUpgrade
	}

	return nil
}

func (p InfraAgentVersion) getGithubPublishDate(version string) (time.Time, error) {
	wrapper := httpHelper.RequestWrapper{
		Method:         "GET",
		URL:            githubAPIReleaseURL + version,
		TimeoutSeconds: 30,
	}

	response, err := p.httpGetter(wrapper)
	if err != nil {
		return time.Time{}, err
	}
	defer response.Body.Close()

	var apiData githubReleaseData

	err = json.NewDecoder(response.Body).Decode(&apiData)
	if err != nil {
		return time.Time{}, err
	}

	return time.Parse(time.RFC3339, apiData.PublishedAt)
}
