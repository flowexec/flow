package internal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	pkgstore "github.com/flowexec/flow/pkg/store"
)

func RegisterHistoryCmd(ctx *context.Context, rootCmd *cobra.Command) {
	historyCmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "View execution history.",
		Long:    "View execution history recorded in the data store. Optionally filter by executable reference.",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			historyFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, historyCmd, *flags.OutputFormatFlag)

	clearCmd := &cobra.Command{
		Use:   "clear [ref]",
		Short: "Clear execution history.",
		Long:  "Remove execution history records. If a ref is provided, only that executable's history is cleared.",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			historyClearFunc(ctx, args)
		},
	}

	historyCmd.AddCommand(clearCmd)
	rootCmd.AddCommand(historyCmd)
}

func historyFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)

	var records []pkgstore.ExecutionRecord
	var err error
	if len(args) == 1 {
		records, err = ctx.DataStore.GetExecutionHistory(args[0], 50)
	} else {
		records, err = getAllExecutionHistory(ctx.DataStore)
	}
	if err != nil {
		logger.Log().FatalErr(err)
	}

	if len(records) == 0 {
		logger.Log().Println("No execution history found.")
		return
	}

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			logger.Log().FatalErr(err)
		}
		logger.Log().Println(string(data))
	default:
		for _, r := range records {
			status := "ok"
			if r.ExitCode != 0 {
				status = fmt.Sprintf("exit(%d)", r.ExitCode)
			}
			logger.Log().Println(fmt.Sprintf(
				"%s  %-40s  %6s  %s",
				r.StartedAt.Format(time.RFC3339),
				r.Ref,
				r.Duration.Round(time.Millisecond),
				status,
			))
		}
	}
}

func historyClearFunc(ctx *context.Context, args []string) {
	if ctx.DataStore == nil {
		logger.Log().Fatalf("data store is not available")
	}

	if len(args) == 1 {
		if err := ctx.DataStore.DeleteExecutionHistory(args[0]); err != nil {
			logger.Log().FatalErr(err)
		}
		logger.Log().Println(fmt.Sprintf("Cleared history for %s.", args[0]))
	} else {
		logger.Log().Fatalf("Specify a ref to clear, e.g.: flow history clear ws/ns:name")
	}
}

// getAllExecutionHistory retrieves recent history across all refs, up to 10 records per ref.
func getAllExecutionHistory(ds pkgstore.DataStore) ([]pkgstore.ExecutionRecord, error) {
	refs, err := ds.ListExecutionRefs()
	if err != nil {
		return nil, err
	}
	var all []pkgstore.ExecutionRecord
	for _, ref := range refs {
		records, err := ds.GetExecutionHistory(ref, 10)
		if err != nil {
			continue
		}
		all = append(all, records...)
	}
	return all, nil
}
