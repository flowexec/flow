// Package errors provides the central CLI error handler that emits a
// structured JSON/YAML envelope to stderr when --output=json|yaml, and falls
// back to the existing plain-text logger for text/TUI mode.
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/pkg/context"
	flowerrors "github.com/flowexec/flow/pkg/errors"
	"github.com/flowexec/flow/pkg/logger"
)

func ExitFunc(code int) { flowerrors.ExitFunc(code) }

// HandleFatal classifies err, emits a structured envelope to stderr when the
// active --output mode is json or yaml, and terminates the process.
func HandleFatal(ctx *context.Context, cmd *cobra.Command, err error) {
	if err == nil {
		return
	}

	format := outputFormat(cmd)

	switch format {
	case "json", "yaml", "yml":
		writeEnvelope(ctx, format, err)
		flowerrors.ExitFunc(exitCodeFor(err))
		return
	default:
		logger.Log().FatalErr(err)
	}
}

// HandleUsage is a convenience wrapper for invalid flag/argument conditions.
// It formats the message as a UsageError and routes it through HandleFatal.
func HandleUsage(ctx *context.Context, cmd *cobra.Command, format string, args ...any) {
	HandleFatal(ctx, cmd, flowerrors.NewUsageError(format, args...))
}

func writeEnvelope(ctx *context.Context, format string, err error) {
	env := flowerrors.NewEnvelope(codeFor(err), err.Error(), detailsFor(err))

	var out io.Writer = os.Stderr
	if ctx != nil {
		if w := ctx.StdErr(); w != nil {
			out = w
		}
	}

	switch format {
	case "yaml", "yml":
		data, mErr := yaml.Marshal(env)
		if mErr != nil {
			fmt.Fprintln(out, err.Error())
			return
		}
		_, _ = out.Write(data)
	default:
		data, mErr := json.Marshal(env)
		if mErr != nil {
			fmt.Fprintln(out, err.Error())
			return
		}
		_, _ = out.Write(data)
		_, _ = out.Write([]byte("\n"))
	}
}

func codeFor(err error) string {
	var c flowerrors.Coder
	if errors.As(err, &c) {
		return c.Code()
	}
	return flowerrors.ErrCodeInternal
}

func detailsFor(err error) map[string]any {
	var ve flowerrors.ValidationError
	if errors.As(err, &ve) && len(ve.Details) > 0 {
		return ve.Details
	}
	return nil
}

func exitCodeFor(err error) int {
	if codeFor(err) == flowerrors.ErrCodeInvalidInput {
		return 2
	}
	return 1
}

// outputFormat reads --output from cmd (or its root), defensively handling
// commands that don't register the flag.
func outputFormat(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	for _, c := range []*cobra.Command{cmd, cmd.Root()} {
		if c == nil {
			continue
		}
		if flags.HasFlag(c, *flags.OutputFormatFlag) {
			if v := flags.ValueFor[string](c, *flags.OutputFormatFlag, false); v != "" {
				return v
			}
		}
	}
	return ""
}
