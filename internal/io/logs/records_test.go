package logs_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"go.uber.org/mock/gomock"

	"github.com/flowexec/flow/v2/internal/io/logs"
	"github.com/flowexec/flow/v2/pkg/store"
	storeMocks "github.com/flowexec/flow/v2/pkg/store/mocks"
)

func rec(ref string, exit int, at time.Time) logs.UnifiedRecord {
	return logs.UnifiedRecord{ExecutionRecord: store.ExecutionRecord{Ref: ref, ExitCode: exit, StartedAt: at}}
}

func provRec(ref, source, client, session string) logs.UnifiedRecord {
	return logs.UnifiedRecord{ExecutionRecord: store.ExecutionRecord{
		Ref:        ref,
		StartedAt:  time.Now(),
		Source:     source,
		ClientName: client,
		SessionID:  session,
	}}
}

func TestFilterRecords_EmptyFilterReturnsAll(t *testing.T) {
	now := time.Now()
	records := []logs.UnifiedRecord{
		rec("exec a/ns:one", 0, now),
		rec("exec b/ns:two", 1, now),
	}
	got := logs.FilterRecords(records, logs.RecordFilter{})
	if len(got) != 2 {
		t.Fatalf("expected all 2 records with empty filter, got %d", len(got))
	}
}

func TestFilterRecords_Workspace(t *testing.T) {
	records := []logs.UnifiedRecord{
		rec("exec alpha/ns:a", 0, time.Now()),
		rec("exec beta/ns:b", 0, time.Now()),
		rec("exec alpha/ns:c", 0, time.Now()),
	}
	got := logs.FilterRecords(records, logs.RecordFilter{Workspace: "alpha"})
	if len(got) != 2 {
		t.Fatalf("expected 2 alpha records, got %d", len(got))
	}
	for _, r := range got {
		if !strings.Contains(r.Ref, " alpha/") {
			t.Fatalf("unexpected workspace in filtered result: %q", r.Ref)
		}
	}
}

func TestFilterRecords_Status(t *testing.T) {
	records := []logs.UnifiedRecord{
		rec("exec a/ns:ok", 0, time.Now()),
		rec("exec a/ns:bad", 1, time.Now()),
		rec("exec a/ns:ok2", 0, time.Now()),
	}
	successes := logs.FilterRecords(records, logs.RecordFilter{Status: "success"})
	if len(successes) != 2 {
		t.Fatalf("expected 2 success records, got %d", len(successes))
	}
	failures := logs.FilterRecords(records, logs.RecordFilter{Status: "failure"})
	if len(failures) != 1 {
		t.Fatalf("expected 1 failure record, got %d", len(failures))
	}
}

func TestFilterRecords_StatusLifecycle(t *testing.T) {
	now := time.Now()
	running := logs.UnifiedRecord{ExecutionRecord: store.ExecutionRecord{
		Ref: "exec a/ns:live", Status: store.RunRunning, StartedAt: now,
	}}
	completed := rec("exec a/ns:done", 0, now) // canonical "completed"
	failed := rec("exec a/ns:bad", 1, now)     // canonical "failed"
	records := []logs.UnifiedRecord{running, completed, failed}

	cases := map[string]string{
		"running":   "exec a/ns:live",
		"completed": "exec a/ns:done",
		"success":   "exec a/ns:done", // alias
		"failed":    "exec a/ns:bad",
		"failure":   "exec a/ns:bad", // alias
	}
	for status, wantRef := range cases {
		got := logs.FilterRecords(records, logs.RecordFilter{Status: status})
		if len(got) != 1 {
			t.Fatalf("status %q: expected 1 record, got %d", status, len(got))
		}
		if got[0].Ref != wantRef {
			t.Fatalf("status %q: expected %q, got %q", status, wantRef, got[0].Ref)
		}
	}

	// A running record must NOT be counted as a success (exit code is still zero).
	if got := logs.FilterRecords(records, logs.RecordFilter{Status: "success"}); len(got) != 1 {
		t.Fatalf("expected 'success' to exclude the running record, got %d", len(got))
	}
}

