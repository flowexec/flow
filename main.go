package main

import (
	stdCtx "context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

func main() {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("user config load error: %w", err))
	}

	var archiveDir string
	if args := os.Args; len(args) > 1 && slices.Contains(executable.ValidVerbs(), executable.Verb(args[1])) {
		// only create a log archive file for exec commands
		archiveDir = filesystem.LogsDir()
	}
	loggerOpts := logger.InitOptions{
		StdOut:           io.Stdout,
		LogMode:          cfg.DefaultLogMode,
		Theme:            io.Theme(cfg.Theme.String()),
		ArchiveDirectory: archiveDir,
	}
	logger.Init(loggerOpts)
	defer func() {
		if err := logger.Log().Flush(); err != nil {
			if errors.Is(err, os.ErrClosed) {
				return
			}
			panic(err)
		}
	}()

	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
	defer ctx.Finalize()

	if ctx == nil {
		panic("failed to initialize context")
	}
	rootCmd := cmd.NewRootCmd(ctx)
	if err := cmd.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}
