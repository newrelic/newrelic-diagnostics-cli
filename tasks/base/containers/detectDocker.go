package containers

import (
	"bufio"
	"bytes"
	"fmt"

	"github.com/newrelic/NrDiag/tasks"
)

// BaseContainersDetectDocker - This struct defined tests availability of docker
type BaseContainersDetectDocker struct {
	executeCommand tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (t BaseContainersDetectDocker) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Base/Containers/DetectDocker")
}

// Explain - Returns the help text for each individual task
func (t BaseContainersDetectDocker) Explain() string {
	return "Detect Docker Daemon"
}

// Dependencies - Returns the dependencies for each task.
func (t BaseContainersDetectDocker) Dependencies() []string {
	return []string{}
}

// Execute - The core work within each task
func (t BaseContainersDetectDocker) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {
	dockerInfoCLIBytes, infoBytesErr := tasks.GetDockerInfoCLIBytes(t.executeCommand)

	if infoBytesErr != nil {
		return tasks.Result{
			Status:  tasks.None,
			Summary: fmt.Sprintf("Error retrieving Docker Daemon info: %s - %s", infoBytesErr.Error(), dockerInfoCLIBytes),
			Payload: tasks.DockerInfo{},
		}
	}

	//Do nothing with error here since this func will always return original bytes if error
	prettyJSONBytes, _ := tasks.BytesToPrettyJSONBytes(dockerInfoCLIBytes)

	stream := make(chan string)
	go streamDockerInfo(prettyJSONBytes, stream)

	filesToCopy := []tasks.FileCopyEnvelope{
		tasks.FileCopyEnvelope{
			Path:       "docker-info.json",
			Stream:     stream,
			Identifier: t.Identifier().String(),
		},
	}

	dockerInfo, parseErr := tasks.NewDockerInfoFromBytes(dockerInfoCLIBytes)

	if parseErr != nil {
		return tasks.Result{
			Status:       tasks.None,
			Summary:     "Error parsing JSON docker info " + parseErr.Error(),
			Payload:     dockerInfo,
			FilesToCopy: filesToCopy,
		}
	}

	if dockerInfo.ServerVersion != "" {
		return tasks.Result{
			Status:       tasks.Info,
			Summary:     "Docker Daemon is Running",
			Payload:     dockerInfo,
			FilesToCopy: filesToCopy,
		}
	}

	return tasks.Result{
		Status:      tasks.None,
		Summary:     "Can't determine if Docker Daemon is running on host: unexpected output",
		Payload:     dockerInfo,
		FilesToCopy: filesToCopy,
	}

}


func streamDockerInfo(dockerInfo []byte, ch chan string) {
	defer close(ch)

	scanner := bufio.NewScanner(bytes.NewReader(dockerInfo))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}
}

