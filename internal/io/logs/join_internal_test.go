package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/flowexec/flow/v2/pkg/store"
)

// writeArchive creates an archive log file named "<id>__<ts>.log" with content, mirroring
// the tuikit archive naming so ListArchiveEntries can parse and index it.
func writeArchive(t *testing.T, dir, id, content string) string {
	t.Helper()
	name := id + "__2024-01-02-03-04-05.log"
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write archive: %v", err)
	}
	return path
}

func TestJoinRecords_InProgressResolvesByID(t *testing.T) {
	dir := t.TempDir()
	// A running record carries only its archive ID (no resolved path yet), the way
	// recordRunStart writes it before completion.
	writeArchive(t, dir, "run-abc", "partial output so far\n")

	idx := buildArchiveIndex(dir)
	records := []store.ExecutionRecord{{
		ID:        "run-abc",
		Ref:       "run ws/ns:thing",
		Status:    store.RunRunning,
		StartedAt: time.Now(),
		// LogArchiveID (the resolved path) intentionally empty — set only on completion.
	}}

	unified := joinRecords(records, idx)
	if len(unified) != 1 {
		t.Fatalf("expected 1 record, got %d", len(unified))
	}
	if unified[0].LogEntry == nil {
		t.Fatalf("expected in-progress record to resolve a log entry by ID")
	}
	got, err := unified[0].LogEntry.Read()
	if err != nil {
		t.Fatalf("read log entry: %v", err)
	}
	if got != "partial output so far\n" {
		t.Fatalf("unexpected content %q", got)
	}
}

func TestJoinRecords_CompletedResolvesByPath(t *testing.T) {
	dir := t.TempDir()
	path := writeArchive(t, dir, "run-done", "final output\n")

	idx := buildArchiveIndex(dir)
	records := []store.ExecutionRecord{{
		ID:           "run-done",
		Ref:          "run ws/ns:thing",
		Status:       store.RunCompleted,
		LogArchiveID: path, // completed records store the resolved path
		StartedAt:    time.Now(),
	}}

	unified := joinRecords(records, idx)
	if unified[0].LogEntry == nil || unified[0].LogEntry.Path != path {
		t.Fatalf("expected completed record to resolve by path, got %+v", unified[0].LogEntry)
	}
}

func TestJoinRecords_NoArchiveLeavesEntryNil(t *testing.T) {
	dir := t.TempDir()
	idx := buildArchiveIndex(dir)
	records := []store.ExecutionRecord{{ID: "missing", Ref: "run ws/ns:x", StartedAt: time.Now()}}
	unified := joinRecords(records, idx)
	if unified[0].LogEntry != nil {
		t.Fatalf("expected nil log entry when no archive exists")
	}
}
