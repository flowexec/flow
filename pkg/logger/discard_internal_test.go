package logger

import (
	"os"
	"testing"
)

// TestDiscardOutputIsNotStandardInput pins the exact defect: the fallback sink used to be
// os.NewFile(0, os.DevNull), which adopts descriptor 0 rather than opening anything.
//
// In-package because the descriptor is the whole assertion, and it is not reachable from
// outside. An external test can only observe the consequence — a descriptor closed by a
// finalizer at an arbitrary later moment — which is exactly what made this so hard to
// trace the first time.
func TestDiscardOutputIsNotStandardInput(t *testing.T) {
	sink := discardOutput()

	if sink == nil {
		t.Fatal("discardOutput returned nil")
	}
	if sink.Fd() == 0 {
		t.Error("the fallback log sink is standard input; it must open the null device instead")
	}
}

// TestDiscardOutputIsReused guards the second half of the fix. One file per process means
// one finalizer; a file per call leaves a finalizer for every call, each waiting to close
// whichever descriptor number it ends up holding.
func TestDiscardOutputIsReused(t *testing.T) {
	first := discardOutput()
	for range 50 {
		if discardOutput() != first {
			t.Fatal("discardOutput allocated a new file; every call must share one")
		}
	}
}

// TestDiscardOutputIsWritable confirms the sink is usable, so switching away from
// standard input did not simply move the silence somewhere else.
func TestDiscardOutputIsWritable(t *testing.T) {
	if discardOutput() == os.Stdout {
		t.Skip("null device unavailable; the fallback is stdout by design")
	}
	if _, err := discardOutput().WriteString("discarded\n"); err != nil {
		t.Errorf("the fallback sink is not writable: %v", err)
	}
}
