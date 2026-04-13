package logs

import (
	"fmt"
	"slices"

	"github.com/flowexec/tuikit"
	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"
)

func NewLogView(
	container *tuikit.Container,
	archiveDir string,
	lastEntry bool,
) tuikit.View {
	entries, err := tuikitIO.ListArchiveEntries(archiveDir)
	if err != nil {
		return views.NewErrorView(err, container.RenderState().Theme)
	}
	if len(entries) == 0 {
		return views.NewErrorView(fmt.Errorf("no log entries found"), container.RenderState().Theme)
	}

	// Most recent first
	slices.Reverse(entries)

	if lastEntry {
		return logDetailView(container, entries[0])
	}

	return logListView(container, entries)
}

func logDetailView(container *tuikit.Container, entry tuikitIO.ArchiveEntry) tuikit.View {
	content, err := entry.Read()
	if err != nil {
		return views.NewErrorView(err, container.RenderState().Theme)
	}
	if content == "" {
		content = "no data found in log entry"
	}

	metadata := []views.DetailField{
		{Key: "Executable", Value: entry.ID},
		{Key: "Time", Value: entry.Description()},
	}

	detail := views.NewDetailView(container.RenderState(), content, metadata...)
	detail.SetKeyCallbacks([]types.KeyCallback{
		{Key: "d", Label: "delete", Callback: func() error {
			if err := tuikitIO.DeleteArchiveEntry(entry.Path); err != nil {
				container.SetNotice("unable to delete log entry", themes.OutputLevelError)
			} else {
				container.SetNotice("log entry deleted", themes.OutputLevelSuccess)
			}
			return nil
		}},
	})
	return detail
}

func logListView(container *tuikit.Container, entries []tuikitIO.ArchiveEntry) tuikit.View {
	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Logs (%d)", len(entries)), Percentage: 50},
		{Title: "Time", Percentage: 50},
	}
	rows := make([]views.TableRow, 0, len(entries))
	for i, e := range entries {
		rows = append(rows, views.TableRow{
			Data: []string{e.ID, e.Description(), fmt.Sprintf("%d", i)},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) < 3 {
			return fmt.Errorf("no log entry selected")
		}
		var idx int
		if _, err := fmt.Sscanf(row.Data()[2], "%d", &idx); err != nil || idx >= len(entries) {
			return fmt.Errorf("invalid log entry")
		}
		return container.SetView(logDetailView(container, entries[idx]))
	})
	table.SetKeyCallbacks([]types.KeyCallback{
		{Key: "x", Label: "delete all", Callback: func() error {
			for _, e := range entries {
				_ = tuikitIO.DeleteArchiveEntry(e.Path)
			}
			container.SetNotice("all log entries deleted", themes.OutputLevelSuccess)
			return nil
		}},
	})
	return table
}
