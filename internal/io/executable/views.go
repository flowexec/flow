package executable

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
	flowCommon "github.com/flowexec/flow/types/common"
	"github.com/flowexec/flow/types/executable"
)

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
				if err := clipboard.WriteAll(exec.Ref().String()); err != nil {
					container.HandleError(fmt.Errorf("unable to copy reference to clipboard: %w", err))
				} else {
					container.SetNotice("copied reference to clipboard", themes.OutputLevelInfo)
				}
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
	return views.NewEntityView(
		container.RenderState(),
		exec,
		types.EntityFormatDocument,
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

	columns := []views.TableColumn{
		{Title: "Verb", Percentage: 20},
		{Title: fmt.Sprintf("Executables (%d)", len(executables)), Percentage: 60},
		{Title: "Tags", Percentage: 20},
	}
	rows := make([]views.TableRow, 0, len(executables))
	for _, ex := range executables {
		tags := ""
		if len(ex.Tags) > 0 {
			tags = flowCommon.Tags(ex.Tags).PreviewString()
		}
		rows = append(rows, views.TableRow{
			Data: []string{string(ex.Verb), ex.Ref().ID(), tags},
		})
	}
	table := views.NewTable(container.RenderState(), columns, rows, views.TableDisplayMini)
	table.SetOnSelect(func(_ int) error {
		row := table.GetSelectedRow()
		if row == nil {
			return fmt.Errorf("no executable selected")
		}
		data := row.Data()
		if len(data) < 2 {
			return fmt.Errorf("invalid selection")
		}
		ex, err := executables.FindByVerbAndID(executable.Verb(data[0]), data[1])
		if err != nil {
			return fmt.Errorf("executable not found")
		}
		return ctx.SetView(NewExecutableView(ctx, ex, runFunc))
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
				if err := clipboard.WriteAll(template.Location()); err != nil {
					container.HandleError(fmt.Errorf("unable to copy location to clipboard: %w", err))
				} else {
					container.SetNotice("copied location to clipboard", themes.OutputLevelInfo)
				}
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
	return views.NewEntityView(
		container.RenderState(),
		template,
		types.EntityFormatDocument,
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

	columns := []views.TableColumn{
		{Title: fmt.Sprintf("Templates (%d)", len(templates)), Percentage: 100},
	}
	rows := make([]views.TableRow, 0, len(templates))
	for _, t := range templates {
		rows = append(rows, views.TableRow{Data: []string{t.Name()}})
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
