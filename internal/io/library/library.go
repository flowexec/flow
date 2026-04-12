package library

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/io/common"
	execIO "github.com/flowexec/flow/internal/io/executable"
	"github.com/flowexec/flow/internal/services/open"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	flowCommon "github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

// Filter narrows the executables shown by the library view.
type Filter struct {
	Workspace, Namespace string
	Verb                 executable.Verb
	Tags                 flowCommon.Tags
	Substring            string
	Visibility           flowCommon.Visibility
}

// Sentinel labels rendered in the workspace table for top-level entries
// and the implicit "no namespace" child row.
const (
	allWorkspacesLabel = "All Workspaces"
	rootNamespaceLabel = "Root Namespace"
)

// Hidden cells appended to TableRow.Data so the executable page can
// recover the workspace/namespace context that the user selected. The
// table only renders cells that map to a configured column, so cells
// past the column count are silently dropped from display while still
// being returned by Selectable.SelectedData().
const (
	wsRowCellKind   = 3 // "all" | "workspace" | "namespace"
	wsRowCellWsName = 4 // workspace name (for namespace children)
	wsRowCellNsName = 5 // namespace name ("" for root namespace)
	wsRowKindAll    = "all"
	wsRowKindWS     = "workspace"
	wsRowKindNS     = "namespace"
)

// NewLibraryView builds the multi-page library browser. Pages drill down
// from workspaces -> executables -> executable details.
func NewLibraryView(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
	execs executable.ExecutableList,
	filter Filter,
	_ themes.Theme,
	runFunc func(string) error,
) tuikit.View {
	container := ctx.TUIContainer
	return views.NewLibrary(
		container.RenderState(),
		workspacesPage(ctx, workspaces, execs, filter, runFunc),
		executablesPage(ctx, execs, filter, runFunc),
		executableDetailPage(ctx, execs, filter, runFunc),
	)
}

// applyFilter returns the executables visible after applying the user's
// filter (workspace, namespace, verb, tags, substring, visibility).
func applyFilter(execs executable.ExecutableList, filter Filter) executable.ExecutableList {
	visibility := filter.Visibility
	if visibility == "" {
		visibility = flowCommon.VisibilityPrivate
	}
	return execs.
		FilterByWorkspaceWithVisibility(filter.Workspace, visibility).
		FilterByNamespace(filter.Namespace).
		FilterByVerb(filter.Verb).
		FilterByTags(filter.Tags).
		FilterBySubstring(filter.Substring)
}

// workspaceList returns the workspaces matching the filter, sorted by
// assigned name. When no filter is set, all workspaces are returned.
func workspaceList(workspaces workspace.WorkspaceList, filter Filter) workspace.WorkspaceList {
	if filter.Workspace == "" || filter.Workspace == executable.WildcardWorkspace {
		out := append(workspace.WorkspaceList{}, workspaces...)
		sort.Slice(out, func(i, j int) bool {
			return out[i].AssignedName() < out[j].AssignedName()
		})
		return out
	}
	for _, ws := range workspaces {
		if ws.AssignedName() == filter.Workspace {
			return workspace.WorkspaceList{ws}
		}
	}
	return workspace.WorkspaceList{}
}

// namespacesForWorkspace returns the namespaces (and whether a root
// namespace exists) for the given workspace, after applying the
// non-namespace parts of the user's filter. Passing an empty wsName
// returns the union across all workspaces.
func namespacesForWorkspace(
	execs executable.ExecutableList,
	wsName string,
	filter Filter,
) (hasRoot bool, namespaces []string) {
	wsFilter := filter
	wsFilter.Workspace = wsName
	wsFilter.Namespace = executable.WildcardNamespace
	visible := applyFilter(execs, wsFilter)
	nsSet := map[string]struct{}{}
	for _, ex := range visible {
		ns := ex.Namespace()
		if ns == "" || ns == executable.WildcardNamespace {
			hasRoot = true
			continue
		}
		nsSet[ns] = struct{}{}
	}
	for ns := range nsSet {
		namespaces = append(namespaces, ns)
	}
	sort.Strings(namespaces)
	return hasRoot, namespaces
}

