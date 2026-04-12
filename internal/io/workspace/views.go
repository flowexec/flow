package workspace

import (
	"fmt"
	"path/filepath"

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
	container := ctx.TUIContainer
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

	return views.NewEntityView(container.RenderState(), ws, types.EntityFormatDocument, workspaceKeyCallbacks...)
}

func NewWorkspaceListView(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
) tuikit.View {
	container := ctx.TUIContainer
	if len(workspaces) == 0 {
		container.HandleError(fmt.Errorf("no workspaces found"))
	}

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Workspaces (%d)", len(workspaces)), Percentage: 40},
		{Title: "Location", Percentage: 60},
	}
	rows := make([]views.TableRow, 0, len(workspaces))
	for _, ws := range workspaces {
		rows = append(rows, views.TableRow{
			Data: []string{ws.AssignedName(), ws.Location()},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) == 0 {
			return fmt.Errorf("no workspace selected")
		}
		name := row.Data()[0]
		ws := workspaces.FindByName(name)
		if ws == nil {
			return fmt.Errorf("workspace not found")
		}
		return ctx.SetView(NewWorkspaceView(ctx, ws))
	})
	return table
}
