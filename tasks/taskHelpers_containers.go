package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

type DockerInfo struct {
	Driver        string
	ServerVersion string
	MemTotal      int64
	NCPU          int
}

//DockerContainer is a truncated struct of the docker inspect blob
//Since we can write the full blob to a file, the purpose of this struct is to limit ourselves
//only to values we are interested in for validation in our tasks.
type DockerContainer struct {
	Id       string
	Created  string
	State    ContainerState
	Name     string
	Driver   string
	Platform string
	Mounts   []ContainerMount
	Config   ContainerConfig
}

type ContainerState struct {
	Status     string
	Running    bool
	Pause      bool
	Restarting bool
	ExitCode   int
	Error      string
	StartedAt  string
	FinishedAt string
}

type ContainerMount struct {
	Source      string
	Destination string
	Mode        string
	RW          bool
}

type ContainerConfig struct {
	User string
	Env  []string
}

func GetDockerInfoCLIBytes(cmdExec CmdExecFunc) ([]byte, error) {

	cmdOutBytes, err := cmdExec("docker", "info", "--format", "'{{json .}}'")

	if err != nil {
		// Check specifically for unsuccessful exit by Docker info. e.g. Docker installed but Daemon is not running
		if _, isExitErr := err.(*exec.ExitError); isExitErr {
			return cmdOutBytes, err
		}
		return nil, err
	}

	trimmedBytes := bytes.TrimSpace(cmdOutBytes)
	trimmedBytes = bytes.Trim(trimmedBytes, "'")

	return trimmedBytes, nil
}

func NewDockerInfoFromBytes(dockerInfoBytes []byte) (DockerInfo, error) {
	dockerInfo := DockerInfo{}

	parseErr := json.Unmarshal(dockerInfoBytes, &dockerInfo)

	return dockerInfo, parseErr
}

//Example of the command we want to construct:
//docker ps -q --last 4 --filter label=name=synthetics-minion --filter status=running
//list all the last 4 containers filtered to the label 'name' with value 'synthetics-minion' format output with container id and filtered to status running
//https://docs.docker.com/engine/reference/commandline/ps/
//@label = expected docker image label key used by the container eg. "name"
//@value = value of the label eg. "synthetics-minion"
//@numberOf = max number of containers ids to return
//@includeExited = include both active and exited containers
//@cmdExec =  command line executor dependency
func GetContainerIdsByLabel(label string, value string, numberOf int, includeExited bool, cmdExec CmdExecFunc) ([]string, error) {

	var foundContainerIds []string
	//default no filter for only running containers
	statusFilterArg := ""

	if !includeExited {
		statusFilterArg = "--filter status=running"
	}

	queryArgs := fmt.Sprintf(`ps -q --last %v --filter label=%s=%s %s`, numberOf, label, value, statusFilterArg)
	queryArgsArray := strings.Fields(queryArgs)

	cmdOutBytes, err := cmdExec("docker", queryArgsArray...)

	if err != nil {
		return nil, errors.New("error querying for container: " + err.Error() + ": " + string(cmdOutBytes))
	}

	cmdOutString := string(cmdOutBytes)

	if len(cmdOutString) > 0 {
		containerIdsTrimmed := strings.TrimSpace(cmdOutString)
		foundContainerIds = strings.Split(containerIdsTrimmed, "\n")
	}

	return foundContainerIds, nil
}

//Get inspect blobs of containers from slice of ids. Docker client will take several ids as arguments
//and return blobs for each.
func InspectContainersById(containerIds []string, cmdExec CmdExecFunc) ([]byte, error) {
	//docker inspect can take multiple object id arguments in single command
	// will output objects a JSON array

	queryArgsArray := []string{"inspect"}
	queryArgsArray = append(queryArgsArray, containerIds...)

	cmdOutBytes, cmdExecErr := cmdExec("docker", queryArgsArray...)

	if cmdExecErr != nil {
		return []byte{}, errors.New(cmdExecErr.Error() + " " + string(cmdOutBytes))
	}

	return cmdOutBytes, nil
}

