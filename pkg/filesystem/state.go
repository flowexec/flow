package filesystem

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/pkg/errors"
)

const FlowStateDirEnvVar = "FLOW_STATE_DIR"

// StateDirPath returns the platform-appropriate directory for persistent state
// data (logs, history) that is not configuration and should not be treated as
// disposable cache.
//
// Platform defaults:
//   - macOS:   ~/Library/Logs/flow
//   - Linux:   $XDG_STATE_HOME/flow (defaults to ~/.local/state/flow)
//   - Windows: %LOCALAPPDATA%\flow
func StateDirPath() string {
	if dir := os.Getenv(FlowStateDirEnvVar); dir != "" {
		return dir
	}

	switch runtime.GOOS {
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			panic(errors.Wrap(err, "unable to get home directory"))
		}
		return filepath.Join(home, "Library", "Logs", dataDirName)
	case "windows":
		// %LOCALAPPDATA% is the standard location for app state on Windows.
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			panic("LOCALAPPDATA is not set")
		}
		return filepath.Join(dir, dataDirName)
	default: // linux and other unix
		if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
			return filepath.Join(dir, dataDirName)
		}
		home, err := os.UserHomeDir()
		if err != nil {
			panic(errors.Wrap(err, "unable to get home directory"))
		}
		return filepath.Join(home, ".local", "state", dataDirName)
	}
}
