package logs

import (
	"sort"

	tuikitIO "github.com/flowexec/tuikit/io"

	"github.com/flowexec/flow/pkg/store"
)

// UnifiedRecord joins an execution history record with its corresponding log archive entry (if available).
type UnifiedRecord struct {
	store.ExecutionRecord
	LogEntry *tuikitIO.ArchiveEntry
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
