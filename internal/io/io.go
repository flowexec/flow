package io

import (
	"os"

	"golang.org/x/term"
)

const DisableInteractiveEnvKey = "DISABLE_FLOW_INTERACTIVE"

var (
	Stdout = os.Stdout
	Stdin  = os.Stdin
)

// TTYAttached reports whether both streams are real terminals.
//
// The TUI reads key events from stdin and paints stdout, so a pipe on either end
// leaves it unusable: it writes escape sequences into whatever is consuming the
// output and then blocks until the container readiness timeout expires.
func TTYAttached(in, out *os.File) bool {
	if in == nil || out == nil {
		return false
	}
	return term.IsTerminal(int(in.Fd())) && term.IsTerminal(int(out.Fd()))
}
