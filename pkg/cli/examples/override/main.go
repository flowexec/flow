// Package main demonstrates command overriding using the Flow CLI extension API.
//
// This example shows how to:
// - Replace existing commands with custom implementations
// - Add new commands to the CLI
// - Customize command behavior
package main

import (
	stdCtx "context"
	"fmt"

	"github.com/spf13/cobra"

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

	// Build root command
	rootCmd := cli.BuildRootCommand(ctx,
		cli.WithVersion("1.0.0-override-example"),
		cli.WithShort("Flow CLI - Command Override Example"),
	)

	// Register all Flow commands
	cli.RegisterAllCommands(ctx, rootCmd)

	// Add a custom "premium" command
	premiumCmd := &cobra.Command{
		Use:   "premium",
		Short: "Premium features",
		Long:  "Access premium features of the Flow CLI",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Log().PlainTextInfo("Welcome to Premium Flow CLI!")
			logger.Log().PlainTextInfo("This is a custom command added via the extension API.")
		},
	}

	// Add subcommands to the premium command
	premiumCmd.AddCommand(&cobra.Command{
		Use:   "status",
		Short: "Check premium status",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Log().PlainTextInfo("Premium Status: Active")
			logger.Log().PlainTextInfo("License: Enterprise")
			logger.Log().PlainTextInfo("Expiry: Never")
		},
	})

	rootCmd.AddCommand(premiumCmd)

	// Override the version command with custom behavior
	// First, let's add a custom version command that shows additional info
	customVersionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information with premium details",
		Run: func(cmd *cobra.Command, args []string) {
			logger.Log().PlainTextInfo("Flow CLI Version: 1.0.0-override-example")
			logger.Log().PlainTextInfo("Edition: Premium")
			logger.Log().PlainTextInfo("Build Date: 2025-12-26")
			logger.Log().PlainTextInfo("License: Enterprise")
		},
	}

	// Since version is built into the root command via --version flag,
	// we'll add it as a subcommand instead
	rootCmd.AddCommand(customVersionCmd)

	// Walk all commands and add annotations
	cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations["edition"] = "premium"
		cmd.Annotations["customized"] = "true"
	})

	// Execute
	if err := cli.Execute(ctx, rootCmd); err != nil {
		logger.Log().FatalErr(err)
	}
}
