package logs

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"

	"github.com/flowexec/flow/v2/internal/utils/process"
	"github.com/flowexec/flow/v2/pkg/store"
)

// RecordFilter holds optional criteria for filtering unified records.
type RecordFilter struct {
	Workspace string
	Status    string // "success" or "failure"
	Since     time.Time
	Limit     int
}

// extractWorkspace parses the workspace from a ref formatted as "verb ws/ns:name".
func extractWorkspace(ref string) string {
	_, rest, ok := strings.Cut(ref, " ")
	if !ok {
		return ""
	}
	ws, _, ok := strings.Cut(rest, "/")
	if !ok {
		return ""
	}
	return ws
}

// FilterRecords applies the filter criteria to a slice of unified records.
func FilterRecords(records []UnifiedRecord, f RecordFilter) []UnifiedRecord {
	var filtered []UnifiedRecord
	for _, r := range records {
		if f.Workspace != "" {
			// Refs are formatted as "verb ws/ns:name" — workspace is between the space and the first "/".
			ws := extractWorkspace(r.Ref)
			if ws != f.Workspace {
				continue
			}
		}
		if f.Status != "" {
			switch f.Status {
			case "success":
				if r.ExitCode != 0 {
					continue
				}
			case "failure":
				if r.ExitCode == 0 {
					continue
				}
			}
		}
		if !f.Since.IsZero() && r.StartedAt.Before(f.Since) {
			continue
		}
		filtered = append(filtered, r)
	}
	if f.Limit > 0 && len(filtered) > f.Limit {
		filtered = filtered[:f.Limit]
	}
	return filtered
}

// UnifiedRecord joins an execution history record with its corresponding log archive entry (if available).
type UnifiedRecord struct {
	store.ExecutionRecord
	LogEntry *tuikitIO.ArchiveEntry
}

// CanonicalStatus returns the lifecycle status of a record (running/completed/failed), deriving it
// from the exit code for legacy records that predate the Status field.
func CanonicalStatus(r UnifiedRecord) store.RunStatus {
	if r.Status != "" {
		return r.Status
	}
	if r.ExitCode == 0 {
		return store.RunCompleted
	}
	return store.RunFailed
}

// StatusText returns a human-readable status for display: "running" for in-progress runs,
// otherwise "ok" or "exit(N)" derived from the exit code.
func StatusText(r UnifiedRecord) string {
	if CanonicalStatus(r) == store.RunRunning {
		return "running"
	}
	if r.ExitCode == 0 {
		return "ok"
	}
	return fmt.Sprintf("exit(%d)", r.ExitCode)
}

// reconcileStale marks any "running" record whose process is no longer alive as failed, persisting
// the correction back to the store. This prevents an abnormally-terminated run from lingering as
// active, mirroring the stale-detection done for background runs in `flow logs --running`.
func reconcileStale(ds store.DataStore, records []store.ExecutionRecord) []store.ExecutionRecord {
	for i := range records {
		r := &records[i]
		if r.Status != store.RunRunning || r.PID == 0 || process.Alive(r.PID) {
			continue
		}
		now := time.Now()
		r.Status = store.RunFailed
		r.ExitCode = 1
		r.CompletedAt = &now
		if r.Error == "" {
			r.Error = "process exited unexpectedly"
		}
		if ds != nil && r.ID != "" {
			_ = ds.RecordExecution(*r)
		}
	}
	return records
}

// LoadRecords retrieves all execution history from the data store, joined with any matching log archive entries.
// If ds is nil, returns empty (log-only fallback is not supported without metadata).
func LoadRecords(ds store.DataStore, logsDir string) ([]UnifiedRecord, error) {
	if ds == nil {
		return nil, nil
	}

	records, err := getAllExecutionHistory(ds)
	if err != nil {
		return nil, err
	}
	records = reconcileStale(ds, records)

	archiveIndex := buildArchiveIndex(logsDir)
	return joinRecords(records, archiveIndex), nil
}

// LoadRecordsForRef retrieves execution history for a specific ref, joined with matching log archive entries.
func LoadRecordsForRef(ds store.DataStore, logsDir string, ref string, limit int) ([]UnifiedRecord, error) {
	if ds == nil {
		return nil, nil
	}

	records, err := ds.GetExecutionHistory(ref, limit)
	if err != nil {
		return nil, err
	}
	records = reconcileStale(ds, records)

	archiveIndex := buildArchiveIndex(logsDir)
	return joinRecords(records, archiveIndex), nil
}

// getAllExecutionHistory retrieves recent history across all refs, up to 10 records per ref.
func getAllExecutionHistory(ds store.DataStore) ([]store.ExecutionRecord, error) {
	refs, err := ds.ListExecutionRefs()
	if err != nil {
		return nil, err
	}
	var all []store.ExecutionRecord
	for _, ref := range refs {
		records, err := ds.GetExecutionHistory(ref, 10)
		if err != nil {
			continue
		}
		all = append(all, records...)
	}
	return all, nil
}

// buildArchiveIndex loads log archive entries from disk and indexes them by path for O(1) lookup.
func buildArchiveIndex(logsDir string) map[string]tuikitIO.ArchiveEntry {
	entries, err := tuikitIO.ListArchiveEntries(logsDir)
	if err != nil || len(entries) == 0 {
		return nil
	}
	index := make(map[string]tuikitIO.ArchiveEntry, len(entries))
	for _, e := range entries {
		index[e.Path] = e
	}
	return index
}

// joinRecords merges execution records with their log archive entries and sorts by StartedAt descending.
func joinRecords(records []store.ExecutionRecord, archiveIndex map[string]tuikitIO.ArchiveEntry) []UnifiedRecord {
	unified := make([]UnifiedRecord, 0, len(records))
	for _, r := range records {
		ur := UnifiedRecord{ExecutionRecord: r}
		if archiveIndex != nil {
			if entry, ok := archiveIndex[r.LogArchiveID]; ok {
				ur.LogEntry = &entry
			}
		}
		unified = append(unified, ur)
	}
	sort.Slice(unified, func(i, j int) bool {
		return unified[i].StartedAt.After(unified[j].StartedAt)
	})
	return unified
}
