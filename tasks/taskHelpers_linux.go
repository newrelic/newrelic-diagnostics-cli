package tasks

import (
	"bytes"
	"os/exec"
)

// GetProcessorArch - calls uname -m to find the kernel architecture
func GetProcessorArch() (procType string, retErr error) {
	// x/sys/unix has a Uname implementation but...
	// it's not reliable. it will return the HOST's uname, not the CONTAINER, if using docker
	// that is why we are using exec.Command and running uname -m manually
	procTypeBytes, retErr := exec.Command("uname", "-m").Output()
	procTypeBytes = bytes.Trim(procTypeBytes[:], "\x00") // remove null characters
	procTypeBytes = bytes.Trim(procTypeBytes[:], "\n")   // remove new lines
	procType = string(procTypeBytes[:])
	return
}