// namespaceChildren builds the table child rows for a workspace's
// namespaces. wsName is the parent workspace ("" for "All Workspaces").
func namespaceChildren(execs executable.ExecutableList, wsName string, filter Filter) []views.TableRow {
	hasRoot, namespaces := namespacesForWorkspace(execs, wsName, filter)
	children := make([]views.TableRow, 0, len(namespaces)+1)
	if hasRoot {
		children = append(children, views.TableRow{
			Data: padRow([]string{rootNamespaceLabel, "", ""}, wsRowKindNS, wsName, ""),
		})
	}
	for _, ns := range namespaces {
		children = append(children, views.TableRow{
			Data: padRow([]string{ns, "", ""}, wsRowKindNS, wsName, ns),
		})
	}
	return children
}

func workspacePageRows(
	wsList workspace.WorkspaceList,
	execs executable.ExecutableList,
	filter Filter,
) []views.TableRow {
	allCount := len(applyFilter(execs, withWorkspace(filter, "")))
	rows := []views.TableRow{
		{
			Data: padRow([]string{
				allWorkspacesLabel,
				fmt.Sprintf("%d executables", allCount),
				"",
			}, wsRowKindAll, "", ""),
			Children: namespaceChildren(execs, "", filter),
		},
	}

	for _, ws := range wsList {
		count := len(applyFilter(execs, withWorkspace(filter, ws.AssignedName())))
		desc := ws.DisplayName
		if len(ws.Tags) > 0 {
			desc = fmt.Sprintf("%s [%s]", desc, flowCommon.Tags(ws.Tags).PreviewString())
		}
		if desc == "" {
			desc = fmt.Sprintf("%d executables", count)
		}
		rows = append(rows, views.TableRow{
			Data:     padRow([]string{ws.AssignedName(), desc, ws.Location()}, wsRowKindWS, ws.AssignedName(), ""),
			Children: namespaceChildren(execs, ws.AssignedName(), filter),
		})
	}
	return rows
}

func workspacePageKeys(ctx *context.Context, table *views.Table, wsList workspace.WorkspaceList) []types.KeyCallback {
	container := ctx.TUIContainer
	return []types.KeyCallback{
		{Key: "o", Label: "open", Callback: func() error {
			ws := selectedWorkspace(table, wsList)
			if ws == nil {
				container.SetNotice("no workspace selected", themes.OutputLevelError)
				return nil
			}
			if err := open.Open(ws.Location()); err != nil {
				logger.Log().WrapError(err, "unable to open workspace")
				container.SetNotice("unable to open workspace", themes.OutputLevelError)
			}
			return nil
		}},
		{Key: "e", Label: "edit", Callback: func() error {
			ws := selectedWorkspace(table, wsList)
			if ws == nil {
				container.SetNotice("no workspace selected", themes.OutputLevelError)
				return nil
			}
			path := filepath.Join(ws.Location(), filesystem.WorkspaceConfigFileName)
			if err := common.OpenInEditor(path, ctx.StdIn(), ctx.StdOut()); err != nil {
				logger.Log().WrapError(err, "unable to open workspace in editor")
				container.SetNotice("unable to open workspace in editor", themes.OutputLevelError)
			}
			return nil
		}},
		{Key: "s", Label: "set context", Callback: func() error {
			wsName, nsName := decodeSelection(table)
			if wsName == "" {
				container.SetNotice("no workspace selected", themes.OutputLevelError)
				return nil
			}
			ws := findWorkspace(wsList, wsName)
			if ws == nil {
				container.SetNotice("no workspace selected", themes.OutputLevelError)
				return nil
			}
			curCfg, err := filesystem.LoadConfig()
			if err != nil {
				logger.Log().WrapError(err, "unable to load user config")
				container.SetNotice("unable to load user config", themes.OutputLevelError)
				return nil
			}
			curCfg.CurrentWorkspace = ws.AssignedName()
			curCfg.CurrentNamespace = nsName
			if err := filesystem.WriteConfig(curCfg); err != nil {
				logger.Log().WrapError(err, "unable to write user config")
				container.SetNotice("unable to write user config", themes.OutputLevelError)
				return nil
			}
			ctx.Config.CurrentWorkspace = curCfg.CurrentWorkspace
			ctx.Config.CurrentNamespace = curCfg.CurrentNamespace
			container.SetNotice("context updated", themes.OutputLevelSuccess)
			return nil
		}},
	}
}

