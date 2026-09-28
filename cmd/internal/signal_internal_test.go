//go:build !windows

package internal

import (
	stdCtx "context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/flowexec/flow/v2/pkg/context"
	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
)

func TestCancelOnTermSignal(t *testing.T) {
	t.Run("cancels the run and records it before propagating", func(t *testing.T) {
		base, cancel := stdCtx.WithCancel(stdCtx.Background())
		ctx := &context.Context{}
		ctx.SetContext(base, cancel)

		var recorded error
		stop := cancelOnTermSignal(ctx, func(err error) {
			if ctx.Err() != nil {
				t.Error("expected the run to be recorded before the context is cancelled")
			}
			recorded = err
		})
		if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
			t.Fatalf("unable to signal self: %v", err)
		}
		select {
		case <-ctx.Done():
		case <-time.After(5 * time.Second):
			t.Fatal("expected the context to be cancelled by SIGTERM")
		}

		err := stop()
		var cancelled flowErrors.RunCancelledError
		if !errors.As(err, &cancelled) {
			t.Fatalf("expected a RunCancelledError, got %v", err)
		}
		if !errors.As(recorded, &cancelled) {
			t.Fatalf("expected onCancel to receive a RunCancelledError, got %v", recorded)
		}
	})

	t.Run("reports nothing when no signal arrives", func(t *testing.T) {
		base, cancel := stdCtx.WithCancel(stdCtx.Background())
		defer cancel()
		ctx := &context.Context{}
		ctx.SetContext(base, cancel)

		stop := cancelOnTermSignal(ctx, func(error) { t.Error("onCancel must not run without a signal") })
		if err := stop(); err != nil {
			t.Fatalf("expected no cancellation, got %v", err)
		}
		if ctx.Err() != nil {
			t.Fatal("expected the context to stay live")
		}
	})
}
