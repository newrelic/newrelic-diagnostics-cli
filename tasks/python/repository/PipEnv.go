package repository

import (
	"fmt"
	"os/exec"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type PipEnv struct {
	CmdExec tasks.CmdExecFunc
}

func (p PipEnv) CheckPipVersion(pipCmd string) (result tasks.Result) {
	cmdBuild := exec.Command(pipCmd, "freeze")
	pipFreezeOutput, cmdBuildErr := cmdBuild.CombinedOutput()

	if cmdBuildErr != nil {
		return tasks.Result{
			Summary: "Unable to execute command: $ pip freeze. Error: " + cmdBuildErr.Error(),
			Status:  tasks.Error,
		}
	}

	if pipFreezeOutput == nil {
		return tasks.Result{
			Summary: "Collected pip freeze but output was empty",
			Status:  tasks.Error,
		}
	}

	// This will return the output of 'pip freeze' as a string, with packages separated by newlines
	pipFreezeOutputString := string(pipFreezeOutput)

	// stream payload to zip file to allow troubleshooting with the parsed config as needed
	stream := make(chan string)

	go tasks.StreamBlob(pipFreezeOutputString, stream)

	log.Debug("Result of running 'pip freeze': ")
	log.Debug(pipFreezeOutputString)

	pipFreezeOutputSlice := filterPipFreeze(pipFreezeOutputString)
	summary := fmt.Sprintf("Collected pip freeze using %s --freeze. See pipFreeze.txt for more info.", pipCmd)
	return tasks.Result{
		Summary:     summary, //where Info line prints from
		Status:      tasks.Info,
		Payload:     pipFreezeOutputSlice,
		FilesToCopy: []tasks.FileCopyEnvelope{{Path: "pipFreeze.txt", Stream: stream}},
	}
}

// this function will remove any warnings from pipFreezeOutput so only actionable data is left in the data structure.
func filterPipFreeze(pipFreezeOutput string) []string {
	var filteredOutput []string
	keys := make(map[string]bool)

	pipFreezeOutput = strings.ToLower(pipFreezeOutput)

	pipFreezeOutputSlice := strings.Fields(strings.ToLower(pipFreezeOutput))
	for _, v := range pipFreezeOutputSlice {
		if strings.Contains(v, "==") {
			if _, value := keys[v]; !value {
				keys[v] = true
				filteredOutput = append(filteredOutput, v)

			}
		}
	}

	return filteredOutput
}
