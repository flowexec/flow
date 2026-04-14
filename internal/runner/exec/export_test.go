package exec

import (
	"os"

	"github.com/flowexec/tuikit/io"
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
