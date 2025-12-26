package cli

import (
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/pkg/context"
)

// BuildRootCommand creates a new root command with optional configuration.
// The root command is the main entry point for the CLI.
//
// By default, this creates a root command with Flow's standard configuration.
// Use RootOption functions to customize the command.
func BuildRootCommand(ctx *context.Context, opts ...RootOption) *cobra.Command {
	rootCmd := cmd.NewRootCmd(ctx)

	if len(opts) == 0 {
		return rootCmd
	}

	config := &RootConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.Use != "" {
		rootCmd.Use = config.Use
	}
	if config.Short != "" {
		rootCmd.Short = config.Short
	}
	if config.Long != "" {
		rootCmd.Long = config.Long
	}
	if config.Version != "" {
		rootCmd.Version = config.Version
	}

	return rootCmd
}

// RegisterAllCommands registers all Flow commands to the root command.
func RegisterAllCommands(ctx *context.Context, rootCmd *cobra.Command) {
	cmd.RegisterSubCommands(ctx, rootCmd)
}

// Execute runs the root command. This is a convenience wrapper around cobra's Execute.
func Execute(ctx *context.Context, rootCmd *cobra.Command) error {
	return cmd.Execute(ctx, rootCmd)
}
