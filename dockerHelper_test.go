package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"

	log "github.com/newrelic/NrDiag/logger"
)

func CreateDockerImage(imageName string, dockerFROM string, docker_cmd string, dockerLines []string) error {

	//Create the Dockerfile
	dockerfile, dockerfilecontent, err := CreateDockerfile(imageName, dockerFROM, docker_cmd, dockerLines)
	defer os.Remove(dockerfile) // clean up
	if err != nil {
		log.Fatal("Error creating integrationDockerfile: ", err)
	}

	log.Debug("Running docker build -f integrationDockerfile -t ", imageName, " .")

	cmdBuild := exec.Command("docker", "build", "-f", dockerfile, "-t", imageName, ".")

	output, cmdBuildErr := cmdBuild.CombinedOutput()

	if cmdBuildErr != nil {
		log.Infof("Error running cmd: docker build -f %s -t %s .", dockerfile, imageName)
		log.Info("Error was: ", string(output))
		log.Info("Content in dockerfile is: ", dockerfilecontent)
		return cmdBuildErr
	}
	return nil
}

//CreateDockerfile - This builds the raw Dockerfile from the slice of tests
func CreateDockerfile(imageName string, dockerFROM string, dockerCMD string, dockerfileLines []string) (string, []string, error) {

	f, err := ioutil.TempFile("temp", imageName)
	//Build base Dockerfile

	var dockerfile []string

	if _, err = f.WriteString("\r\n"); err != nil {
		log.Info("Error writing output file", err)
		return "", dockerfile, err
	}

	baseDockerFrom := []string{
		"FROM ubuntu:16.04",
		"RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -qqy unzip apt-transport-https ca-certificates",
	}

	baseWindowsDockerFrom := []string{
		"docker run mcr.microsoft.com/windows/nanoserver:1903",
		`SHELL ["powershell"]`,
	}

	baseDockerApp := []string{
		"COPY ./bin/linux/nrdiag /app/nrdiag",
		"WORKDIR /app",
	}
	baseWindowsDockerApp := []string{
		"COPY bin/win /app",
		"WORKDIR /app",
	}

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
		if _, err = f.WriteString(line + "\r\n"); err != nil {
			log.Info("Error writing output file", err)
			return "", dockerfile, err
		}
	}
	return f.Name(), dockerfile, nil
}

//RunDockerContainer - This runs the docker container from the image previously built
func RunDockerContainer(imageName string, hostsAdditions []string) (string, error) {
	//Create docker container based on test name

	args := []string{"run", "--rm"}
	if len(hostsAdditions) > 0 {
		for _, hostAddition := range hostsAdditions {
			args = append(args, "--add-host", hostAddition)
		}
	}
	args = append(args, imageName)
	cmd := exec.Command("docker", args...)

	out, cmdErr := cmd.CombinedOutput()
	if cmdErr != nil {
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
