package run

import (
	"context"
	"errors"
	"fmt"
	stdio "io"
	"os"
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
			stdOutWriter(logMode, logger, flattenedFields...),
			stdErrWriter(logMode, logger, flattenedFields...),
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

// RunFile executes a file in the current shell in a specific directory.
func RunFile(
	filename, dir string,
	envList []string,
	logMode io.LogMode,
	logger io.Logger,
	stdIn *os.File,
	logFields map[string]interface{},
) error {
	logger.Debugf("executing file (%s)", filepath.Join(dir, filename))

	ctx := context.Background()
	fullPath := filepath.Join(dir, filename)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist - %s", fullPath)
	}
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
			stdOutWriter(logMode, logger, flattenedFields...),
			stdErrWriter(logMode, logger, flattenedFields...),
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

func stdOutWriter(mode io.LogMode, logger io.Logger, logFields ...any) stdio.Writer {
	return io.StdOutWriter{LogFields: logFields, Logger: logger, LogMode: &mode}
}

func stdErrWriter(mode io.LogMode, logger io.Logger, logFields ...any) stdio.Writer {
	return io.StdErrWriter{LogFields: logFields, Logger: logger, LogMode: &mode}
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