//Redact values of unwhitelisted environment variables.
func RedactContainerEnv(containers []byte, whitelist []string) ([]byte, error) {
	//expect a JSON array, so we unmarshal into a slice of interfaces
	parsedContainers := []interface{}{}
	//we are going to search for values in the whitelist later,
	//so we want to sort it for effective search
	sort.Strings(whitelist)

	unMarshalError := json.Unmarshal(containers, &parsedContainers)

	if unMarshalError != nil {
		log.Debug("Error parsing json:", unMarshalError)
		return []byte{}, unMarshalError
	}

	//for every container
	for _, container := range parsedContainers {
		//we unmarshalled as anon interfaces so we have to coerce types on our path
		//JSON format: {"Config":{"Env":["FOO=BAR", "BAR=FOO"]}}
		//The shorthand without check for nil is : env = container.(map[string]interface{})["Config"].(map[string]interface{})["Env"].([]interface{})
		var env []interface{}
		containerMap, containerMapOk := container.(map[string]interface{})

		if containerMapOk {
			config, configOk := containerMap["Config"].(map[string]interface{})

			if configOk {
				env = config["Env"].([]interface{})
			}
		}

		if env == nil {
			return []byte{}, errors.New("could not find Env variables in container inspect blob")
		}

		//for every env variable of a container
		for i, envVar := range env {
			envVarPair := strings.Split(envVar.(string), "=")

			// we assume docker inspect will return env vars in `key=value` format
			// if that ever changes, we'll play it safe and redact the whole thing
			if len(envVarPair) < 2 {
				env[i] = "_REDACTED_"
				continue
			}

			//whitelist is uppercase so should our search term
			envVarNameUpper := strings.ToUpper(envVarPair[0])
			searchIndex := sort.SearchStrings(whitelist, envVarNameUpper)
			//SearchStrings will return index of where search value should be inserted in ordered list if not found
			//So we check if the index value matches search value to validate if present
			//And that index is not len of whitelist
			if (searchIndex != len(whitelist)) && (!strings.EqualFold(whitelist[searchIndex], envVarNameUpper)) {
				//if searched value not found we redact the value and reconstruct the string
				env[i] = fmt.Sprintf(`%s=_REDACTED_`, envVarPair[0])
			}
		}
	}

	//pretty marshal back to JSON with 4 space indents
	redactedJSON, marshalError := json.MarshalIndent(parsedContainers, "", "    ")
	if marshalError != nil {
		log.Debug("Error with marshal to json:", marshalError)
		return []byte{}, marshalError
	}

	return redactedJSON, nil
}

//StreamContainerLogsById will perform a buffered stream of `docker logs` command for a containerId.
//We perform a buffered read since logging can be quite large and we dont want to put it all in memory.
//@containerId - the containerId to collect logs from
//@bufferedCmdExec - the buffered command exec to use, which should return a scanner.
//@sw - StreamWrapper that has the channel to send log output to and the channel to send errors through

func StreamContainerLogsById(containerId string, bufferedCmdExec BufferedCommandExecFunc, sw *StreamWrapper) {
	defer close(sw.Stream)

	//150 MB - in case we find need to impose a read limit later. For now defaulting to no limit with 0
	//var MAX_LOG_OUTPUT_SIZE int64 = 150 * 1000 * 1024
	var MAX_LOG_OUTPUT_SIZE int64 = 0

	//docker logs <container-id>
	queryArgsArray := []string{"logs"}
	queryArgsArray = append(queryArgsArray, containerId)

	cmdOutScanner, cmdExecErr := bufferedCmdExec(MAX_LOG_OUTPUT_SIZE, "docker", queryArgsArray...)

	if cmdExecErr != nil {
		sw.ErrorStream <- cmdExecErr
		return
	}
	sw.ErrorWg.Done()

	for cmdOutScanner.Scan() {
		sw.Stream <- cmdOutScanner.Text() + "\n"
	}
}
