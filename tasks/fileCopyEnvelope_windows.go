package tasks

import (
	"strings"
)

// IsExecutable - checks if a file is an executable file
func (e FileCopyEnvelope) IsExecutable() (bool, error) {
	if strings.HasSuffix(e.Path, ".exe") {
		return true, nil
	}
}
