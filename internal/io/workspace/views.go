package workspace

import (
	"fmt"
	"path/filepath"
	"sort"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/services/open"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/types/workspace"
)

func NewWorkspaceView(
	ctx *context.Context,
	ws *workspace.Workspace,
) tuikit.View {
	container := ctx.TUIContainer()
	var workspaceKeyCallbacks = []types.KeyCallback{
		{
			Key: "o", Label: "open",
			Callback: func() error {
				if err := open.Open(ws.Location()); err != nil {
					container.HandleError(fmt.Errorf("unable to open workspace: %w", err))
				}
				return nil
			},
		},
		{
			Key: "e", Label: "edit",
			Callback: func() error {
				fullPath := filepath.Join(ws.Location(), filesystem.WorkspaceConfigFileName)
				if err := common.OpenInEditor(fullPath, ctx.StdIn(), ctx.StdOut()); err != nil {
					container.HandleError(fmt.Errorf("unable to edit workspace: %w", err))
				}
				return nil
			},
		},
		{
			Key: "s", Label: "set",
			Callback: func() error {
				curCfg, err := filesystem.LoadConfig()
				if err != nil {
					container.HandleError(err)
					return nil
				}
				curCfg.CurrentWorkspace = ws.AssignedName()
				if err := filesystem.WriteConfig(curCfg); err != nil {
					container.HandleError(err)
				}
				container.SetState(common.HeaderContextKey, fmt.Sprintf("%s/*", ws.AssignedName()))
				container.SetNotice("workspace updated", themes.OutputLevelInfo)
				return nil
			},
		},
	}

	opts := workspaceDetailOpts(ws)
	return views.NewDetailContentView(container.RenderState(), opts, workspaceKeyCallbacks...)
}

func NewWorkspaceListView(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
) tuikit.View {
	container := ctx.TUIContainer()
	if len(workspaces) == 0 {
		container.HandleError(fmt.Errorf("no workspaces found"))
	}

	sort.Slice(workspaces, func(i, j int) bool {
		return workspaces[i].AssignedName() < workspaces[j].AssignedName()
	})

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Workspaces (%d)", len(workspaces)), Percentage: 35},
		{Title: "Tags", Percentage: 30},
		{Title: "Location", Percentage: 35},
	}
	rows := make([]views.TableRow, 0, len(workspaces))
	for _, ws := range workspaces {
		name := ws.AssignedName()
		if ws.DisplayName != "" {
			name = ws.DisplayName
		}
		tags := common.ColorizeTags(ws.Tags)
		rows = append(rows, views.TableRow{
			Data: []string{name, tags, common.ShortenPath(ws.Location()), ws.AssignedName()},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) < 4 {
			return fmt.Errorf("no workspace selected")
		}
		// Hidden cell [3] holds the assigned name for lookup
		ws := workspaces.FindByName(row.Data()[3])
		if ws == nil {
			return fmt.Errorf("workspace not found")
		}
		return ctx.SetView(NewWorkspaceView(ctx, ws))
	})
	return table
}
