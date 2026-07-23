package config

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func TestExecute_binaryNotFound(t *testing.T) {
	p := AgentControlConfigAgent{
		binaryChecker: func() (bool, string) { return false, "" },
	}

	result := p.Execute(tasks.Options{}, map[string]tasks.Result{})

	if result.Status != tasks.None {
		t.Errorf("Status = %v, want None", result.Status)
	}
	if result.Summary != tasks.NoAgentDetectedSummary {
		t.Errorf("Summary = %q, want NoAgentDetectedSummary", result.Summary)
	}
}

func TestExecute_binaryFound(t *testing.T) {
	p := AgentControlConfigAgent{
		binaryChecker: func() (bool, string) { return true, "/usr/bin/newrelic-agent-control" },
	}

	result := p.Execute(tasks.Options{}, map[string]tasks.Result{})

	if result.Status != tasks.Success {
		t.Errorf("Status = %v, want Success", result.Status)
	}
	if !strings.Contains(result.Summary, "/usr/bin/newrelic-agent-control") {
		t.Errorf("Summary %q does not contain binary path", result.Summary)
	}
}

// --- collectYAMLFiles ---

func TestCollectYAMLFiles_emptyRoots(t *testing.T) {
	if got := collectYAMLFiles([]collectRoot{}); len(got) != 0 {
		t.Errorf("expected 0 envelopes, got %d", len(got))
	}
}

func TestCollectYAMLFiles_missingDirectory(t *testing.T) {
	roots := []collectRoot{{base: "/nonexistent/path", subDir: "fleet-data"}}
	if got := collectYAMLFiles(roots); len(got) != 0 {
		t.Errorf("expected 0 envelopes for missing dir, got %d", len(got))
	}
}

func TestCollectYAMLFiles_onlyYAMLCollected(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, tmp, "fleet-data/agent-control/config.yaml", "key: val")
	mustWrite(t, tmp, "fleet-data/agent-control/notes.txt", "ignore me")
	mustWrite(t, tmp, "fleet-data/agent-control/data.yml", "key: val")
	mustWrite(t, tmp, "fleet-data/agent-control/binary.bin", "ignore me")

	got := collectYAMLFiles([]collectRoot{{base: tmp, subDir: "fleet-data"}})

	if len(got) != 2 {
		t.Errorf("expected 2 envelopes (.yaml + .yml), got %d", len(got))
	}
}

func TestCollectYAMLFiles_preservesSubdirStructure(t *testing.T) {
	// Simulates a real install with agent-control + one sub-agent across both data dirs.
	tmp := t.TempDir()
	mustWrite(t, tmp, "fleet-data/agent-control/instance_id.yaml", "")
	mustWrite(t, tmp, "fleet-data/agent-control/remote_config.yaml", "")
	mustWrite(t, tmp, "fleet-data/newrelic-agent/instance_id.yaml", "")
	mustWrite(t, tmp, "local-data/agent-control/local_config.yaml", "")
	mustWrite(t, tmp, "local-data/newrelic-agent/local_config.yaml", "")

	roots := []collectRoot{
		{base: tmp, subDir: "fleet-data"},
		{base: tmp, subDir: "local-data"},
	}

	envelopes := collectYAMLFiles(roots)
	if len(envelopes) != 5 {
		t.Fatalf("expected 5 envelopes, got %d", len(envelopes))
	}

	var got []string
	for _, e := range envelopes {
		got = append(got, e.Identifier)
	}
	sort.Strings(got)

	want := []string{
		"AgentControl/Config/fleet-data/agent-control/instance_id.yaml",
		"AgentControl/Config/fleet-data/agent-control/remote_config.yaml",
		"AgentControl/Config/fleet-data/newrelic-agent/instance_id.yaml",
		"AgentControl/Config/local-data/agent-control/local_config.yaml",
		"AgentControl/Config/local-data/newrelic-agent/local_config.yaml",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("identifiers mismatch:\n got:  %v\nwant: %v", got, want)
	}
}

func TestCollectYAMLFiles_zipPath(t *testing.T) {
	// Verifies the full chain: Identifier → FileCopyEnvelope.Name() → zip path.
	tmp := t.TempDir()
	mustWrite(t, tmp, "fleet-data/agent-control/instance_id.yaml", "id: test")

	envelopes := collectYAMLFiles([]collectRoot{{base: tmp, subDir: "fleet-data"}})
	if len(envelopes) != 1 {
		t.Fatalf("expected 1 envelope, got %d", len(envelopes))
	}

	got := envelopes[0].Name()
	want := "AgentControl/Config/fleet-data/agent-control/instance_id.yaml"
	if got != want {
		t.Errorf("Name() = %q, want %q", got, want)
	}
}

// mustWrite creates a file at base/relPath (using forward slashes), creating all
// parent directories as needed, and writes content into it.
func mustWrite(t *testing.T, base, relPath, content string) {
	t.Helper()
	full := filepath.Join(base, filepath.FromSlash(relPath))
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
