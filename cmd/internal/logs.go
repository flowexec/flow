package internal

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"strings"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/io/logs"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

func RegisterLogsCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "logs [ref]",
		Aliases: []string{"log", "history", "hist"},
		Short:   "View execution history and logs.",
		Long: "View execution history recorded in the data store, with associated log output. " +
			"Optionally filter by executable reference.",
		Args:    cobra.ArbitraryArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run: func(cmd *cobra.Command, args []string) {
			logFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.LastLogEntryFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, subCmd, *flags.LogFilterWorkspaceFlag)
	RegisterFlag(ctx, subCmd, *flags.LogFilterStatusFlag)
	RegisterFlag(ctx, subCmd, *flags.LogFilterSinceFlag)
	RegisterFlag(ctx, subCmd, *flags.LogFilterLimitFlag)

	clearCmd := &cobra.Command{
		Use:   "clear [ref]",
		Short: "Clear execution history and logs.",
		Long: "Remove execution history records and associated log files. " +
			"If a ref is provided, only that executable's data is cleared.",
		Args: cobra.ArbitraryArgs,
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
	if len(args) > 0 {
		ref := expandRefArgs(ctx, args)
		records, err = logs.LoadRecordsForRef(ctx.DataStore, filesystem.LogsDir(), ref, 50)
	} else {
		records, err = logs.LoadRecords(ctx.DataStore, filesystem.LogsDir())
	}
	if err != nil {
		logger.Log().FatalErr(err)
	}

	filter := buildRecordFilter(cmd)
	records = logs.FilterRecords(records, filter)

	if TUIEnabled(ctx, cmd) {
		view := logs.NewUnifiedLogView(ctx.TUIContainer, records, lastEntry, ctx.DataStore)
		SetView(ctx, cmd, view)
		return
	}

	if lastEntry {
		if len(records) == 0 {
			logger.Log().Fatalf("No execution history found")
		}
		logs.PrintLastRecord(outputFormat, records[0], ctx.StdOut())
	} else {
		logs.PrintRecords(outputFormat, records)
	}
}

// durationWithDays extends time.ParseDuration to support a "d" (day) suffix.
var durationWithDaysRe = regexp.MustCompile(`^(\d+)d(.*)$`)

func parseDurationWithDays(s string) (time.Duration, error) {
	if m := durationWithDaysRe.FindStringSubmatch(s); m != nil {
		days, _ := strconv.Atoi(m[1])
		d := time.Duration(days) * 24 * time.Hour
		if m[2] != "" {
			rest, err := time.ParseDuration(m[2])
			if err != nil {
				return 0, fmt.Errorf("invalid duration %q: %w", s, err)
			}
			d += rest
		}
		return d, nil
	}
	return time.ParseDuration(s)
}

func buildRecordFilter(cmd *cobra.Command) logs.RecordFilter {
	var f logs.RecordFilter

	f.Workspace = flags.ValueFor[string](cmd, *flags.LogFilterWorkspaceFlag, false)
	f.Status = strings.ToLower(flags.ValueFor[string](cmd, *flags.LogFilterStatusFlag, false))
	f.Limit = flags.ValueFor[int](cmd, *flags.LogFilterLimitFlag, false)

	sinceStr := flags.ValueFor[string](cmd, *flags.LogFilterSinceFlag, false)
	if sinceStr != "" {
		d, err := parseDurationWithDays(sinceStr)
		if err != nil {
			logger.Log().Fatalf("Invalid --since value %q: %v", sinceStr, err)
		}
		f.Since = time.Now().Add(-d)
	}

	return f
}

// expandRefArgs builds a fully-qualified ref string from args, expanding the
// current workspace/namespace when not provided — matching how exec records refs.
func expandRefArgs(ctx *context.Context, args []string) string {
	verb := executable.Verb(args[0])
	id := strings.Join(args[1:], " ")
	ref := context.ExpandRef(ctx, executable.NewRef(id, verb))
	return ref.String()
}

func logsClearFunc(ctx *context.Context, args []string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	if len(args) > 0 {
		ref := expandRefArgs(ctx, args)
		clearRefHistory(ctx, ref)
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
