package run

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flowexec/tuikit/io"

	flowErrors "github.com/flowexec/flow/v2/pkg/errors"
)

const (
	// PythonBinEnv overrides interpreter discovery entirely. It is read from the
	// flow-resolved environment before the process environment, so it can be set
	// per-workspace in a .env file or per-executable via params.
	PythonBinEnv = "FLOW_PYTHON_BIN"

	// virtualEnvEnv is the variable an activated virtualenv exports.
	virtualEnvEnv = "VIRTUAL_ENV"

	// workspacePathEnv is set by env.DefaultEnv for every run; it is how the
	// workspace root reaches this package without widening the run signatures.
	workspacePathEnv = "FLOW_WORKSPACE_PATH"

	// conventionalVenvDir is the in-repo virtualenv location flow looks for when
	// none is active.
	conventionalVenvDir = ".venv"
)

// ResolvePython locates the python interpreter to run with, preferring a project's
// virtualenv over bare system python so that an agent running python inside a repo
// gets that repo's dependencies.
//
// Order: FLOW_PYTHON_BIN, the active VIRTUAL_ENV, the workspace's .venv, then
// python3 and python on the PATH.
func ResolvePython(envList []string) (string, error) {
	if override := envValue(envList, PythonBinEnv); override != "" {
		// An override may be a bare name or a path; resolve the former on PATH and
		// trust the latter, so a deliberately-chosen interpreter never silently
		// falls through to a different one.
		if strings.ContainsRune(override, os.PathSeparator) || strings.ContainsRune(override, '/') {
			if isExecutableFile(override) {
				return override, nil
			}
		} else if path, err := lookPath(override); err == nil {
			return path, nil
		}
		return "", flowErrors.NewInterpreterNotFoundError("python", []string{PythonBinEnv + "=" + override})
	}

	searched := []string{PythonBinEnv}

	if venv := envValue(envList, virtualEnvEnv); venv != "" {
		candidate := venvPython(venv)
		searched = append(searched, candidate)
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	if wsRoot := envValue(envList, workspacePathEnv); wsRoot != "" {
		candidate := venvPython(filepath.Join(wsRoot, conventionalVenvDir))
		searched = append(searched, candidate)
		if isExecutableFile(candidate) {
			return candidate, nil
		}
	}

	for _, name := range pathCandidates() {
		searched = append(searched, name)
		if path, err := lookPath(name); err == nil {
			return path, nil
		}
	}

	return "", flowErrors.NewInterpreterNotFoundError("python", searched)
}

// pathCandidates returns the bare interpreter names to try on the PATH, in order.
//
// Windows deliberately tries "python" first: "python3.exe" there is usually the
// Microsoft Store App Execution Alias, which is on the PATH even when Python is
// not installed and merely prints a store advertisement before exiting non-zero.
// Preferring "python" finds a real installation and only falls back to the alias.
func pathCandidates() []string {
	if runtime.GOOS == "windows" {
		return []string{"python", "python3"}
	}
	return []string{"python3", "python"}
}

// venvPython returns the interpreter path inside a virtualenv directory.
func venvPython(venvDir string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(venvDir, "Scripts", "python.exe")
	}
	return filepath.Join(venvDir, "bin", "python")
}

func isExecutableFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// envValue reads a key from a KEY=VALUE list, falling back to the process
// environment. The list is scanned in reverse because later entries win, matching
// how RunCmd layers the flow environment over os.Environ.
func envValue(envList []string, key string) string {
	prefix := key + "="
	for i := len(envList) - 1; i >= 0; i-- {
		if strings.HasPrefix(envList[i], prefix) {
			return strings.TrimPrefix(envList[i], prefix)
		}
	}
	return os.Getenv(key)
}

// RunPythonCmd executes python source in a specific directory.
//
// The code is written to a 0600 temp file rather than passed with `python -c`:
// that keeps user code (which may interpolate secrets) out of the process table,
// produces tracebacks with real line numbers, and sidesteps shell quoting for
// multi-line scripts.
func RunPythonCmd(
	code, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	pythonBin, err := ResolvePython(envList)
	if err != nil {
		return err
	}

	scriptPath, cleanup, err := WritePythonScript(code)
	if err != nil {
		return err
	}
	defer cleanup()

	logger.Debugf("running python (%s) in dir (%s)", pythonBin, dir)
	return runNativeFile(
		pythonBin, []string{scriptPath}, dir,
		pythonEnv(envList), logMode, logger, stdIn, logFields, task,
	)
}

// RunPythonFile executes an existing .py file with the resolved interpreter.
func RunPythonFile(
	fullPath, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist - %s", fullPath)
	}

	pythonBin, err := ResolvePython(envList)
	if err != nil {
		return err
	}

	logger.Debugf("executing python file (%s) with %s", fullPath, pythonBin)
	return runNativeFile(
		pythonBin, []string{fullPath}, dir,
		pythonEnv(envList), logMode, logger, stdIn, logFields, task,
	)
}

// WritePythonScript writes python source to a 0600 temp file and returns its path
// alongside a cleanup func. Exported so callers that need the script to outlive
// the call - notably container execution, which bind-mounts it - can manage it.
func WritePythonScript(code string) (path string, cleanup func(), err error) {
	file, err := os.CreateTemp("", "flow-*.py")
	if err != nil {
		return "", func() {}, fmt.Errorf("unable to create python script file - %w", err)
	}
	name := file.Name()
	remove := func() { _ = os.Remove(name) }

	if _, err := file.WriteString(code); err != nil {
		_ = file.Close()
		remove()
		return "", func() {}, fmt.Errorf("unable to write python script file - %w", err)
	}
	if err := file.Close(); err != nil {
		remove()
		return "", func() {}, fmt.Errorf("unable to close python script file - %w", err)
	}
	if err := os.Chmod(name, 0600); err != nil {
		remove()
		return "", func() {}, fmt.Errorf("unable to set python script permissions - %w", err)
	}
	return name, remove, nil
}

// pythonEnv layers python-specific defaults over the resolved environment.
//
// PYTHONUNBUFFERED: flow pipes stdout to a log writer rather than a tty, and
// CPython block-buffers to a pipe - without this a long run emits nothing until
// it exits, so it looks hung to anyone (or any agent) watching the output.
//
// PYTHONDONTWRITEBYTECODE: `file:` execution runs from the workspace, and flow
// should not litter a user's repository with __pycache__ directories.
//
// Both stay overridable: an explicit value in the run's environment wins.
func pythonEnv(envList []string) []string {
	out := append([]string{}, envList...)
	for _, kv := range [][2]string{
		{"PYTHONUNBUFFERED", "1"},
		{"PYTHONDONTWRITEBYTECODE", "1"},
	} {
		if envValue(envList, kv[0]) == "" {
			out = append(out, kv[0]+"="+kv[1])
		}
	}
	return out
}
