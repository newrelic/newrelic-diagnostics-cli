package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

const (
	dockerBuildTimeout = 10 * time.Minute
	dockerRunTimeout   = 10 * time.Minute
)

var dockerBuildRetryPatterns = []string{
	"hcs::CreateComputeSystem",
	"failed to connect to the docker API",
	"docker_engine",
}

func CreateDockerImage(imageName string, dockerFROM string, docker_cmd string, dockerLines []string) error {

	//Create the Dockerfile
	dockerfile, err := CreateDockerfile(imageName, dockerFROM, docker_cmd, dockerLines)
	defer os.Remove(dockerfile) // clean up
	if err != nil {
		log.Info("Error creating integrationDockerfile", err)
	}

	log.Debug("Running docker build -f integrationDockerfile -t ", imageName, " .")

	const maxAttempts = 3
	var output []byte
	var cmdBuildErr error
	var timedOut bool
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), dockerBuildTimeout)
		cmdBuild := exec.CommandContext(ctx, "docker", "build", "-f", dockerfile, "-t", imageName, ".")
		output, cmdBuildErr = cmdBuild.CombinedOutput()
		timedOut = ctx.Err() == context.DeadlineExceeded
		cancel()

		if cmdBuildErr == nil {
			break
		}
		if timedOut {
			log.Info("Docker build TIMED OUT for ", imageName, " after ", dockerBuildTimeout, " (attempt ", attempt, " of ", maxAttempts, ")")
			log.Info("Partial build output for ", imageName, ":\n", string(output))
			logDockerDiagnostics(imageName)
		}
		if attempt == maxAttempts || !isTransientDockerError(string(output)) {
			break
		}
		log.Info("Transient docker build error for ", imageName, " (attempt ", attempt, " of ", maxAttempts, "), retrying in ", attempt*5, "s")
		time.Sleep(time.Duration(attempt*5) * time.Second)
	}

	if cmdBuildErr != nil {
		log.Info("Error running docker build -", cmdBuildErr)
		log.Info("Error was ", string(output))
		return cmdBuildErr
	}
	log.Info("Docker build output for ", imageName, ":\n", string(output))
	return nil
}

func isTransientDockerError(output string) bool {
	for _, pattern := range dockerBuildRetryPatterns {
		if strings.Contains(output, pattern) {
			return true
		}
	}
	return false
}

func logDockerDiagnostics(imageName string) {
	log.Info("=== Docker diagnostics for ", imageName, " ===")

	psCtx, psCancel := context.WithTimeout(context.Background(), 30*time.Second)
	psOut, psErr := exec.CommandContext(psCtx, "docker", "ps", "-a").CombinedOutput()
	psCancel()
	if psErr != nil {
		log.Info("docker ps -a failed: ", psErr, " output: ", string(psOut))
	} else {
		log.Info("docker ps -a:\n", string(psOut))
	}

	infoCtx, infoCancel := context.WithTimeout(context.Background(), 30*time.Second)
	infoOut, infoErr := exec.CommandContext(infoCtx, "docker", "info").CombinedOutput()
	infoCancel()
	if infoErr != nil {
		log.Info("docker info failed: ", infoErr, " output: ", string(infoOut))
	} else {
		log.Info("docker info:\n", string(infoOut))
	}

	log.Info("=== End docker diagnostics for ", imageName, " ===")
}

// CreateDockerfile - This builds the raw Dockerfile from the slice of tests
func CreateDockerfile(imageName string, dockerFROM string, dockerCMD string, dockerfileLines []string) (string, error) {
	f, _ := os.CreateTemp("temp", imageName)
	//Build base Dockerfile

	if _, err := f.WriteString("\r\n"); err != nil {
		log.Info("Error writing output file", err)
		return "", err
	}

	baseDockerFrom := []string{
		"FROM ubuntu:22.04",
		"RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -qqy unzip apt-transport-https ca-certificates",
	}

	baseWindowsDockerFrom := []string{
		"FROM mcr.microsoft.com/windows/servercore:ltsc2025",
		`SHELL ["powershell"]`,
		"RUN NET USER nrdiagadmin /add",
		"RUN NET LOCALGROUP administrators /add nrdiagadmin",
		"USER nrdiagadmin",
	}

	baseDockerApp := []string{
		"COPY ./bin/linux/nrdiag /app/nrdiag",
		"WORKDIR /app",
	}
	baseWindowsDockerApp := []string{
		"COPY bin/win /app",
		"WORKDIR /app",
	}

	var dockerfile []string
	if runtime.GOOS == "windows" && dockerFROM == "" {
		dockerfile = append(baseWindowsDockerFrom, baseWindowsDockerApp...)
		dockerfile = append(dockerfile, dockerfileLines...)
	} else if dockerFROM == "" {
		dockerfile = append(baseDockerFrom, baseDockerApp...)
		dockerfile = append(dockerfile, dockerfileLines...)
	} else if runtime.GOOS == "windows" {
		dockerfile = append(dockerfile, "FROM "+dockerFROM)
		dockerfile = append(dockerfile, baseWindowsDockerApp...)
		dockerfile = append(dockerfile, dockerfileLines...)
	} else {
		dockerfile = append(dockerfile, "FROM "+dockerFROM)
		dockerfile = append(dockerfile, baseDockerApp...)
		dockerfile = append(dockerfile, dockerfileLines...)
	}

	var cmdPrefix string
	var binaryName string
	var cmdSuffix = "\"]"
	if runtime.GOOS == "windows" {
		cmdPrefix = "CMD [\"powershell\", \""
		binaryName = "./nrdiag_x64.exe -y"
	} else {
		cmdPrefix = "CMD [\"/bin/sh\", \"-c\", \""
		binaryName = "./nrdiag"
	}

	var cmdLine string
	if dockerCMD == "" {
		cmdLine = cmdPrefix + binaryName + cmdSuffix
	} else {
		cmdLine = cmdPrefix + dockerCMD + cmdSuffix
	}
	dockerfile = append(dockerfile, cmdLine)

	for _, line := range dockerfile {
		log.Debug(line)
		if _, err := f.WriteString(line + "\r\n"); err != nil {
			log.Info("Error writing output file", err)
			return "", err
		}
	}
	return f.Name(), nil
}

// RunDockerContainer - This runs the docker container from the image previously built
func RunDockerContainer(imageName string, hostsAdditions []string) (string, error) {
	//Create docker container based on test name

	args := []string{"run", "--rm"}
	if len(hostsAdditions) > 0 {
		for _, hostAddition := range hostsAdditions {
			args = append(args, "--add-host", hostAddition)
		}
	}
	args = append(args, imageName)

	ctx, cancel := context.WithTimeout(context.Background(), dockerRunTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "docker", args...)

	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Info("Docker run TIMED OUT for ", imageName, " after ", dockerRunTimeout)
			log.Info("Partial run output for ", imageName, ":\n", string(out))
			logDockerDiagnostics(imageName)
		}
		// Docker daemon returns exit code 125 in jenkins but runs normally otherwise
		if cmdErr.Error() == "exit status 125" {
			return string(out[:]), nil
		}
		log.Info("Error running docker run", cmdErr)
		log.Info("Error was ", string(out))
		return string(out[:]), cmdErr
	}

	logs := string(out[:])
	// Full verbose output
	log.Debug(logs)

	return logs, nil
}
