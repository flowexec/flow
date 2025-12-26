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
	"os"

	"github.com/spf13/cobra"

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
		StdOut:  os.Stdout,
		LogMode: cfg.DefaultLogMode,
		Theme:   logger.Theme(cfg.Theme.String()),
	}
	logger.Init(loggerOpts)
	defer logger.Log().Flush()

	// Create context
	bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
	ctx := context.NewContext(bkgCtx, cancelFunc, os.Stdin, os.Stdout)
	defer ctx.Finalize()

	// Build root command with custom version
	rootCmd := cli.BuildRootCommand(ctx,
		cli.WithVersion("1.0.0-example"),
		cli.WithShort("Flow CLI - Basic Extension Example"),
	)

	// Register all Flow commands
	cli.RegisterAllCommands(ctx, rootCmd)

	// Add a persistent hook to the root for global initialization/cleanup
	cli.AddPersistentPreRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
		logger.Log().Debugf("Global PreRun: Initializing resources")
	})

	cli.AddPersistentPostRunHook(rootCmd, func(cmd *cobra.Command, args []string) {
		logger.Log().Debugf("Global PostRun: Cleaning up resources")
	})

	// Add a custom command
	premiumCmd := &cobra.Command{
		Use:   "premium",
		Short: "Premium features",
		Long:  "Access premium features of the Flow CLI",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Log().PlainTextInfo("Welcome to Premium Flow CLI!")
			logger.Log().PlainTextInfo("This is a custom command added via the extension API.")
		},
	}

	rootCmd.AddCommand(premiumCmd)

	// Walk all commands and add annotations
	cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations["edition"] = "customized"
	})

	// Execute
	if err := cli.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}
