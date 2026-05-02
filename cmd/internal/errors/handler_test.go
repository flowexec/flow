package errors_test

import (
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/pkg/context"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
)

type exitCall struct{ code int }

func newCmdWithOutput(format string) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	_, _ = flags.ToPflag(cmd, *flags.OutputFormatFlag, false)
	if format != "" {
		_ = cmd.Flags().Set(flags.OutputFormatFlag.Name, format)
	}
	return cmd
}

func newCtxWithStderr(t *testing.T) (*context.Context, *os.File) {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "stderr-*")
	if err != nil {
		t.Fatalf("create temp stderr: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
	ctx := &context.Context{}
	ctx.SetStdErr(f)
	return ctx, f
}

func readAll(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

func withExitCapture(t *testing.T) {
	t.Helper()
	orig := flowerrors.ExitFunc
	t.Cleanup(func() { flowerrors.ExitFunc = orig })
	flowerrors.ExitFunc = func(code int) {
		panic(&exitCall{code: code})
	}
}

func runHandleFatal(ctx *context.Context, cmd *cobra.Command, err error) (code int, recovered bool) {
	defer func() {
		if r := recover(); r != nil {
			if ec, ok := r.(*exitCall); ok {
				code = ec.code
				recovered = true
				return
			}
			panic(r)
		}
	}()
	errhandler.HandleFatal(ctx, cmd, err)
	return 0, false
}

func TestHandleFatal_JSONEnvelope(t *testing.T) {
	ctx, stderr := newCtxWithStderr(t)
	cmd := newCmdWithOutput("json")
	withExitCapture(t)

	code, recovered := runHandleFatal(ctx, cmd, flowerrors.NewExecutableNotFoundError("ws/ns:foo"))
	if !recovered {
		t.Fatalf("expected ExitFunc to be invoked")
	}
	if code != 1 {
		t.Fatalf("want exit code 1, got %d", code)
	}

	out := readAll(t, stderr.Name())
	var env flowerrors.Envelope
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("envelope not valid JSON: %v\npayload: %s", err, out)
	}
	if env.Error.Code != flowerrors.ErrCodeNotFound {
		t.Fatalf("want code=%s, got %s", flowerrors.ErrCodeNotFound, env.Error.Code)
	}
	if env.Error.Message == "" {
		t.Fatalf("expected non-empty message")
	}
}

func TestHandleFatal_YAMLEnvelope(t *testing.T) {
	ctx, stderr := newCtxWithStderr(t)
	cmd := newCmdWithOutput("yaml")
	withExitCapture(t)

	_, recovered := runHandleFatal(ctx, cmd, flowerrors.NewUsageError("bad flag combination"))
	if !recovered {
		t.Fatalf("expected ExitFunc to be invoked")
	}

	out := readAll(t, stderr.Name())
	var env flowerrors.Envelope
	if err := yaml.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("envelope not valid YAML: %v\npayload: %s", err, out)
	}
	if env.Error.Code != flowerrors.ErrCodeInvalidInput {
		t.Fatalf("want code=%s, got %s", flowerrors.ErrCodeInvalidInput, env.Error.Code)
	}
}

func TestHandleFatal_ExitCodeForUsage(t *testing.T) {
	ctx, _ := newCtxWithStderr(t)
	cmd := newCmdWithOutput("json")
	withExitCapture(t)

	code, recovered := runHandleFatal(ctx, cmd, flowerrors.NewUsageError("oops"))
	if !recovered {
		t.Fatalf("expected ExitFunc to be invoked")
	}
	if code != 2 {
		t.Fatalf("want exit code 2 for usage error, got %d", code)
	}
}

func TestHandleFatal_ClassifiesUnknownAsInternal(t *testing.T) {
	ctx, stderr := newCtxWithStderr(t)
	cmd := newCmdWithOutput("json")
	withExitCapture(t)

	_, recovered := runHandleFatal(ctx, cmd, errors.New("something exploded"))
	if !recovered {
		t.Fatalf("expected ExitFunc to be invoked")
	}

	var env flowerrors.Envelope
	if err := json.Unmarshal([]byte(readAll(t, stderr.Name())), &env); err != nil {
		t.Fatalf("envelope not valid JSON: %v", err)
	}
	if env.Error.Code != flowerrors.ErrCodeInternal {
		t.Fatalf("want INTERNAL_ERROR, got %s", env.Error.Code)
	}
}

func TestHandleFatal_ValidationDetails(t *testing.T) {
	ctx, stderr := newCtxWithStderr(t)
	cmd := newCmdWithOutput("json")
	withExitCapture(t)

	ve := flowerrors.NewValidationError("bad field", map[string]any{"field": "verb"})
	_, recovered := runHandleFatal(ctx, cmd, ve)
	if !recovered {
		t.Fatalf("expected ExitFunc to be invoked")
	}

	var env flowerrors.Envelope
	if err := json.Unmarshal([]byte(readAll(t, stderr.Name())), &env); err != nil {
		t.Fatalf("envelope not valid JSON: %v", err)
	}
	if env.Error.Code != flowerrors.ErrCodeValidationFailed {
		t.Fatalf("want VALIDATION_FAILED, got %s", env.Error.Code)
	}
	if got, ok := env.Error.Details["field"]; !ok || got != "verb" {
		t.Fatalf("want details.field=verb, got %v", env.Error.Details)
	}
}