func workspacesPage(
	ctx *context.Context,
	workspaces workspace.WorkspaceList,
	execs executable.ExecutableList,
	filter Filter,
	_ func(string) error,
) views.LibraryPage {
	return views.LibraryPage{
		Title: "Workspaces",
		Factory: func(render *types.RenderState, _ []views.PageSelection) (tea.Model, []types.KeyCallback) {
			wsList := workspaceList(workspaces, filter)
			rows := workspacePageRows(wsList, execs, filter)

			columns := []views.TableColumn{
				{Title: fmt.Sprintf("Workspaces (%d)", len(wsList)), Percentage: 30},
				{Title: "Description", Percentage: 50},
				{Title: "Location", Percentage: 20},
			}
			table := views.NewTable(render, columns, rows, views.TableDisplayFull)
			keys := workspacePageKeys(ctx, table, wsList)
			return table, keys
		},
	}
}

func executablePageKeys(
	ctx *context.Context,
	table *views.Table,
	visible executable.ExecutableList,
	runFunc func(string) error,
) []types.KeyCallback {
	container := ctx.TUIContainer
	selectedExec := func() *executable.Executable {
		row := table.GetSelectedRow()
		if row == nil {
			return nil
		}
		data := row.Data()
		if len(data) < 2 {
			return nil
		}
		ex, err := visible.FindByVerbAndID(executable.Verb(data[1]), data[0])
		if err != nil {
			return nil
		}
		return ex
	}
	return []types.KeyCallback{
		{Key: "r", Label: "run", Callback: func() error {
			ex := selectedExec()
			if ex == nil {
				container.SetNotice("no executable selected", themes.OutputLevelError)
				return nil
			}
			container.Shutdown(func() {
				if err := runFunc(ex.Ref().String()); err != nil {
					logger.Log().Fatal("unable to execute command", "error", err)
				}
			})
			os.Exit(0) // prevent the app from hanging after the command is run
			return nil
		}},
		{Key: "e", Label: "edit", Callback: func() error {
			ex := selectedExec()
			if ex == nil {
				container.SetNotice("no executable selected", themes.OutputLevelError)
				return nil
			}
			if err := common.OpenInEditor(ex.FlowFilePath(), ctx.StdIn(), ctx.StdOut()); err != nil {
				logger.Log().WrapError(err, "unable to open executable in editor")
				container.SetNotice("unable to open executable in editor", themes.OutputLevelError)
			}
			return nil
		}},
		{Key: "c", Label: "copy ref", Callback: func() error {
			ex := selectedExec()
			if ex == nil {
				container.SetNotice("no executable selected", themes.OutputLevelError)
				return nil
			}
			if err := clipboard.WriteAll(ex.Ref().String()); err != nil {
				logger.Log().WrapError(err, "unable to copy reference to clipboard")
				container.SetNotice("unable to copy reference to clipboard", themes.OutputLevelError)
			} else {
				container.SetNotice("copied reference to clipboard", themes.OutputLevelInfo)
			}
			return nil
		}},
	}
}

func executablesPage(
	ctx *context.Context,
	execs executable.ExecutableList,
	filter Filter,
	runFunc func(string) error,
) views.LibraryPage {
	return views.LibraryPage{
		Title: "Executables",
		Factory: func(render *types.RenderState, selections []views.PageSelection) (tea.Model, []types.KeyCallback) {
			pageFilter := executableFilterFromSelection(filter, selections)
			visible := applyFilter(execs, pageFilter)
			slices.SortFunc(visible, func(a, b *executable.Executable) int {
				return strings.Compare(a.Ref().String(), b.Ref().String())
			})

			columns := []views.TableColumn{
				{Title: fmt.Sprintf("Executables (%d)", len(visible)), Percentage: 45},
				{Title: "Verb", Percentage: 15},
				{Title: "Tags", Percentage: 40},
			}
			rows := make([]views.TableRow, 0, len(visible))
			for _, ex := range visible {
				tags := ""
				if len(ex.Tags) > 0 {
					tags = flowCommon.Tags(ex.Tags).PreviewString()
				}
				rows = append(rows, views.TableRow{
					Data: []string{ex.Ref().ID(), string(ex.Verb), tags},
				})
			}
			table := views.NewTable(render, columns, rows, views.TableDisplayFull)
			keys := executablePageKeys(ctx, table, visible, runFunc)
			return table, keys
		},
	}
}

