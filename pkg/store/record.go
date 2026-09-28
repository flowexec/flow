package store

import (
	"time"

	"github.com/flowexec/flow/v2/internal/utils/process"
)

// staleRunError is recorded on a run whose process died without writing its terminal state.
const staleRunError = "process exited unexpectedly"

// EffectiveStatus returns the record's lifecycle status as flow reports it: a "running" record whose
// process is gone is failed, and a legacy record without a status is derived from its exit code.
// Consumers reading the store directly should use this rather than the raw Status field.
func (r ExecutionRecord) EffectiveStatus() RunStatus {
	switch {
	case r.Status == RunRunning && r.isStale():
		return RunFailed
	case r.Status != "":
		return r.Status
	case r.ExitCode == 0:
		return RunCompleted
	default:
		return RunFailed
	}
}

// Reconciled returns the record with a stale "running" state resolved to failed, closing it out
// at now with the elapsed duration. The bool reports whether anything changed, i.e. whether the
// corrected record should be written back.
func (r ExecutionRecord) Reconciled(now time.Time) (ExecutionRecord, bool) {
	if r.Status != RunRunning || !r.isStale() {
		return r, false
	}
	r.Status = RunFailed
	r.ExitCode = 1
	r.CompletedAt = &now
	if r.Duration == 0 && !r.StartedAt.IsZero() {
		r.Duration = now.Sub(r.StartedAt)
	}
	if r.Error == "" {
		r.Error = staleRunError
	}
	return r, true
}

func (r ExecutionRecord) isStale() bool {
	return r.PID != 0 && !process.Alive(r.PID)
}
