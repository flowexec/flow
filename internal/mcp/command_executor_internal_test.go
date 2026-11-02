package mcp

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/flowexec/flow/v2/pkg/store"
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

// A non-zero exit must come back as the captured output plus an error handlers can
// act on, not as a nil error (the failure goes unreported) and not as a panic (the
// tool handler's goroutine dies and the MCP client waits on a reply that never comes).
func TestExecuteContext_NonZeroExitReturnsOutputAndExitError(t *testing.T) {
	fakeFlowBinary(t, 3)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ExecuteContext panicked on a non-zero exit: %v", r)
		}
	}()

	out, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), "anything")
	var exitErr *CLIExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("err = %v, want *CLIExitError", err)
	}
	if exitErr.Code != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.Code)
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

func TestExecuteContext_DisablesColor(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in is not portable to Windows")
	}
	path := filepath.Join(t.TempDir(), "fake-flow")
	script := "#!/bin/sh\necho \"no=$NO_COLOR clicolor=${CLICOLOR_FORCE-unset} force=${FORCE_COLOR-unset} " +
		"tty=${TTY_FORCE-unset} other=$KEEP_ME\"\n"
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatalf("writing fake binary: %v", err)
	}
	t.Setenv(cliBinaryEnvKey, path)
	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("FORCE_COLOR", "1")
	t.Setenv("TTY_FORCE", "1")
	t.Setenv("KEEP_ME", "kept")

	out, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), "anything")
	if err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
	want := "no=1 clicolor=unset force=unset tty=unset other=kept"
	if strings.TrimSpace(out) != want {
		t.Errorf("output = %q, want %q", out, want)
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

func TestExecuteContext_ReportsWhetherARunStarted(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in is not portable to Windows")
	}
	script := "#!/bin/sh\n[ \"$1\" = started ] && : > \"$" + store.RunStartedFileEnv + "\"\nexit 7\n"
	path := filepath.Join(t.TempDir(), "fake-flow")
	if err := os.WriteFile(path, []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv(cliBinaryEnvKey, path)

	for arg, wantStarted := range map[string]bool{"started": true, "rejected": false} {
		_, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background(), arg)
		var exitErr *CLIExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("%s: got %v, want a CLIExitError", arg, err)
		}
		if exitErr.Code != 7 || exitErr.Started != wantStarted {
			t.Errorf("%s: got %+v, want code 7 and started %v", arg, *exitErr, wantStarted)
		}
	}
}

func TestExecuteContext_RunsWithoutATempDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell script stand-in is not portable to Windows")
	}
	path := filepath.Join(t.TempDir(), "fake-flow")
	if err := os.WriteFile(path, []byte("#!/bin/sh\necho ran\nexit 3\n"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv(cliBinaryEnvKey, path)
	t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))

	out, err := (&FlowCLIExecutor{}).ExecuteContext(context.Background())
	var exitErr *CLIExitError
	if !errors.As(err, &exitErr) || !exitErr.Started || !strings.Contains(out, "ran") {
		t.Errorf("got output %q and error %v, want the run to happen and report started", out, err)
	}
}
