package logs

import (
	"fmt"
	"time"

	"github.com/flowexec/tuikit"
	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/v2/pkg/store"
)

// NewUnifiedLogView creates a TUI view for unified execution records.
// If lastEntry is true, it shows the detail view for the most recent record.
func NewUnifiedLogView(
	container *tuikit.Container,
	records []UnifiedRecord,
	lastEntry bool,
	ds store.DataStore,
) tuikit.View {
	if len(records) == 0 {
		return views.NewErrorView(fmt.Errorf("no execution history found"), container.RenderState().Theme)
	}

	if lastEntry {
		return unifiedDetailView(container, records[0], ds)
	}
	return unifiedListView(container, records, ds)
}

func unifiedListView(container *tuikit.Container, records []UnifiedRecord, ds store.DataStore) tuikit.View {
	columns := []views.TableColumn{
		{Title: fmt.Sprintf("History (%d)", len(records)), Percentage: 30},
		{Title: "Time", Percentage: 22},
		{Title: "Duration", Percentage: 16},
		{Title: "Status", Percentage: 16},
		{Title: "Origin", Percentage: 16},
	}
	rows := make([]views.TableRow, 0, len(records))
	for i, r := range records {
		rows = append(rows, views.TableRow{
			Data: []string{
				r.Ref,
				r.StartedAt.Format(time.RFC3339),
				r.Duration.Round(time.Millisecond).String(),
				StatusText(r),
				OriginText(r),
				fmt.Sprintf("%d", i),
			},
		})
	}

	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) < 6 {
			return fmt.Errorf("no record selected")
		}
		var idx int
		if _, err := fmt.Sscanf(row.Data()[5], "%d", &idx); err != nil || idx >= len(records) {
			return fmt.Errorf("invalid record")
		}
		return container.SetView(unifiedDetailView(container, records[idx], ds))
	})
	table.SetKeyCallbacks([]types.KeyCallback{
		{Key: "x", Label: "delete all", Callback: func() error {
			for _, r := range records {
				if r.LogEntry != nil {
					_ = tuikitIO.DeleteArchiveEntry(r.LogEntry.Path)
				}
			}
			if ds != nil {
				refs := make(map[string]bool)
				for _, r := range records {
					refs[r.Ref] = true
				}
				for ref := range refs {
					_ = ds.DeleteExecutionHistory(ref)
				}
			}
			container.SetNotice("all execution history deleted", themes.OutputLevelSuccess)
			return nil
		}},
	})
	return table
}

func unifiedDetailView(container *tuikit.Container, record UnifiedRecord, ds store.DataStore) tuikit.View {
	var body string
	switch record.LogEntry {
	case nil:
		body = "no log content available"
	default:
		content, err := record.LogEntry.Read()
		switch {
		case err != nil:
			body = fmt.Sprintf("error reading log: %v", err)
		case content == "":
			body = "no data found in log entry"
		default:
			body = content
		}
	}

	metadata := []views.DetailField{
		{Key: "Executable", Value: record.Ref},
	}
	if record.Label != "" {
		metadata = append(metadata, views.DetailField{Key: "Label", Value: record.Label})
	}
	if record.Command != "" {
		metadata = append(metadata, views.DetailField{Key: "Command", Value: record.Command})
	}
	metadata = append(metadata,
		views.DetailField{Key: "Time", Value: record.StartedAt.Format(time.RFC3339)},
		views.DetailField{Key: "Duration", Value: record.Duration.Round(time.Millisecond).String()},
		views.DetailField{Key: "Status", Value: StatusText(record)},
	)
	if record.Source != "" {
		metadata = append(metadata, views.DetailField{Key: "Source", Value: record.Source})
	}
	if record.ClientName != "" {
		metadata = append(metadata, views.DetailField{Key: "Client", Value: record.ClientName})
	}
	if record.SessionID != "" {
		metadata = append(metadata, views.DetailField{Key: "Session", Value: record.SessionID})
	}
	if record.Error != "" {
		metadata = append(metadata, views.DetailField{Key: "Error", Value: record.Error})
	}

	detail := views.NewDetailView(container.RenderState(), body, metadata...)
	detail.SetKeyCallbacks([]types.KeyCallback{
		{Key: "d", Label: "delete", Callback: func() error {
			if record.LogEntry != nil {
				_ = tuikitIO.DeleteArchiveEntry(record.LogEntry.Path)
			}
			if ds != nil {
				_ = ds.DeleteExecutionHistory(record.Ref)
			}
			container.SetNotice("execution record deleted", themes.OutputLevelSuccess)
			return nil
		}},
	})
	return detail
}
