package internal

import (
	"fmt"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/io/logs"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

func RegisterLogsCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "logs [ref]",
		Aliases: []string{"log", "history", "hist"},
		Short:   "View execution history and logs.",
		Long: "View execution history recorded in the data store, with associated log output. " +
			"Optionally filter by executable reference.",
		Args:    cobra.MaximumNArgs(1),
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run: func(cmd *cobra.Command, args []string) {
			logFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.LastLogEntryFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)

	clearCmd := &cobra.Command{
		Use:   "clear [ref]",
		Short: "Clear execution history and logs.",
		Long: "Remove execution history records and associated log files. " +
			"If a ref is provided, only that executable's data is cleared.",
		Args: cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logsClearFunc(ctx, args)
		},
	}

	subCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(subCmd)
}

func logFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	lastEntry := flags.ValueFor[bool](cmd, *flags.LastLogEntryFlag, false)
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if err := filesystem.EnsureLogsDir(); err != nil {
		logger.Log().FatalErr(err)
	}

	var records []logs.UnifiedRecord
	var err error
	if len(args) == 1 {
		records, err = logs.LoadRecordsForRef(ctx.DataStore, filesystem.LogsDir(), args[0], 50)
	} else {
		records, err = logs.LoadRecords(ctx.DataStore, filesystem.LogsDir())
	}
	if err != nil {
		logger.Log().FatalErr(err)
	}

	if TUIEnabled(ctx, cmd) {
		view := logs.NewUnifiedLogView(ctx.TUIContainer, records, lastEntry, ctx.DataStore)
		SetView(ctx, cmd, view)
		return
	}

	if lastEntry {
		if len(records) == 0 {
			logger.Log().Fatalf("No execution history found")
		}
		logs.PrintLastRecord(records[0], ctx.StdOut())
	} else {
		logs.PrintRecords(outputFormat, records)
	}
}

func logsClearFunc(ctx *context.Context, args []string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	if len(args) == 1 {
		clearRefHistory(ctx, args[0])
		return
	}
	clearAllHistory(ctx)
}

func clearRefHistory(ctx *context.Context, ref string) {
	deleteAssociatedLogs(ctx, ref)
	if err := ctx.DataStore.DeleteExecutionHistory(ref); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().Println(fmt.Sprintf("Cleared history and logs for %s.", ref))
}

func clearAllHistory(ctx *context.Context) {
	refs, err := ctx.DataStore.ListExecutionRefs()
	if err != nil {
		logger.Log().FatalErr(err)
	}
	for _, ref := range refs {
		deleteAssociatedLogs(ctx, ref)
		_ = ctx.DataStore.DeleteExecutionHistory(ref)
	}
	// Also clean up any orphaned log archive files.
	entries, _ := tuikitIO.ListArchiveEntries(filesystem.LogsDir())
	for _, e := range entries {
		_ = tuikitIO.DeleteArchiveEntry(e.Path)
	}
	logger.Log().Println("Cleared all execution history and logs.")
}

func deleteAssociatedLogs(ctx *context.Context, ref string) {
	records, err := ctx.DataStore.GetExecutionHistory(ref, 0)
	if err != nil {
		return
	}
	for _, r := range records {
		if r.LogArchiveID != "" {
			_ = tuikitIO.DeleteArchiveEntry(r.LogArchiveID)
		}
	}
}
