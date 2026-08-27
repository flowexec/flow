package mcp

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// fakeFlowBinary writes a script that echoes to both streams and exits with the
// given code, then points FLOW_CLI_BINARY at it so FlowCLIExecutor runs it in
// place of the real CLI.
func fakeFlowBinary(t *testing.T, exitCode int) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in is not portable to Windows")
	}

	path := filepath.Join(t.TempDir(), "fake-flow")
	script := "#!/bin/sh\necho stdout-line\necho stderr-line >&2\nexit " +
		strings.TrimSpace(string(rune('0'+exitCode))) + "\n"
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	t.Setenv(cliBinaryEnvKey, path)
}

// A non-zero exit is a normal outcome for a flow command (a failing test, a
// script that raises). It must come back as output with a nil error, not as a
// panic: a panic here kills the tool handler's goroutine before it replies, and
// the MCP client then waits on a request that never gets a response.
func TestExecuteContext_NonZeroExitReturnsOutputNotPanic(t *testing.T) {
	fakeFlowBinary(t, 1)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ExecuteContext panicked on a non-zero exit: %v", r)
		}
	}()

	out, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), "anything")
	if err != nil {
		t.Errorf("err = %v, want nil for a non-zero exit", err)
	}
	if !strings.Contains(out, "stdout-line") || !strings.Contains(out, "stderr-line") {
		t.Errorf("output = %q, want both streams captured", out)
	}
}

func TestExecuteContext_SuccessReturnsOutput(t *testing.T) {
	fakeFlowBinary(t, 0)

	out, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), "anything")
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	if !strings.Contains(out, "stdout-line") {
		t.Errorf("output = %q, want stdout captured", out)
	}
}

// A missing binary is a real failure, not an exit status, so it must surface as
// an error rather than being swallowed alongside exit codes.
func TestExecuteContext_MissingBinaryReturnsError(t *testing.T) {
	t.Setenv(cliBinaryEnvKey, filepath.Join(t.TempDir(), "does-not-exist"))

	if _, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), "anything"); err == nil {
		t.Error("err = nil, want an error when the binary cannot be run")
	}
}
