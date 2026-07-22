package exec

import (
	stdctx "context"
	"os"

	"github.com/flowexec/tuikit/io"

	"github.com/flowexec/flow/v2/internal/services/run"
)

// RunFunc matches the signature of run.RunCmd / run.RunFile so tests can
// substitute either seam without importing the run package.
type RunFunc = func(
	s, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]any,
	task *io.TaskContext,
) error

// SetRunCmdFnForTest swaps the command-runner seam and returns a restore func.
func SetRunCmdFnForTest(fn RunFunc) func() {
	prev := runCmdFn
	runCmdFn = fn
	return func() { runCmdFn = prev }
}

// SetRunFileFnForTest swaps the file-runner seam and returns a restore func.
func SetRunFileFnForTest(fn RunFunc) func() {
	prev := runFileFn
	runFileFn = fn
	return func() { runFileFn = prev }
}

// ContainerRunFunc matches the signature of run.RunContainer.
type ContainerRunFunc = func(
	ctx stdctx.Context,
	spec run.ContainerSpec,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]any,
	task *io.TaskContext,
) error

// SetRunContainerFnForTest swaps the container-runner seam and returns a restore func.
func SetRunContainerFnForTest(fn ContainerRunFunc) func() {
	prev := runContainerFn
	runContainerFn = fn
	return func() { runContainerFn = prev }
}

// SetLookPathForContainerTest forces container runtime resolution to succeed with
// "docker" so container tests do not require a real runtime on PATH. Returns a
// restore func.
func SetLookPathForContainerTest() func() {
	prev := resolveRuntimeFn
	resolveRuntimeFn = func(string) (string, error) { return "docker", nil }
	return func() { resolveRuntimeFn = prev }
}

// ExpandVolumeHostForTest exposes expandVolumeHost.
func ExpandVolumeHostForTest(host, wsRoot string) (string, error) {
	return expandVolumeHost(host, wsRoot)
}
