package executable

import (
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/executable"
)

func sortByID(a, b *executable.Executable) int {
	return strings.Compare(a.ID(), b.ID())
}

func NewExecutableView(
	ctx *context.Context,
	exec *executable.Executable,
	runFunc func(string) error,
) tuikit.View {
	container := ctx.TUIContainer
	var executableKeyCallbacks = []types.KeyCallback{
		{
			Key: "r", Label: "run",
			Callback: func() error {
				ctx.TUIContainer.Shutdown(func() {
					err := runFunc(exec.Ref().String())
					if err != nil {
						logger.Log().WrapError(err, "executable view runner error")
					}
				})
				return nil
			},
		},
		{
			Key: "c", Label: "copy ref",
			Callback: func() error {
				common.CopyToClipboard(container, exec.Ref().String(), "copied reference to clipboard")
				return nil
			},
		},
		{
			Key: "e", Label: "edit",
			Callback: func() error {
				if err := common.OpenInEditor(exec.FlowFilePath(), ctx.StdIn(), ctx.StdOut()); err != nil {
					container.HandleError(fmt.Errorf("unable to open executable: %w", err))
				}
				return nil
			},
		},
	}
	opts := executableDetailOpts(exec)
	return views.NewDetailContentView(
		container.RenderState(),
		opts,
		executableKeyCallbacks...,
	)
}

func NewExecutableListView(
	ctx *context.Context,
	executables executable.ExecutableList,
	runFunc func(string) error,
) tuikit.View {
	container := ctx.TUIContainer
	if len(executables) == 0 {
		container.HandleError(fmt.Errorf("no executables found"))
	}

	slices.SortFunc(executables, sortByID)

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Executables (%d)", len(executables)), Percentage: 40},
		{Title: "Flowfile", Percentage: 30},
		{Title: "Tags", Percentage: 30},
	}
	rows := make([]views.TableRow, 0, len(executables))
	for _, ex := range executables {
		tags := common.ColorizeTags(ex.Tags)
		flowfile := filepath.Base(ex.FlowFilePath())
		rows = append(rows, views.TableRow{
			Data: []string{fmt.Sprintf("%s %s", ex.Verb, ex.Ref().ID()), flowfile, tags},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	selectedExec := func() *executable.Executable {
		row := table.GetSelectedRow()
		if row == nil {
			return nil
		}
		data := row.Data()
		if len(data) < 1 {
			return nil
		}
		parts := strings.SplitN(data[0], " ", 2)
		if len(parts) < 2 {
			return nil
		}
		ex, err := executables.FindByVerbAndID(executable.Verb(parts[0]), parts[1])
		if err != nil {
			return nil
		}
		return ex
	}
	table.SetOnSelect(func(_ int) error {
		ex := selectedExec()
		if ex == nil {
			return fmt.Errorf("no executable selected")
		}
		return ctx.SetView(NewExecutableView(ctx, ex, runFunc))
	})
	table.SetKeyCallbacks([]types.KeyCallback{
		{Key: "r", Label: "run", Callback: func() error {
			ex := selectedExec()
			if ex == nil {
				container.SetNotice("no executable selected", themes.OutputLevelError)
				return nil
			}
			container.Shutdown(func() {
				if err := runFunc(ex.Ref().String()); err != nil {
					logger.Log().WrapError(err, "executable list view runner error")
				}
			})
			return nil
		}},
	})
	return table
}

func NewTemplateView(
	ctx *context.Context,
	template *executable.Template,
	runFunc func(string) error,
) tuikit.View {
	container := ctx.TUIContainer
	var templateKeyCallbacks = []types.KeyCallback{
		{
			Key: "r", Label: "run",
			Callback: func() error {
				ctx.TUIContainer.Shutdown()
				return runFunc(template.Name())
			},
		},
		{
			Key: "c", Label: "copy location",
			Callback: func() error {
				common.CopyToClipboard(container, template.Location(), "copied location to clipboard")
				return nil
			},
		},
		{
			Key: "e", Label: "edit",
			Callback: func() error {
				if err := common.OpenInEditor(template.Location(), ctx.StdIn(), ctx.StdOut()); err != nil {
					container.HandleError(fmt.Errorf("unable to open template: %w", err))
				}
				return nil
			},
		},
	}
	opts := templateDetailOpts(template)
	return views.NewDetailContentView(
		container.RenderState(),
		opts,
		templateKeyCallbacks...,
	)
}

func NewTemplateListView(
	ctx *context.Context,
	templates executable.TemplateList,
	runFunc func(string) error,
) tuikit.View {
	container := ctx.TUIContainer
	if len(templates) == 0 {
		container.HandleError(fmt.Errorf("no templates found"))
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name() < templates[j].Name()
	})

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Templates (%d)", len(templates)), Percentage: 50},
		{Title: "Location", Percentage: 50},
	}
	rows := make([]views.TableRow, 0, len(templates))
	for _, t := range templates {
		rows = append(rows, views.TableRow{
			Data: []string{t.Name(), common.ShortenPath(t.Location())},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil || len(row.Data()) == 0 {
			return fmt.Errorf("no template selected")
		}
		t := templates.Find(row.Data()[0])
		if t == nil {
			return fmt.Errorf("template %s not found", row.Data()[0])
		}
		return ctx.SetView(NewTemplateView(ctx, t, runFunc))
	})
	return table
}
