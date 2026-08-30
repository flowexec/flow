package internal

import (
	"os"
	"testing"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/v2/cmd/internal/flags"
	flowIO "github.com/flowexec/flow/v2/internal/io"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/types/config"
)

// nonTTYContext builds a context whose config asks for the TUI but whose streams
// are regular files — the shape of every piped, redirected, or agent-driven run.
func nonTTYContext(t *testing.T) *context.Context {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "flow-tui-test")
	if err != nil {
		t.Fatalf("unable to create temp file: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })

	ctx := &context.Context{
		Config: &config.Config{Interactive: &config.Interactive{Enabled: true}},
	}
	ctx.SetIO(f, f)
	return ctx
}

func TestTUIEnabled_NonTTY(t *testing.T) {
	cases := []struct {
		name   string
		format string
	}{
		{name: "no output flag set", format: ""},
		{name: "explicit tui does not override", format: "tui"},
		{name: "explicit json", format: "json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(flowIO.DisableInteractiveEnvKey, "")
			ctx := nonTTYContext(t)

			cmd := &cobra.Command{Use: "test"}
			RegisterFlag(ctx, cmd, *flags.OutputFormatFlag)
			if tc.format != "" {
				if err := cmd.Flags().Set(flags.OutputFormatFlag.Name, tc.format); err != nil {
					t.Fatalf("unable to set output flag: %v", err)
				}
			}

			if TUIEnabled(ctx, cmd) {
				t.Error("TUIEnabled() = true without a terminal, want false")
			}
		})
	}
}
