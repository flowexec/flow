package logs

import (
	"encoding/json"
	"fmt"
	"time"

	"os"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"gopkg.in/yaml.v3"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/pkg/store"
)

type backgroundRunOutput struct {
	ID        string `json:"id"              yaml:"id"`
	PID       int    `json:"pid"             yaml:"pid"`
	Ref       string `json:"ref"             yaml:"ref"`
	StartedAt string `json:"startedAt"       yaml:"startedAt"`
	Status    string `json:"status"          yaml:"status"`
	Error     string `json:"error,omitempty" yaml:"error,omitempty"`
}

type backgroundRunsResponse struct {
	Runs []backgroundRunOutput `json:"runs" yaml:"runs"`
}

func toBackgroundRunOutput(r store.BackgroundRun) backgroundRunOutput {
	return backgroundRunOutput{
		ID:        r.ID,
		PID:       r.PID,
		Ref:       r.Ref,
		StartedAt: r.StartedAt.Format(time.RFC3339),
		Status:    string(r.Status),
		Error:     r.Error,
	}
}

// PrintBackgroundRuns outputs background runs in the specified format (json, yaml, or plain text).
func PrintBackgroundRuns(format string, runs []store.BackgroundRun) {
	out := make([]backgroundRunOutput, len(runs))
	for i, r := range runs {
		out[i] = toBackgroundRunOutput(r)
	}

	switch common.NormalizeFormat(format) {
	case common.JSONFormat:
		data, err := json.MarshalIndent(backgroundRunsResponse{Runs: out}, "", "  ")
		if err != nil {
			logger.Log().Fatalf("Failed to marshal background runs - %v", err)
		}
		logger.Log().Println(string(data))
	case common.YAMLFormat:
		data, err := yaml.Marshal(backgroundRunsResponse{Runs: out})
		if err != nil {
			logger.Log().Fatalf("Failed to marshal background runs - %v", err)
		}
		logger.Log().Println(string(data))
	default:
		if len(runs) == 0 {
			logger.Log().Println("No active background processes.")
			return
		}
		printBackgroundRunsText(runs)
	}
}

func printBackgroundRunsText(runs []store.BackgroundRun) {
	for _, r := range runs {
		dur := time.Since(r.StartedAt).Round(time.Second)
		logger.Log().Println(fmt.Sprintf(
			"%-8s  PID %-7d  %-40s  running %s",
			r.ID,
			r.PID,
			r.Ref,
			dur,
		))
	}
}

// NewBackgroundRunsView creates a TUI view for active background runs.
func NewBackgroundRunsView(
	container *tuikit.Container,
	runs []store.BackgroundRun,
	ds store.DataStore,
) tuikit.View {
	if len(runs) == 0 {
		return views.NewErrorView(fmt.Errorf("no active background processes"), container.RenderState().Theme)
	}

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Background (%d)", len(runs)), Percentage: 20},
		{Title: "Executable", Percentage: 40},
		{Title: "PID", Percentage: 10},
		{Title: "Running", Percentage: 30},
	}
	rows := make([]views.TableRow, 0, len(runs))
	for i, r := range runs {
		dur := time.Since(r.StartedAt).Round(time.Second)
		rows = append(rows, views.TableRow{
			Data: []string{
				r.ID,
				r.Ref,
				fmt.Sprintf("%d", r.PID),
				dur.String(),
				fmt.Sprintf("%d", i),
			},
		})
	}

	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetKeyCallbacks([]types.KeyCallback{
		{Key: "x", Label: "kill all", Callback: func() error {
			for _, r := range runs {
				killAndUpdate(r, ds)
			}
			container.SetNotice("all background runs terminated", themes.OutputLevelSuccess)
			return nil
		}},
	})
	return table
}

func killAndUpdate(r store.BackgroundRun, ds store.DataStore) {
	proc, err := os.FindProcess(r.PID)
	if err != nil {
		return
	}
	_ = killProcess(proc)
	now := time.Now()
	r.Status = store.BackgroundFailed
	r.Error = "killed by user"
	r.CompletedAt = &now
	if ds != nil {
		_ = ds.SaveBackgroundRun(r)
	}
}
