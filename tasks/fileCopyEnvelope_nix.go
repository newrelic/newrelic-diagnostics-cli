//go:build linux || darwin
// +build linux darwin

package tasks

import (
	"os"
)

// IsExecutable - checks if a file is an executable file
func (e FileCopyEnvelope) IsExecutable() (bool, error) {
	f, err := os.Open(e.Path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return false, err
	}

	fm := fi.Mode()
	return fm.IsRegular() && (fm.Perm()&0111) > 0, nil
}
