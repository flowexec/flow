// Package main demonstrates basic usage of the Flow CLI extension API.
//
// This example shows how to:
// - Create a context
// - Build a root command with custom configuration
// - Register all Flow commands
// - Execute the CLI
package main

import (
	stdCtx "context"
	"fmt"

	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/cli"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

func main() {
	// Load Flow configuration
	cfg, err := filesystem.LoadConfig()
	if err != nil {
		panic(fmt.Errorf("user config load error: %w", err))
	}

	// Initialize logger
	loggerOpts := logger.InitOptions{
		StdOut:  io.Stdout,
		LogMode: cfg.DefaultLogMode,
		Theme:   io.Theme(cfg.Theme.String()),
	}
	logger.Init(loggerOpts)
	defer logger.Log().Flush()

	// Create context
	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
	defer ctx.Finalize()

	// Build root command with custom version
	rootCmd := cli.BuildRootCommand(ctx,
		cli.WithVersion("1.0.0-example"),
		cli.WithShort("Flow CLI - Basic Extension Example"),
	)

	// Register all Flow commands
	cli.RegisterAllCommands(ctx, rootCmd)

	// Execute
	if err := cli.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}
