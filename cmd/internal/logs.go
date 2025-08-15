package internal

import (
	"fmt"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/views"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/filesystem"
	"github.com/flowexec/flow/internal/io/logs"
	"github.com/flowexec/flow/internal/logger"
)

func RegisterLogsCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "logs",
		Aliases: []string{"log"},
		Short:   "View execution history and logs.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run: func(cmd *cobra.Command, args []string) {
			logFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.LastLogEntryFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func logFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	lastEntry := flags.ValueFor[bool](cmd, *flags.LastLogEntryFlag, false)
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if err := filesystem.EnsureLogsDir(); err != nil {
		logger.Log().FatalErr(err)
	}
	if TUIEnabled(ctx, cmd) {
		view := views.NewLogArchiveView(ctx.TUIContainer.RenderState(), filesystem.LogsDir(), lastEntry)
		SetView(ctx, cmd, view)
		return
	}
	entries, err := tuikitIO.ListArchiveEntries(filesystem.LogsDir())
	if err != nil {
		logger.Log().FatalErr(err)
	}

	if lastEntry {
		if len(entries) == 0 {
			logger.Log().Fatalf("No log entries found")
		}
		data, err := entries[0].Read()
		if err != nil {
			logger.Log().FatalErr(err)
		}
		_, _ = fmt.Fprint(ctx.StdOut(), data)
	} else {
		logs.PrintEntries(outputFormat, entries)
	}
}