func TestFilterRecords_Provenance(t *testing.T) {
	records := []logs.UnifiedRecord{
		provRec("exec a/ns:1", "mcp", "claude", "sess-1"),
		provRec("exec a/ns:2", "mcp", "cursor", "sess-2"),
		provRec("exec a/ns:3", "cli", "", ""),
	}

	if got := logs.FilterRecords(records, logs.RecordFilter{Source: "mcp"}); len(got) != 2 {
		t.Fatalf("expected 2 mcp records, got %d", len(got))
	}
	// Source matching is case-insensitive.
	if got := logs.FilterRecords(records, logs.RecordFilter{Source: "MCP"}); len(got) != 2 {
		t.Fatalf("expected case-insensitive source match, got %d", len(got))
	}
	if got := logs.FilterRecords(records, logs.RecordFilter{Session: "sess-2"}); len(got) != 1 {
		t.Fatalf("expected 1 record for session sess-2, got %d", len(got))
	} else if got[0].Ref != "exec a/ns:2" {
		t.Fatalf("unexpected record for session filter: %q", got[0].Ref)
	}
	if got := logs.FilterRecords(records, logs.RecordFilter{Client: "claude"}); len(got) != 1 {
		t.Fatalf("expected 1 claude record, got %d", len(got))
	}
	// Session + source combine (AND).
	got := logs.FilterRecords(records, logs.RecordFilter{Source: "mcp", Session: "sess-1"})
	if len(got) != 1 || got[0].Ref != "exec a/ns:1" {
		t.Fatalf("expected the mcp/sess-1 record only, got %+v", got)
	}
}

func TestFilterRecords_Since(t *testing.T) {
	old := time.Now().Add(-2 * time.Hour)
	mid := time.Now().Add(-30 * time.Minute)
	recent := time.Now()
	records := []logs.UnifiedRecord{
		rec("exec a/ns:old", 0, old),
		rec("exec a/ns:mid", 0, mid),
		rec("exec a/ns:new", 0, recent),
	}
	since := time.Now().Add(-1 * time.Hour)
	got := logs.FilterRecords(records, logs.RecordFilter{Since: since})
	if len(got) != 2 {
		t.Fatalf("expected 2 records after %v, got %d", since, len(got))
	}
}

func TestFilterRecords_Limit(t *testing.T) {
	now := time.Now()
	records := []logs.UnifiedRecord{
		rec("exec a/ns:1", 0, now),
		rec("exec a/ns:2", 0, now),
		rec("exec a/ns:3", 0, now),
		rec("exec a/ns:4", 0, now),
	}
	got := logs.FilterRecords(records, logs.RecordFilter{Limit: 2})
	if len(got) != 2 {
		t.Fatalf("expected limit to cap at 2, got %d", len(got))
	}

	// Limit greater than length should be a no-op.
	got = logs.FilterRecords(records, logs.RecordFilter{Limit: 99})
	if len(got) != 4 {
		t.Fatalf("expected all 4 records when limit > len, got %d", len(got))
	}
}

func TestFilterRecords_Combined(t *testing.T) {
	now := time.Now()
	past := now.Add(-24 * time.Hour)
	records := []logs.UnifiedRecord{
		rec("exec alpha/ns:old-ok", 0, past),
		rec("exec alpha/ns:new-ok", 0, now),
		rec("exec alpha/ns:new-fail", 1, now),
		rec("exec beta/ns:new-ok", 0, now),
	}
	got := logs.FilterRecords(records, logs.RecordFilter{
		Workspace: "alpha",
		Status:    "success",
		Since:     now.Add(-1 * time.Hour),
	})
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 record matching all filters, got %d", len(got))
	}
	if got[0].Ref != "exec alpha/ns:new-ok" {
		t.Fatalf("unexpected record: %q", got[0].Ref)
	}
}

func TestLoadRecords_SortsResultsByStartedAtDescending(t *testing.T) {
	ctrl := gomock.NewController(t)
	ds := storeMocks.NewMockDataStore(ctrl)
	t1 := time.Now().Add(-2 * time.Hour)
	t2 := time.Now().Add(-1 * time.Hour)
	t3 := time.Now()

	ds.EXPECT().ListExecutionRefs().Return([]string{"one"}, nil)
	ds.EXPECT().GetExecutionHistory("one", 10).Return([]store.ExecutionRecord{
		{Ref: "a", StartedAt: t1},
		{Ref: "b", StartedAt: t3},
		{Ref: "c", StartedAt: t2},
	}, nil)

	got, err := logs.LoadRecords(ds, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("expected 3 records, got %d", len(got))
	}
	if got[0].Ref != "b" || got[1].Ref != "c" || got[2].Ref != "a" {
		t.Fatalf("expected descending StartedAt order b,c,a; got %s,%s,%s", got[0].Ref, got[1].Ref, got[2].Ref)
	}
	if got[0].LogEntry != nil {
		t.Fatalf("expected LogEntry to be nil when logsDir is empty")
	}
}

