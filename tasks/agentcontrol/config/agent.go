package config

import (
	"io/fs"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	log "github.com/newrelic/newrelic-diagnostics-cli/logger"
	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

type AgentControlConfigAgent struct {
	binaryChecker binaryFunc
}

type binaryFunc func() (bool, string)

type collectRoot struct {
	base   string
	subDir string
}

var acCollectLinux = []collectRoot{
	{base: "/var/lib/newrelic-agent-control", subDir: "fleet-data"},
	{base: "/etc/newrelic-agent-control", subDir: "local-data"},
}

var acCollectWindows = []collectRoot{
	{base: `C:\ProgramData\New Relic\newrelic-agent-control`, subDir: "fleet-data"},
	{base: `C:\Program Files\New Relic\newrelic-agent-control`, subDir: "local-data"},
}

func (p AgentControlConfigAgent) Identifier() tasks.Identifier {
	return tasks.IdentifierFromString("AgentControl/Config/Agent")
}

func (p AgentControlConfigAgent) Explain() string {
	return "Detect New Relic AgentControl agent"
}

func (p AgentControlConfigAgent) Dependencies() []string {
	return []string{}
}

func (p AgentControlConfigAgent) Execute(_ tasks.Options, _ map[string]tasks.Result) tasks.Result {
	binaryFound, binaryFilename := p.binaryChecker()
	if !binaryFound {
		log.Debug("No AgentControl agent found on system")
		return tasks.Result{
			Status:  tasks.None,
			Summary: tasks.NoAgentDetectedSummary,
		}
	}

	log.Debug("Identified AgentControl from binary: " + binaryFilename)

	roots := acCollectLinux
	if runtime.GOOS == "windows" {
		roots = acCollectWindows
	}

	filesToCopy := collectYAMLFiles(roots)

	return tasks.Result{
		Status:      tasks.Success,
		Summary:     "AgentControl agent identified from binary: " + binaryFilename,
		FilesToCopy: filesToCopy,
	}
}

func collectYAMLFiles(roots []collectRoot) []tasks.FileCopyEnvelope {
	var envelopes []tasks.FileCopyEnvelope

	for _, root := range roots {
		walkDir := filepath.Join(root.base, root.subDir)
		if !tasks.FileExists(walkDir) {
			log.Debug("agent-control data directory not found: " + walkDir)
			continue
		}

		_ = filepath.WalkDir(walkDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".yaml" && ext != ".yml" {
				return nil
			}

			// Compute path relative to base so the zip mirrors the real layout:
			//   fleet-data/<agent>/file.yaml
			//   local-data/<agent>/file.yaml
			rel, relErr := filepath.Rel(root.base, path)
			if relErr != nil {
				log.Debug("could not compute relative path for " + path)
				return nil
			}
			identifier := "AgentControl/Config/" + filepath.ToSlash(rel)

			envelopes = append(envelopes, tasks.FileCopyEnvelope{
				Path:       path,
				Identifier: identifier,
			})
			return nil
		})
	}

	return envelopes
}

// acWindowsBinaryPaths lists default install locations on Windows, where the
// binary is typically not added to PATH by the installer.
var acWindowsBinaryPaths = []string{
	`C:\Program Files\New Relic\newrelic-agent-control\newrelic-agent-control.exe`,
}

func checkForBinary() (bool, string) {
	for _, name := range []string{"newrelic-agent-control", "newrelic-agent-control.exe"} {
		if path, err := exec.LookPath(name); err == nil {
			log.Debug("Found agent-control binary in PATH: " + path)
			return true, path
		}
	}

	if runtime.GOOS == "windows" {
		for _, path := range acWindowsBinaryPaths {
			if tasks.FileExists(path) {
				log.Debug("Found agent-control binary at default Windows path: " + path)
				return true, path
			}
		}
	}

	log.Debug("No AgentControl binary found")
	return false, ""
}
