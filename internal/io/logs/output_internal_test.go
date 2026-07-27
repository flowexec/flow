package logs

import (
	"strings"
	"testing"
	"time"

	"github.com/flowexec/flow/v2/pkg/store"
)

func TestToRecordOutput_RunIdentity(t *testing.T) {
	started := time.Date(2026, 7, 27, 17, 20, 29, 0, time.UTC)
	completed := started.Add(45 * time.Millisecond)

	t.Run("carries the run ID, completion time and spec", func(t *testing.T) {
		out := toRecordOutput(UnifiedRecord{ExecutionRecord: store.ExecutionRecord{
			ID:          "4d75c29b-36e5-4b40-b840-98951746254e",
			Ref:         "exec flow/spec-build",
			StartedAt:   started,
			CompletedAt: &completed,
			Duration:    45 * time.Millisecond,
			Status:      store.RunCompleted,
			Spec:        "verb: run\nexec:\n  cmd: echo hi\n",
			Label:       "build",
		}}, ContentOptions{}, false)

		if out.ID != "4d75c29b-36e5-4b40-b840-98951746254e" {
			t.Errorf("expected the run ID to survive; got %q", out.ID)
		}
		if out.CompletedAt != completed.Format(time.RFC3339) {
			t.Errorf("expected RFC3339 completedAt, got %q", out.CompletedAt)
		}
		if !strings.Contains(out.Spec, "cmd: echo hi") {
			t.Errorf("expected the spec to survive; got %q", out.Spec)
		}
	})

	t.Run("omits completedAt for a run still in progress", func(t *testing.T) {
		out := toRecordOutput(UnifiedRecord{ExecutionRecord: store.ExecutionRecord{
			Ref:       "run flow/build",
			StartedAt: started,
			Status:    store.RunRunning,
		}}, ContentOptions{}, false)

		if out.CompletedAt != "" {
			t.Errorf("expected an empty completedAt while running, got %q", out.CompletedAt)
		}
		if out.Status != string(store.RunRunning) {
			t.Errorf("expected status %q, got %q", store.RunRunning, out.Status)
		}
	})
}

func TestIndentBlock(t *testing.T) {
	t.Run("aligns continuation lines under the label column", func(t *testing.T) {
		got := indentBlock("verb: run\nexec:\n  cmd: echo hi\n", 4)
		want := "verb: run\n    exec:\n      cmd: echo hi"
		if got != want {
			t.Errorf("expected %q, got %q", want, got)
		}
	})

	t.Run("empty in, empty out", func(t *testing.T) {
		// printRecordMetadata keys "should I print this field at all?" off the result.
		if got := indentBlock("", 12); got != "" {
			t.Errorf("expected an empty string, got %q", got)
		}
	})
}
