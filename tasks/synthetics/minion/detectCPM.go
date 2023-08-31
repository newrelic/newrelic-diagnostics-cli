package minion

import (
	"bufio"
	"bytes"
	"encoding/json"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

//

// SyntheticsMinionDetectCPM - This struct defined the sample plugin which can be used as a starting point
type SyntheticsMinionDetectCPM struct {
	executeCommand tasks.CmdExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p SyntheticsMinionDetectCPM) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Synthetics/Minion/DetectCPM")
}

// Explain - Returns the help text for each individual task
func (p SyntheticsMinionDetectCPM) Explain() string {
	return "Collect docker inspect of New Relic Synthetics containerized private minions"
}

// Dependencies - Returns the dependencies for each task.
func (p SyntheticsMinionDetectCPM) Dependencies() []string {
	return []string{"Base/Containers/DetectDocker"}
}

// Execute - The core work within each task
func (p SyntheticsMinionDetectCPM) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	if upstream["Base/Containers/DetectDocker"].Status != tasks.Info {
		result := tasks.Result{
			Status:  tasks.None,
			Summary: "Docker Daemon not available to detect CPM",
		}
		return result
	}

	//Query docker for last 4 CPMs active or exited by label 'name' with expected value of 'synthetics-minion'
	//Note if customer wraps CPM in their own image or re-names the image label 'name' it wont be detected
	//but otherwise labels are inherited from base images
	containerIds, err := tasks.GetContainerIdsByLabel("name", "synthetics-minion", 4, true, p.executeCommand)

	if err != nil {
		result := tasks.Result{
			Status:  tasks.Error,
			Summary: err.Error(),
		}
		return result
	}

	if len(containerIds) == 0 {
		result := tasks.Result{
			Status:  tasks.None,
			Summary: "No Containerized Private Minions found",
			Payload: nil,
		}
		return result
	}

	//Query Docker for CPMs container inspect blobs by ids.
	containerJSONbytes, inspectErr := tasks.InspectContainersById(containerIds, p.executeCommand)

	if inspectErr != nil {
		result := tasks.Result{
			Status:  tasks.Error,
			Summary: inspectErr.Error(),
		}
		return result
	}

	//Any unexpected ENV values in the container get the value redacted. We want to preserve the keys though
	//As they can provide good configuration and environment context for troubleshooting.
	redactedContainersBytes, redactionError := tasks.RedactContainerEnv(containerJSONbytes, CPMenvWhitelist)

	if redactionError != nil {
		result := tasks.Result{
			Status:  tasks.Error,
			Summary: redactionError.Error(),
		}
		return result
	}

	//Umarshal bytes to slice of tasks.DockerContainers
	cpmContainers := []tasks.DockerContainer{}

	parseErr := json.Unmarshal(redactedContainersBytes, &cpmContainers)

	if parseErr != nil {
		result := tasks.Result{
			Status:  tasks.Error,
			Summary: parseErr.Error(),
		}
		return result
	}

	result := tasks.Result{
		Status:  tasks.Success,
		Summary: "Found Containerized Private Minions",
		Payload: cpmContainers,
	}

	stream := make(chan string)
	go streamContainers(redactedContainersBytes, stream)

	result.FilesToCopy = []tasks.FileCopyEnvelope{
		{
			Path:       "inspected-CPMs.json",
			Stream:     stream,
			Identifier: p.Identifier().String(),
		},
	}

	return result
}

func streamContainers(containerJSONbytes []byte, ch chan string) {
	defer close(ch)

	scanner := bufio.NewScanner(bytes.NewReader(containerJSONbytes))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		ch <- scanner.Text() + "\n"
	}
}
