package internal

import (
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/internal/cache"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/logger"
)

func RegisterSyncCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:   "sync",
		Short: "Refresh workspace cache and discover new executables.",
		Args:  cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			printContext(ctx, cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			syncFunc(ctx, cmd, args)
		},
	}
	rootCmd.AddCommand(subCmd)
}

func syncFunc(_ *context.Context, _ *cobra.Command, _ []string) {
	if err := cache.UpdateAll(); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess("Synced flow cache")
}
