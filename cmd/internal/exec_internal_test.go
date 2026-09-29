package internal

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"mvdan.cc/sh/v3/interp"
)

func TestBackgroundChildArgs_StripsBackgroundFlag(t *testing.T) {
	cases := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "long form among ad-hoc flags",
			in:   []string{"exec", "--cmd", "echo hi", "--background", "--label", "x"},
			want: []string{"exec", "--cmd", "echo hi", "--label", "x"},
		},
		{
			name: "short form with a named executable",
			in:   []string{"run", "build", "-b"},
			want: []string{"run", "build"},
		},
		{
			name: "explicit value form",
			in:   []string{"exec", "--cmd", "echo hi", "--background=true"},
			want: []string{"exec", "--cmd", "echo hi"},
		},
		{
			name: "no background flag is a passthrough",
			in:   []string{"run", "build", "--", "--arg"},
			want: []string{"run", "build", "--", "--arg"},
		},
		{
			name: "preserves args after a -- separator",
			in:   []string{"run", "test", "--background", "--", "-v", "--run", "TestX"},
			want: []string{"run", "test", "--", "-v", "--run", "TestX"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := backgroundChildArgs(tc.in)
			if !slices.Equal(got, tc.want) {
				t.Fatalf("backgroundChildArgs(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestTransientSpecName(t *testing.T) {
	cases := map[string]string{
		"":              "spec",
		"   ":           "spec",
		"Deploy API":    "spec-deploy-api",
		"my-run":        "spec-my-run",
		"!!!":           "spec",
		"nightly_build": "spec-nightly-build",
	}
	for label, want := range cases {
		if got := transientSpecName(label); got != want {
			t.Fatalf("transientSpecName(%q) = %q, want %q", label, got, want)
		}
	}
}

func TestExitCode_ReportsTheCommandsOwnStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "wrapped shell exit status",
			err:  fmt.Errorf("command exited with non-zero status %w", interp.ExitStatus(2)),
			want: 2,
		},
		{
			name: "nested in a serial step",
			err:  fmt.Errorf("serial execution failed: %w", fmt.Errorf("step: %w", interp.ExitStatus(127))),
			want: 127,
		},
		{name: "any other failure", err: errors.New("unable to create runner"), want: 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := exitCode(tc.err); got != tc.want {
				t.Errorf("exitCode() = %d, want %d", got, tc.want)
			}
		})
	}
}
