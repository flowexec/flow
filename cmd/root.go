// Package cmd handle the cli commands
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/v2/cmd/internal"
	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/internal/updater"
	"github.com/flowexec/flow/v2/internal/version"
	"github.com/flowexec/flow/v2/pkg/cache"
	"github.com/flowexec/flow/v2/pkg/context"
	flowerrors "github.com/flowexec/flow/v2/pkg/errors"
	"github.com/flowexec/flow/v2/pkg/logger"
)

func NewRootCmd(ctx *context.Context) *cobra.Command {
	rootCmd := &cobra.Command{
		// SilenceErrors prevents cobra from auto-printing "Error: <msg>" on RunE
		// errors; our Execute wrapper re-emits the error via HandleFatal so it can
		// honor --output=json. Usage printing on bad flags is left intact.
		SilenceErrors: true,
		Use:           "flow",
		Short:         "flow is a command line interface designed to make managing and running development workflows easier.",
		Long: "flow is a command line interface designed to make managing and running development workflows easier." +
			"It's driven by executables organized across workspaces and namespaces defined in a workspace.\n\n" +
			"See https://flowexec.io for more information.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level := flags.ValueFor[string](cmd.Root(), *flags.LogLevel, true)
			// TODO: make the tuikit less ambiguous about the log level
			switch level {
			case "debug":
				logger.Log().SetLevel(1)
			case "info":
				logger.Log().SetLevel(0)
			case "fatal":
				logger.Log().SetLevel(-1)
			}
			sync := flags.ValueFor[bool](cmd.Root(), *flags.SyncCacheFlag, true)
			if sync {
				if err := cache.UpdateAll(ctx.DataStore); err != nil {
					errhandler.HandleFatal(ctx, cmd, err)
				}
			}
			updater.CheckInBackground(ctx.DataStore, ctx.Config.UpdateCheck)
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if !isCliUpdateCmd(cmd) {
				printUpdateNotice(ctx, cmd)
			}
			ctx.Finalize()
		},
		Version: version.String(),
	}
	internal.RegisterPersistentFlag(ctx, rootCmd, *flags.LogLevel)
	internal.RegisterPersistentFlag(ctx, rootCmd, *flags.SyncCacheFlag)

	rootCmd.SetOut(ctx.StdOut())
	rootCmd.SetErr(ctx.StdErr())
	rootCmd.SetIn(ctx.StdIn())

	return rootCmd
}

func Execute(ctx *context.Context, rootCmd *cobra.Command) error {
	if ctx == nil {
		panic("current context is not initialized")
	} else if rootCmd == nil {
		panic("root command is not initialized")
	}

	if err := rootCmd.Execute(); err != nil {
		// Cobra-level errors (unknown command, bad flag, arg validation) surface
		// here. Route them through the structured handler so --output=json
		// produces an envelope instead of plain text.
		errhandler.HandleFatal(ctx, rootCmd, flowerrors.NewUsageError("%s", err.Error()))
		return fmt.Errorf("failed to execute command: %w", err)
	}
	return nil
}

func isCliUpdateCmd(cmd *cobra.Command) bool {
	return cmd.Use == "update" && cmd.Parent() != nil && cmd.Parent().Use == "cli"
}

func printUpdateNotice(ctx *context.Context, cmd *cobra.Command) {
	// Skip notice when structured output is requested to avoid polluting JSON/YAML.
	for _, c := range []*cobra.Command{cmd, cmd.Root()} {
		if flags.HasFlag(c, *flags.OutputFormatFlag) {
			if v := flags.ValueFor[string](c, *flags.OutputFormatFlag, false); v == "json" || v == "yaml" || v == "yml" {
				return
			}
		}
	}
	if notice := updater.CachedUpdateNotice(ctx.DataStore); notice != "" {
		logger.Log().Infof("\n%s", notice)
	}
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
	internal.RegisterSchemaCmd(ctx, rootCmd)
	internal.RegisterMCPCmd(ctx, rootCmd)
	internal.RegisterCliCmd(ctx, rootCmd)
}
