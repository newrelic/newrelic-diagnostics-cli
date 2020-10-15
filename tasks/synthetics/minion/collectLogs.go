package minion

import (
	"fmt"
	"sync"

	log "github.com/newrelic/NrDiag/logger"
	"github.com/newrelic/NrDiag/tasks"
)

//

// SyntheticsMinionDetectCPM - This struct defined the sample plugin which can be used as a starting point
type SyntheticsMinionCollectLogs struct {
	executeCommand tasks.BufferedCommandExecFunc
}

// Identifier - This returns the Category, Subcategory and Name of each task
func (p SyntheticsMinionCollectLogs) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("Synthetics/Minion/CollectLogs")
}

// Explain - Returns the help text for each individual task
func (p SyntheticsMinionCollectLogs) Explain() string {
	return "Collect logs of found Containerized Private Minions"
}

// Dependencies - Returns the dependencies for each task.
func (p SyntheticsMinionCollectLogs) Dependencies() []string {
	return []string{"Synthetics/Minion/DetectCPM"}
}

// Execute - The core work within each task
func (p SyntheticsMinionCollectLogs) Execute(options tasks.Options, upstream map[string]tasks.Result) tasks.Result {

	detectCPMResult := upstream["Synthetics/Minion/DetectCPM"]

	if detectCPMResult.Status != tasks.Success {
		return tasks.Result{
			Status:  tasks.None,
			Summary: "No CPMs detected to collect logs from",
		}
	}

	// We pipe docker log output to streams, and pass those stream (unconsumed) to FileCopyEnvelopes returned in the task result
	// These streams are then consumed after all tasks have completed
	logFileCopyEnvelopes, cmdErrors := initStreamsForFileCopy(detectCPMResult.Payload.([]tasks.DockerContainer), p.Identifier().String(), p.executeCommand)

	if len(cmdErrors) > 0 {

		return tasks.Result{
			Status:  tasks.Error,
			Summary: collectErrorsSliceToString(cmdErrors),
		}
	}

	return tasks.Result{
		Status:      tasks.Success,
		Summary:     "Logs from CPMs collected",
		FilesToCopy: logFileCopyEnvelopes,
	}
}

//Initializes streams for the output the `docker logs` command (executed by StreamContainerLogsById) for each container provided,
//Each stream initialized is then added to a new fileCopyEnvelope which are collected into a slice and returned
func initStreamsForFileCopy(containers []tasks.DockerContainer, taskIdentifier string, bufferedCommandExec tasks.BufferedCommandExecFunc) ([]tasks.FileCopyEnvelope, []error) {
	fileCopyEnvelopes := []tasks.FileCopyEnvelope{}
	cmdErrors := []error{}

	errStream := make(chan error)
	var errWg sync.WaitGroup

	for _, container := range containers {
		stream := make(chan string)

		log.Debugf("initStream for container: %s\n", container.Id)

		logStreamWrapper := tasks.StreamWrapper{
			Stream:      stream,
			ErrorStream: errStream,
			ErrorWg:     &errWg,
		}

		errWg.Add(1)

		go tasks.StreamContainerLogsById(container.Id, bufferedCommandExec, &logStreamWrapper)

		logEnvelope := tasks.FileCopyEnvelope{
			Path:       fmt.Sprintf("%s-minion.log", container.Id),
			Stream:     logStreamWrapper.Stream,
			Identifier: taskIdentifier,
		}

		fileCopyEnvelopes = append(fileCopyEnvelopes, logEnvelope)
	}

	go collectErrorsFromStream(errStream, &errWg, &cmdErrors)
	//Wait until streams have been initialized for each container before ending execution
	log.Debugf("Blocking initStreamsForFileCopy until all log streams created\n")
	errWg.Wait()
	close(errStream)
	log.Debugf("All log streams completed\n")
	return fileCopyEnvelopes, cmdErrors

}

func collectErrorsSliceToString(errors []error) string {
	errorString := ""

	for _, err := range errors {
		errorString = errorString + fmt.Sprintf("Error collecting logs from containers: %s\n", err.Error())
	}

	return errorString
}

func collectErrorsFromStream(errStream chan error, errWg *sync.WaitGroup, errCollection *[]error) {
	for err := range errStream {
		*errCollection = append(*errCollection, err)
		errWg.Done()
	}
}
