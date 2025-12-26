// Package main demonstrates hook injection using the Flow CLI extension API.
//
// This example shows how to:
// - Add telemetry/logging hooks to all commands
// - Use PreRun and PostRun hooks
// - Apply hooks recursively
package main

import (
	stdCtx "context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/pkg/cli"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

// commandTiming stores timing information for commands
var commandTiming = make(map[string]time.Time)

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

	// Build root command
	rootCmd := cli.BuildRootCommand(ctx,
		cli.WithVersion("1.0.0-hooks-example"),
		cli.WithShort("Flow CLI - Hook Injection Example"),
	)

	// Register all Flow commands
	cli.RegisterAllCommands(ctx, rootCmd)

	// Add telemetry hooks to all commands
	cli.ApplyHooksRecursive(rootCmd,
		// PreRun: Start timing and log command execution
		func(cmd *cobra.Command, args []string) {
			commandTiming[cmd.Name()] = time.Now()
			logger.Log().Infof("Starting command: %s", cmd.Name())
			if len(args) > 0 {
				logger.Log().Debugf("Arguments: %v", args)
			}
		},
		// PostRun: Log completion and timing
		func(cmd *cobra.Command, args []string) {
			if startTime, ok := commandTiming[cmd.Name()]; ok {
				duration := time.Since(startTime)
				logger.Log().Infof("Completed command: %s (took %v)", cmd.Name(), duration)
				delete(commandTiming, cmd.Name())
			}
		},
	)

	// Add a persistent hook to the root for global initialization/cleanup
	cli.AddPersistentPreRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
		logger.Log().Debugf("Global PreRun: Initializing resources")
	})

	cli.AddPersistentPostRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
		logger.Log().Debugf("Global PostRun: Cleaning up resources")
	})

	// Execute
	if err := cli.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}
