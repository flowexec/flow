package worktree

import (
	"fmt"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/v2/internal/io/common"
	"github.com/flowexec/flow/v2/internal/services/open"
	"github.com/flowexec/flow/v2/pkg/context"
)

func NewWorktreeListView(
	ctx *context.Context,
	worktrees []DisplayInfo,
) tuikit.View {
	container := ctx.TUIContainer()
	if len(worktrees) == 0 {
		container.HandleError(fmt.Errorf("no worktrees found"))
	}

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Worktrees (%d)", len(worktrees)), Percentage: 25},
		{Title: "Status", Percentage: 15},
		{Title: "Size", Percentage: 10},
		{Title: "Tracking", Percentage: 20},
		{Title: "Path", Percentage: 30},
	}
	rows := buildTableRows(worktrees)

	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.CopyFunc = func(row []string) string {
		if len(row) >= 6 {
			return row[5] // full path in hidden cell
		}
		return ""
	}
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) < 6 {
			return fmt.Errorf("no worktree selected")
		}
		if err := open.Open(row.Data()[5]); err != nil {
			container.HandleError(fmt.Errorf("unable to open worktree: %w", err))
		}
		return nil
	})
	return table
}

func buildTableRows(worktrees []DisplayInfo) []views.TableRow {
	rows := make([]views.TableRow, 0, len(worktrees))
	for _, wt := range worktrees {
		branch := wt.Branch
		if wt.Detached {
			branch = wt.HEAD[:min(8, len(wt.HEAD))] + " (detached)"
		}
		if wt.IsMain {
			branch = "* " + branch
		}
		if wt.IsCurrent {
			branch = "→ " + branch
		}

		status := worktreeStatusLabel(wt)
		tracking := worktreeTrackingLabel(wt)

		rows = append(rows, views.TableRow{
			Data: []string{branch, status, wt.Size, tracking, common.ShortenPath(wt.Path), wt.Path},
		})
	}
	return rows
}

func worktreeStatusLabel(wt DisplayInfo) string {
	switch {
	case wt.Prunable:
		return "prunable"
	case wt.Locked:
		return "locked"
	case wt.Dirty:
		return fmt.Sprintf("dirty (%d)", wt.Changed)
	default:
		return "clean"
	}
}

func worktreeTrackingLabel(wt DisplayInfo) string {
	switch {
	case wt.Merged:
		return "merged"
	case wt.Ahead > 0 && wt.Behind > 0:
		return fmt.Sprintf("↑%d ↓%d", wt.Ahead, wt.Behind)
	case wt.Ahead > 0:
		return fmt.Sprintf("↑%d", wt.Ahead)
	case wt.Behind > 0:
		return fmt.Sprintf("↓%d", wt.Behind)
	default:
		return ""
	}
}
