// Package response provides a structured success-response helper that is
// symmetric with the error envelope in cmd/internal/errors.  When the active
// --output mode is json or yaml the helper marshals a SuccessEnvelope to
// stdout; otherwise it falls back to the plain-text logger.
package response

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
)

// SuccessEnvelope is the top-level JSON/YAML payload for structured success responses.
type SuccessEnvelope struct {
	Result ResultDetail `json:"result" yaml:"result"`
}

// ResultDetail is the structured body inside a SuccessEnvelope.
type ResultDetail struct {
	Message string         `json:"message"        yaml:"message"`
	Data    map[string]any `json:"data,omitempty" yaml:"data,omitempty"`
}

// HandleSuccess emits a structured success response when --output is json or
// yaml, and falls back to logger.Log().PlainTextSuccess for plain-text mode.
func HandleSuccess(ctx *context.Context, cmd *cobra.Command, message string, data map[string]any) {
	format := outputFormat(cmd)

	switch format {
	case "json", "yaml", "yml":
		writeEnvelope(ctx, format, message, data)
	default:
		logger.Log().PlainTextSuccess(message)
	}
}

func writeEnvelope(ctx *context.Context, format, message string, data map[string]any) {
	env := SuccessEnvelope{
		Result: ResultDetail{
			Message: message,
			Data:    data,
		},
	}

	var out io.Writer = os.Stdout
	if ctx != nil {
		if w := ctx.StdOut(); w != nil {
			out = w
		}
	}

	switch format {
	case "yaml", "yml":
		b, err := yaml.Marshal(env)
		if err != nil {
			fmt.Fprintln(out, message)
			return
		}
		_, _ = out.Write(b)
	default:
		b, err := json.Marshal(env)
		if err != nil {
			fmt.Fprintln(out, message)
			return
		}
		_, _ = out.Write(b)
		_, _ = out.Write([]byte("\n"))
	}
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
