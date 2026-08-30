package io_test

import (
	"os"
	"testing"

	flowIO "github.com/flowexec/flow/v2/internal/io"
)

func TestTTYAttached(t *testing.T) {
	regular, err := os.CreateTemp(t.TempDir(), "flow-tty-test")
	if err != nil {
		t.Fatalf("unable to create temp file: %v", err)
	}
	t.Cleanup(func() { _ = regular.Close() })

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		t.Fatalf("unable to create pipe: %v", err)
	}
	t.Cleanup(func() { _ = pipeR.Close(); _ = pipeW.Close() })

	cases := []struct {
		name string
		in   *os.File
		out  *os.File
		want bool
	}{
		{name: "nil input", in: nil, out: regular, want: false},
		{name: "nil output", in: regular, out: nil, want: false},
		{name: "regular files", in: regular, out: regular, want: false},
		{name: "pipes", in: pipeR, out: pipeW, want: false},
		{name: "redirected output only", in: pipeR, out: regular, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := flowIO.TTYAttached(tc.in, tc.out); got != tc.want {
				t.Errorf("TTYAttached() = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestTTYAttachedWithTerminal guards against the negative cases above being
// satisfied by a function that always returns false. It needs a controlling
// terminal, which CI and container runs do not have, so it skips there.
func TestTTYAttachedWithTerminal(t *testing.T) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		t.Skipf("no controlling terminal available: %v", err)
	}
	t.Cleanup(func() { _ = tty.Close() })

	if !flowIO.TTYAttached(tty, tty) {
		t.Error("TTYAttached() = false for /dev/tty, want true")
	}
}