func TestLoadRecords_NilDataStoreReturnsEmpty(t *testing.T) {
	got, err := logs.LoadRecords(nil, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil records for nil data store, got %d", len(got))
	}
}

func TestLoadRecordsForRef_NilDataStoreReturnsEmpty(t *testing.T) {
	got, err := logs.LoadRecordsForRef(nil, "", "ref", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil records for nil data store, got %d", len(got))
	}
}

func TestLoadRecords_PropagatesListRefsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	ds := storeMocks.NewMockDataStore(ctrl)
	wantErr := errors.New("boom")
	ds.EXPECT().ListExecutionRefs().Return(nil, wantErr)

	_, err := logs.LoadRecords(ds, "")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestLoadRecords_SkipsRefsThatErrorDuringHistoryLookup(t *testing.T) {
	ctrl := gomock.NewController(t)
	ds := storeMocks.NewMockDataStore(ctrl)
	now := time.Now()

	ds.EXPECT().ListExecutionRefs().Return([]string{"good", "bad"}, nil)
	ds.EXPECT().GetExecutionHistory("good", 10).Return([]store.ExecutionRecord{
		{Ref: "good", StartedAt: now},
	}, nil)
	ds.EXPECT().GetExecutionHistory("bad", 10).Return(nil, errors.New("history error"))

	got, err := logs.LoadRecords(ds, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Ref != "good" {
		t.Fatalf("expected only the 'good' record to survive; got %+v", got)
	}
}

func TestStatusText(t *testing.T) {
	running := logs.UnifiedRecord{ExecutionRecord: store.ExecutionRecord{Status: store.RunRunning}}
	if got := logs.StatusText(running); got != "running" {
		t.Fatalf("expected 'running', got %q", got)
	}
	// Legacy records (no Status) derive from exit code.
	if got := logs.StatusText(rec("x", 0, time.Now())); got != "ok" {
		t.Fatalf("expected 'ok', got %q", got)
	}
	if got := logs.StatusText(rec("x", 2, time.Now())); got != "exit(2)" {
		t.Fatalf("expected 'exit(2)', got %q", got)
	}
}

func TestLoadRecords_ReconcilesStaleRunningRecord(t *testing.T) {
	ctrl := gomock.NewController(t)
	ds := storeMocks.NewMockDataStore(ctrl)
	now := time.Now()

	// A "running" record whose process is no longer alive (an implausibly high PID).
	ds.EXPECT().ListExecutionRefs().Return([]string{"one"}, nil)
	ds.EXPECT().GetExecutionHistory("one", 10).Return([]store.ExecutionRecord{
		{ID: "run-1", Ref: "one", StartedAt: now, Status: store.RunRunning, PID: 2_000_000_000},
	}, nil)
	// The stale record is persisted back as failed.
	ds.EXPECT().RecordExecution(gomock.Any()).DoAndReturn(func(r store.ExecutionRecord) error {
		if r.Status != store.RunFailed {
			t.Fatalf("expected stale record persisted as failed, got %q", r.Status)
		}
		return nil
	})

	got, err := logs.LoadRecords(ds, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if logs.StatusText(got[0]) == "running" {
		t.Fatalf("expected stale record to no longer render as running")
	}
}

func TestLoadRecordsForRef_UsesGivenRefAndLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	ds := storeMocks.NewMockDataStore(ctrl)
	now := time.Now()

	ds.EXPECT().GetExecutionHistory("my-ref", 5).Return([]store.ExecutionRecord{
		{Ref: "my-ref", StartedAt: now},
	}, nil)

	got, err := logs.LoadRecordsForRef(ds, "", "my-ref", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 || got[0].Ref != "my-ref" {
		t.Fatalf("unexpected records: %+v", got)
	}
}
