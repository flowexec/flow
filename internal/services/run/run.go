package run

import (
	"context"
	"errors"
	"fmt"
	stdio "io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"strings"

	"github.com/flowexec/tuikit/io"
	"mvdan.cc/sh/v3/expand"
	"mvdan.cc/sh/v3/interp"
	"mvdan.cc/sh/v3/syntax"
)

func init() {
	setupColorEnvironment()
}

// RunCmd executes a command in the current shell in a specific directory.
func RunCmd(
	commandStr, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	logger.Debugf("running command in dir (%s):\n%s", dir, strings.TrimSpace(commandStr))

	ctx := context.Background()
	parser := syntax.NewParser()
	reader := strings.NewReader(strings.TrimSpace(commandStr))
	prog, err := parser.Parse(reader, "")
	if err != nil {
		return fmt.Errorf("unable to parse command - %w", err)
	}

	if envList == nil {
		envList = make([]string, 0)
	}
	envList = append(os.Environ(), envList...)

	flattenedFields := make([]interface{}, 0)
	for k, v := range logFields {
		flattenedFields = append(flattenedFields, k, v)
	}
	runner, err := interp.New(
		interp.Dir(dir),
		interp.Env(expand.ListEnviron(envList...)),
		interp.StdIO(
			stdIn,
			stdOutWriter(logMode, logger, task, flattenedFields...),
			stdErrWriter(logMode, logger, task, flattenedFields...),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to create runner - %w", err)
	}

	err = runner.Run(ctx, prog)
	if err != nil {
		var exitStatus interp.ExitStatus
		if errors.As(err, &exitStatus) {
			return fmt.Errorf("command exited with non-zero status %w", exitStatus)
		}
		return fmt.Errorf("encountered an error executing command - %w", err)
	}

	return nil
}

// RunFile executes a file in a specific directory.
// Shell scripts (.sh) are interpreted via the built-in POSIX shell interpreter.
// Batch files (.bat, .cmd) are executed via cmd.exe and PowerShell scripts (.ps1) via pwsh/powershell.
func RunFile(
	filename, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	logger.Debugf("executing file (%s)", filepath.Join(dir, filename))

	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist - %s", fullPath)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".bat", ".cmd":
		return runNativeFile("cmd", []string{"/C", fullPath}, dir, envList, logMode, logger, stdIn, logFields, task)
	case ".ps1":
		shell := findPowerShell()
		return runNativeFile(shell, []string{"-NoProfile", "-ExecutionPolicy", "Bypass", "-File", fullPath},
			dir, envList, logMode, logger, stdIn, logFields, task)
	default:
		return runShellFile(fullPath, envList, logMode, logger, stdIn, logFields, task)
	}
}

// findPowerShell returns the PowerShell executable name,
// preferring the cross-platform "pwsh" over the Windows-only "powershell".
func findPowerShell() string {
	if _, err := osexec.LookPath("pwsh"); err == nil {
		return "pwsh"
	}
	return "powershell"
}

// runNativeFile executes a file using a native system command (e.g. cmd.exe, pwsh).
func runNativeFile(
	command string, args []string,
	dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	if envList == nil {
		envList = make([]string, 0)
	}
	envList = append(os.Environ(), envList...)

	flattenedFields := make([]interface{}, 0)
	for k, v := range logFields {
		flattenedFields = append(flattenedFields, k, v)
	}

	cmd := osexec.Command(command, args...)
	cmd.Dir = dir
	cmd.Env = envList
	cmd.Stdin = stdIn
	cmd.Stdout = stdOutWriter(logMode, logger, task, flattenedFields...)
	cmd.Stderr = stdErrWriter(logMode, logger, task, flattenedFields...)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("file execution failed - %w", err)
	}
	return nil
}

// runShellFile executes a file using the built-in POSIX shell interpreter.
func runShellFile(
	fullPath string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
	task *io.TaskContext,
) error {
	ctx := context.Background()
	file, err := os.OpenFile(filepath.Clean(fullPath), os.O_RDONLY, 0)
	if err != nil {
		return fmt.Errorf("unable to open file - %w", err)
	}
	defer file.Close()

	parser := syntax.NewParser()
	prog, err := parser.Parse(file, "")
	if err != nil {
		return fmt.Errorf("unable to parse file - %w", err)
	}

	if envList == nil {
		envList = make([]string, 0)
	}
	envList = append(os.Environ(), envList...)

	flattenedFields := make([]interface{}, 0)
	for k, v := range logFields {
		flattenedFields = append(flattenedFields, k, v)
	}
	runner, err := interp.New(
		interp.Env(expand.ListEnviron(envList...)),
		interp.StdIO(
			stdIn,
			stdOutWriter(logMode, logger, task, flattenedFields...),
			stdErrWriter(logMode, logger, task, flattenedFields...),
		),
	)
	if err != nil {
		return fmt.Errorf("unable to create runner - %w", err)
	}

	err = runner.Run(ctx, prog)
	if err != nil {
		var exitStatus interp.ExitStatus
		if errors.As(err, &exitStatus) {
			return fmt.Errorf("file execution exited with non-zero status %w", exitStatus)
		}
		return fmt.Errorf("encountered an error executing file - %w", err)
	}
	return nil
}

func stdOutWriter(mode io.LogMode, logger io.Logger, task *io.TaskContext, logFields ...any) stdio.Writer {
	return io.StdOutWriter{LogFields: logFields, Logger: logger, LogMode: &mode, Task: task}
}

func stdErrWriter(mode io.LogMode, logger io.Logger, task *io.TaskContext, logFields ...any) stdio.Writer {
	return io.StdErrWriter{LogFields: logFields, Logger: logger, LogMode: &mode, Task: task}
}

func setupColorEnvironment() {
	hasColorPreference := os.Getenv("NO_COLOR") != "" ||
		os.Getenv("FORCE_COLOR") != "" ||
		os.Getenv("CLICOLOR_FORCE") != ""

	if !hasColorPreference {
		isCI := os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != ""
		isTesting := strings.HasSuffix(os.Args[0], ".test")

		if isCI || isTesting {
			_ = os.Setenv("NO_COLOR", "1")
		} else {
			_ = os.Setenv("FORCE_COLOR", "1")
			_ = os.Setenv("CLICOLOR_FORCE", "1")
		}
	}

	// Always ensure TERM is set if colors might be used
	if os.Getenv("NO_COLOR") == "" && os.Getenv("TERM") == "" {
		_ = os.Setenv("TERM", "xterm-256color")
	}
}
