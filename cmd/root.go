// Package cmd handle the cli commands
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/version"
	"github.com/flowexec/flow/internal/cache"
	"github.com/flowexec/flow/internal/context"
)

func NewRootCmd(ctx *context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "flow",
		Short: "flow is a command line interface designed to make managing and running development workflows easier.",
		Long: "flow is a command line interface designed to make managing and running development workflows easier." +
			"It's driven by executables organized across workspaces and namespaces defined in a workspace.\n\n" +
			"See https://flowexec.io for more information.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := flags.ValueFor[string](ctx, cmd.Root(), *flags.LogLevel, true)
			// TODO: make the tuikit less ambiguous about the log level
			switch level {
			case "debug":
				ctx.Logger.SetLevel(1)
			case "info":
				ctx.Logger.SetLevel(0)
			case "fatal":
				ctx.Logger.SetLevel(-1)
			}
			sync := flags.ValueFor[bool](ctx, cmd.Root(), *flags.SyncCacheFlag, true)
			if sync {
				if err := cache.UpdateAll(ctx.Logger); err != nil {
					ctx.Logger.FatalErr(err)
				}
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) { ctx.Finalize() },
		Version:           version.String(),
	}
	internal.RegisterPersistentFlag(ctx, rootCmd, *flags.LogLevel)
	internal.RegisterPersistentFlag(ctx, rootCmd, *flags.SyncCacheFlag)
	return rootCmd
}

func Execute(ctx *context.Context, rootCmd *cobra.Command) error {
	if ctx == nil {
		panic("current context is not initialized")
	} else if rootCmd == nil {
		panic("root command is not initialized")
	}

	rootCmd.SetOut(ctx.StdOut())
	rootCmd.SetErr(ctx.StdOut())
	rootCmd.SetIn(ctx.StdIn())
	RegisterSubCommands(ctx, rootCmd)

	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

func RegisterSubCommands(ctx *context.Context, rootCmd *cobra.Command) {
	if ctx == nil {
		panic("current context is not initialized")
	} else if rootCmd == nil {
		panic("root command is not initialized")
	}

	internal.RegisterExecCmd(ctx, rootCmd)
	internal.RegisterBrowseCmd(ctx, rootCmd)
	internal.RegisterConfigCmd(ctx, rootCmd)
	internal.RegisterSecretCmd(ctx, rootCmd)
	internal.RegisterVaultCmd(ctx, rootCmd)
	internal.RegisterCacheCmd(ctx, rootCmd)
	internal.RegisterWorkspaceCmd(ctx, rootCmd)
	internal.RegisterTemplateCmd(ctx, rootCmd)
	internal.RegisterLogsCmd(ctx, rootCmd)
	internal.RegisterSyncCmd(ctx, rootCmd)
	internal.RegisterMCPCmd(ctx, rootCmd)
}
