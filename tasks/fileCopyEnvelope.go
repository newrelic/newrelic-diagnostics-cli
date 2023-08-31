package tasks

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
)

// FileCopyEnvelope - stores information about a file we want to copy
type FileCopyEnvelope struct {
	Path       string
	instance   int
	Stream     chan string
	Identifier string
}

// MarshalJSON - custom JSON marshaling for this task, we'll strip out the passphrase to keep it only in memory, not on disk
func (e FileCopyEnvelope) MarshalJSON() ([]byte, error) {
	//note: this technique can be used to return anything you want, including modified values or nothing at all.
	//anything that gets returned here ends up in the output json file
	if e.StoreName() == e.Name() {
		return json.Marshal(&struct {
			Path     string
			Name     string
			Streamed bool
		}{
			Path:     e.Path,
			Name:     e.Name(),
			Streamed: (e.Stream != nil),
		})

	}

	return json.Marshal(&struct {
		Path       string
		Name       string
		StoredName string
		Streamed   bool
		Identifier string
	}{
		Path:       e.Path,
		Name:       e.Name(),
		StoredName: e.StoreName(),
		Streamed:   (e.Stream != nil),
		Identifier: e.Identifier,
	})
}

// IncrementDuplicateCount - marks this as a duplicate of a file already being tracked
func (e *FileCopyEnvelope) IncrementDuplicateCount() {
	log.Debugf("Increment from %d\n", e.instance)
	e.instance++
}

// Name - the name of the file itself
func (e FileCopyEnvelope) Name() string {
	return filepath.ToSlash(filepath.Dir(e.Identifier)) + "/" + filepath.Base(e.Path) //This will get converted to category/subcategory and then converted to explicitly forward slash so it works on windows too
}

// StoreName - the name of the file as we (might) store it in the zip file
func (e FileCopyEnvelope) StoreName() string {
	if e.instance == 0 {
		return strings.Replace(e.Name(), "./", "", -1)
	}

	name, ext := e.SplitName()
	return fmt.Sprintf("%s(%d)%s", name, e.instance, ext)
}

// SplitName - the base name and file ext.
func (e FileCopyEnvelope) SplitName() (string, string) {
	ext := filepath.Ext(e.Name())
	fullname := e.Name()
	name := fullname[0 : len(fullname)-len(ext)]
	return name, ext
}

// StringsToFileCopyEnvelopes - converts an array of strings to an array of FileCopyEnvelopes
func StringsToFileCopyEnvelopes(fileList []string) []FileCopyEnvelope {
	envelopes := []FileCopyEnvelope{}
	for _, s := range fileList {
		envelopes = append(envelopes, FileCopyEnvelope{Path: s})
	}
	return envelopes
}
