package main

import (
	stdCtx "context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/google/uuid"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/cli"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/pkg/store"
	"github.com/flowexec/flow/types/executable"
)

func main() {
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("user config load error: %w", err))
	}

	archiveDir, archiveID := initLogArchive()
	loggerOpts := logger.InitOptions{
		StdOut:           io.Stdout,
		LogMode:          cfg.DefaultLogMode,
		Theme:            logger.Theme(cfg.Theme.String()),
		ArchiveDirectory: archiveDir,
		ArchiveID:        archiveID,
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

	initStore()

	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, context.WithStdIn(io.Stdin), context.WithStdOut(io.Stdout))
	ctx.LogArchiveID = archiveID
	defer ctx.Finalize()

	if ctx == nil {
		panic("failed to initialize context")
	}
	rootCmd := cli.BuildRootCommand(ctx)
	cli.RegisterAllCommands(ctx, rootCmd)
	if err := cmd.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}

func initLogArchive() (dir, id string) {
	if args := os.Args; len(args) > 1 && slices.Contains(executable.ValidVerbs(), executable.Verb(args[1])) {
		dir = filesystem.LogsDir()
		id = uuid.New().String()
	}
	return
}

func initStore() {
	if err := store.MigrateProcessBuckets(""); err != nil {
		logger.Log().Debug("process bucket migration skipped or failed", "err", err)
	}
}