func executableDetailPage(
	ctx *context.Context,
	execs executable.ExecutableList,
	filter Filter,
	runFunc func(string) error,
) views.LibraryPage {
	return views.LibraryPage{
		Title: "Details",
		Factory: func(render *types.RenderState, selections []views.PageSelection) (tea.Model, []types.KeyCallback) {
			var ex *executable.Executable
			if len(selections) > 1 && len(selections[1].Data) >= 2 {
				row := selections[1].Data
				pageFilter := executableFilterFromSelection(filter, selections)
				visible := applyFilter(execs, pageFilter)
				if found, err := visible.FindByVerbAndID(executable.Verb(row[1]), row[0]); err == nil {
					ex = found
				}
			}
			if ex == nil {
				return views.NewErrorView(
					fmt.Errorf("no executable selected"), render.Theme,
				), nil
			}
			view := execIO.NewExecutableView(ctx, ex, runFunc)
			model, ok := view.(tea.Model)
			if !ok {
				return views.NewErrorView(
					fmt.Errorf("executable view is not a tea.Model"), render.Theme,
				), nil
			}
			return model, nil
		},
	}
}

// executableFilterFromSelection translates the workspace-page selection
// (workspace row, namespace child row, or "All Workspaces") into a
// concrete Filter for the executable page.
func executableFilterFromSelection(base Filter, selections []views.PageSelection) Filter {
	if len(selections) == 0 {
		return base
	}
	data := selections[0].Data
	if len(data) <= wsRowCellKind {
		return base
	}
	out := base
	switch data[wsRowCellKind] {
	case wsRowKindAll:
		out.Workspace = ""
		out.Namespace = executable.WildcardNamespace
	case wsRowKindWS:
		out.Workspace = data[wsRowCellWsName]
		out.Namespace = executable.WildcardNamespace
	case wsRowKindNS:
		out.Workspace = data[wsRowCellWsName]
		out.Namespace = data[wsRowCellNsName]
	}
	return out
}

// padRow returns a row data slice with a hidden context cell appended at
// `wsRowCellKind`. The first three cells are the visible columns; remaining
// cells are not rendered by the table but survive Selectable.SelectedData().
func padRow(visible []string, kind, wsName, nsName string) []string {
	row := make([]string, wsRowCellKind+3)
	copy(row, visible)
	row[wsRowCellKind] = kind
	row[wsRowCellWsName] = wsName
	row[wsRowCellNsName] = nsName
	return row
}

// withWorkspace overrides the workspace field of a filter and resets the
// namespace to the wildcard so workspace-level counts include every
// executable regardless of namespace.
func withWorkspace(filter Filter, wsName string) Filter {
	out := filter
	out.Workspace = wsName
	out.Namespace = executable.WildcardNamespace
	return out
}

func decodeSelection(table *views.Table) (wsName, nsName string) {
	row := table.GetSelectedRow()
	if row == nil {
		return "", ""
	}
	data := row.Data()
	if len(data) <= wsRowCellKind {
		return "", ""
	}
	return data[wsRowCellWsName], data[wsRowCellNsName]
}

func selectedWorkspace(table *views.Table, wsList workspace.WorkspaceList) *workspace.Workspace {
	wsName, _ := decodeSelection(table)
	if wsName == "" {
		return nil
	}
	return findWorkspace(wsList, wsName)
}

func findWorkspace(wsList workspace.WorkspaceList, wsName string) *workspace.Workspace {
	for _, ws := range wsList {
		if ws.AssignedName() == wsName {
			return ws
		}
	}
	return nil
}
