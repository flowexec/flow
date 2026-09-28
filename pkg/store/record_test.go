package store_test

import (
	"os"
	"testing"
	"time"

	"github.com/flowexec/flow/v2/pkg/store"
)

// deadPID is implausibly high, so no live process owns it.
const deadPID = 2_000_000_000

func TestEffectiveStatus(t *testing.T) {
	cases := map[string]struct {
		record store.ExecutionRecord
		want   store.RunStatus
	}{
		"running with live process":  {store.ExecutionRecord{Status: store.RunRunning, PID: os.Getpid()}, store.RunRunning},
		"running with dead process":  {store.ExecutionRecord{Status: store.RunRunning, PID: deadPID}, store.RunFailed},
		"running without a pid":      {store.ExecutionRecord{Status: store.RunRunning}, store.RunRunning},
		"cancelled":                  {store.ExecutionRecord{Status: store.RunCancelled, ExitCode: 1}, store.RunCancelled},
		"legacy record exit zero":    {store.ExecutionRecord{ExitCode: 0}, store.RunCompleted},
		"legacy record exit nonzero": {store.ExecutionRecord{ExitCode: 2}, store.RunFailed},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := tc.record.EffectiveStatus(); got != tc.want {
				t.Errorf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestReconciled(t *testing.T) {
	start := time.Now().Add(-90 * time.Second)
	now := start.Add(90 * time.Second)

	t.Run("closes out a stale running record with its elapsed duration", func(t *testing.T) {
		r := store.ExecutionRecord{Status: store.RunRunning, PID: deadPID, StartedAt: start}
		got, changed := r.Reconciled(now)
		if !changed {
			t.Fatal("expected the stale record to change")
		}
		if got.Status != store.RunFailed || got.ExitCode != 1 || got.Error == "" {
			t.Errorf("expected a failed record with an error, got %+v", got)
		}
		if got.Duration != 90*time.Second {
			t.Errorf("expected duration 90s, got %v", got.Duration)
		}
		if got.CompletedAt == nil || !got.CompletedAt.Equal(now) {
			t.Errorf("expected completedAt %v, got %v", now, got.CompletedAt)
		}
	})

	t.Run("leaves live and finished records alone", func(t *testing.T) {
		for _, r := range []store.ExecutionRecord{
			{Status: store.RunRunning, PID: os.Getpid(), StartedAt: start},
			{Status: store.RunCancelled, PID: deadPID, StartedAt: start},
		} {
			if got, changed := r.Reconciled(now); changed || got.Status != r.Status {
				t.Errorf("expected %q record unchanged, got %+v", r.Status, got)
			}
		}
	})
}
