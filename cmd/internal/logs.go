package internal

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/internal/io/logs"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
	"github.com/flowexec/flow/v2/types/executable"
)

func RegisterLogsCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "logs [ref]",
		Aliases: []string{"log", "history", "hist"},
		Short:   "View execution history and logs.",
		Example: logsExamples,
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

	killCmd := &cobra.Command{
		Use:   "kill RUN_ID",
		Short: "Terminate a running background process by run ID.",
		Long:  "Send a termination signal to a running background process identified by its run ID.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logsKillFunc(ctx, args[0])
		},
	}

	attachCmd := &cobra.Command{
		Use:   "attach RUN_ID",
		Short: "Stream log output from a running background process by run ID.",
		Long:  "Stream the log output of a background process identified by its run ID.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			logsAttachFunc(ctx, args[0])
		},
	}

	subCmd.AddCommand(clearCmd)
	subCmd.AddCommand(killCmd)
	subCmd.AddCommand(attachCmd)
	RegisterFlag(ctx, subCmd, *flags.RunningFlag)
	rootCmd.AddCommand(subCmd)
}

func logFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	running := flags.ValueFor[bool](cmd, *flags.RunningFlag, false)
	if running {
		logsRunningFunc(ctx, cmd)
		return
	}

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
		view := logs.NewUnifiedLogView(ctx.TUIContainer(), records, lastEntry, ctx.DataStore)
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

// logsRunningFunc lists active background processes using the same output/TUI
// patterns as regular execution history.
func logsRunningFunc(ctx *context.Context, cmd *cobra.Command) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	runs, err := ctx.DataStore.ListBackgroundRuns()
	if err != nil {
		logger.Log().FatalErr(err)
	}

	// Prune stale entries and collect active runs.
	var active []store.BackgroundRun
	for _, run := range runs {
		if run.Status != store.BackgroundRunning {
			continue
		}
		if !isProcessAlive(run.PID) {
			now := time.Now()
			run.Status = store.BackgroundFailed
			run.Error = "process exited unexpectedly"
			run.CompletedAt = &now
			_ = ctx.DataStore.SaveBackgroundRun(run)
			continue
		}
		active = append(active, run)
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	if TUIEnabled(ctx, cmd) {
		view := logs.NewBackgroundRunsView(ctx.TUIContainer(), active, ctx.DataStore)
		SetView(ctx, cmd, view)
		return
	}

	logs.PrintBackgroundRuns(outputFormat, active)
}

// logsKillFunc terminates a background process by run ID.
func logsKillFunc(ctx *context.Context, runID string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	run, err := ctx.DataStore.GetBackgroundRun(runID)
	if err != nil {
		logger.Log().FatalErr(fmt.Errorf("background run %s not found: %w", runID, err))
	}

	if run.Status != store.BackgroundRunning {
		logger.Log().Fatalf("background run %s is not running (status: %s)", runID, run.Status)
	}

	proc, err := os.FindProcess(run.PID)
	if err != nil {
		logger.Log().FatalErr(fmt.Errorf("unable to find process %d: %w", run.PID, err))
	}

	if err := terminateProcess(proc); err != nil {
		logger.Log().FatalErr(fmt.Errorf("failed to terminate process %d: %w", run.PID, err))
	}

	now := time.Now()
	run.Status = store.BackgroundFailed
	run.Error = "killed by user"
	run.CompletedAt = &now
	if err := ctx.DataStore.SaveBackgroundRun(run); err != nil {
		logger.Log().Errorf("failed to update background run record: %v", err)
	}

	logger.Log().Println(fmt.Sprintf("Terminated background run %s (PID %d).", runID, run.PID))
}

// logsAttachFunc streams log output from a background process, tail-following
// the log archive file until the process exits or the user interrupts.
func logsAttachFunc(ctx *context.Context, runID string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	run, err := ctx.DataStore.GetBackgroundRun(runID)
	if err != nil {
		logger.Log().FatalErr(fmt.Errorf("background run %s not found: %w", runID, err))
	}

	archivePath := run.LogArchiveID
	if archivePath == "" {
		logger.Log().Fatalf("no log output available for background run %s (archive not yet linked)", runID)
	}

	f, err := os.Open(archivePath)
	if err != nil {
		logger.Log().FatalErr(fmt.Errorf("unable to open log archive: %w", err))
	}
	defer f.Close()

	sigCh := make(chan os.Signal, 1)
	notifyTermSignals(sigCh)
	defer signal.Stop(sigCh)

	buf := make([]byte, 4096)
	var pos int64
	for {
		select {
		case <-sigCh:
			_, _ = fmt.Fprintln(ctx.StdOut(), "\nDetached.")
			return
		default:
		}

		n, readErr := f.ReadAt(buf, pos)
		if n > 0 {
			_, _ = ctx.StdOut().Write(buf[:n])
			pos += int64(n)
			continue
		}

		// No new data — check if the process is still alive.
		if !isProcessAlive(run.PID) {
			// Final drain.
			for {
				n, _ = f.ReadAt(buf, pos)
				if n == 0 {
					break
				}
				_, _ = ctx.StdOut().Write(buf[:n])
				pos += int64(n)
			}
			_, _ = fmt.Fprintln(ctx.StdOut(), "\nBackground process exited.")
			return
		}

		if readErr != nil && readErr != io.EOF {
			logger.Log().FatalErr(fmt.Errorf("error reading log: %w", readErr))
		}

		time.Sleep(200 * time.Millisecond)
	}
}

const logsExamples = `
  flow logs                          # all history
  flow logs --last                   # most recent entry with full output
  flow logs --status failed   # only failed runs
  flow logs run build                # history for 'run build' executable
  flow logs --running                # list active background processes
`
