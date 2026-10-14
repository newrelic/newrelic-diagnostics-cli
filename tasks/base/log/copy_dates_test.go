package log

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/newrelic/newrelic-diagnostics-cli/tasks"
)

func writeLogFile(t *testing.T, dir string, name string, age time.Duration) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("log"), 0644); err != nil {
		t.Fatal(err)
	}
	modTime := time.Now().Add(-age)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatal(err)
	}
	return path
}

func collectedFileNames(elements []LogElement) map[string]bool {
	names := make(map[string]bool)
	for _, e := range elements {
		names[e.FileName] = true
	}
	return names
}

const day = 24 * time.Hour

func TestDetermineFilesDateKeepsDefaultCutoffForAgentLogsAfterAProfilerLog(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	profilerLog := writeLogFile(t, dir, "NewRelic.Profiler.1234.log", 2*time.Hour)
	agentLog := writeLogFile(t, dir, "newrelic_agent_MyApp.log", 3*day)

	// Act
	recent, old := determineFilesDate([]string{profilerLog, agentLog}, time.Now().Add(-7*day), time.Now().Add(-1*day))

	// Assert
	if len(recent) != 2 || len(old) != 0 {
		t.Errorf("expected both files to be recent, got recent=%v old=%v", recent, old)
	}
}

func TestStandardLocationsCollectsAgentLogsWithinSevenDaysAlongsideProfilerLogs(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	for i := 0; i < 20; i++ {
		writeLogFile(t, dir, fmt.Sprintf("NewRelic.Profiler.%d.log", i), 2*time.Hour)
	}
	writeLogFile(t, dir, "NewRelic.Profiler.old.log", 3*day)
	writeLogFile(t, dir, "newrelic_agent_MyApp.log", 3*day)

	// Act
	got := collectedFileNames(getLogPathsFromStandardLocations([]string{dir}, tasks.Options{Options: map[string]string{}}))

	// Assert
	if !got["newrelic_agent_MyApp.log"] {
		t.Error("expected the 3-day-old agent log to be collected")
	}
	if got["NewRelic.Profiler.old.log"] {
		t.Error("expected the 3-day-old profiler log to be skipped by the default 1-day profiler cutoff")
	}
	if !got["NewRelic.Profiler.0.log"] {
		t.Error("expected recent profiler logs to be collected")
	}
}

func TestStandardLocationsAppliesLastModifiedDateOverrideToProfilerLogs(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	writeLogFile(t, dir, "NewRelic.Profiler.new.log", 2*time.Hour)
	writeLogFile(t, dir, "NewRelic.Profiler.old.log", 3*day)
	writeLogFile(t, dir, "NewRelic.Profiler.ancient.log", 20*day)
	override := strconv.FormatInt(time.Now().Add(-10*day).Unix(), 10)

	// Act
	got := collectedFileNames(getLogPathsFromStandardLocations([]string{dir}, tasks.Options{Options: map[string]string{"lastModifiedDate": override}}))

	// Assert
	if !got["NewRelic.Profiler.old.log"] {
		t.Error("expected the 3-day-old profiler log to be collected when the override reaches back 10 days")
	}
	if got["NewRelic.Profiler.ancient.log"] {
		t.Error("expected the 20-day-old profiler log to be outside the override")
	}
}
