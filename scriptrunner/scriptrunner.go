package scriptrunner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/newrelic/newrelic-diagnostics-cli/config"
	"github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type IRunnerDependencies interface {
	ContinueIfExists(savepath string) bool
	SaveToDisk(body []byte, savepath string) error
	RunScript(body []byte, savepath string, scriptOptions string) ([]byte, error)
	GetUUID() string
}

type Runner struct {
	Deps IRunnerDependencies
}

type RunnerDependencies struct {
	CmdLineOptions tasks.Options
}

type ScriptData struct {
	Name               string
	Path               string
	Flags              string
	Description        string
	Content            []byte
	AddtlFilesPatterns []string
	AddtlFiles         []string
	Output             []byte
	OutputPath         string
}

func (sr *Runner) Run(body []byte, savepath string, scriptOptions string) ([]byte, error) {
	savepathWithId := sr.addUUIDToFilename(savepath)
	return sr.Deps.RunScript(body, savepathWithId, scriptOptions)
}

func (sr *Runner) FindScriptAddtlFiles(filePatterns []string) []string {
	if len(filePatterns) < 1 {
		return []string{}
	}
	realPaths := []string{}
	for _, p := range filePatterns {
		matches, err := filepath.Glob(p)
		if err != nil {
			logger.Infof("Error collecting script file by pattern: %s\n%s", p, err.Error())
			continue
		}
		realPaths = append(realPaths, matches...)

	}
	return realPaths
}

func (r *RunnerDependencies) RunScript(body []byte, savepathWithId string, scriptOptions string) ([]byte, error) {
	err := r.SaveToDisk(body, savepathWithId)
	if err != nil {
		return nil, err
	}
	absPath, err := filepath.Abs(savepathWithId)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(absPath, scriptOptions)
	cmd.Dir = config.Flags.OutputPath
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	err = os.Remove(savepathWithId)
	if err != nil {
		return nil, err
	}
	return stdout, nil
}

func (r *RunnerDependencies) SaveToDisk(body []byte, savepath string) error {
	if !r.ContinueIfExists(savepath) {
		return os.ErrExist
	}
	return os.WriteFile(savepath, body, 0700)
}

func (r *RunnerDependencies) ContinueIfExists(savepath string) bool {
	if tasks.FileExists(savepath) {
		logger.Infof("File already exists: %s\n", savepath)
		return tasks.PromptUser("Would you like to overwrite it?", r.CmdLineOptions)
	}
	return true
}

func (r *RunnerDependencies) GetUUID() string {
	return uuid.New().String()
}

func (sr *Runner) addUUIDToFilename(savepath string) string {
	runid := sr.Deps.GetUUID()
	dir, fullFilename := filepath.Split(savepath)
	fileExt := filepath.Ext(fullFilename)
	filenameNoExt := strings.TrimSuffix(fullFilename, fileExt)
	newFilename := filenameNoExt + "-" + runid + fileExt
	return filepath.Join(dir, newFilename)
}
