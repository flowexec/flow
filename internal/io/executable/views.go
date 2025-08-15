package executable

import (
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/themes"
	"github.com/flowexec/tuikit/types"
	"github.com/flowexec/tuikit/views"

	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/io/common"
	"github.com/flowexec/flow/internal/logger"
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
						logger.Log().Error(err, "executable view runner error")
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
	if len(executables.Items()) == 0 {
		container.HandleError(fmt.Errorf("no workspaces found"))
	}

	selectFunc := func(filterVal string) error {
		s := strings.Split(filterVal, " ")
		if len(s) != 2 {
			return fmt.Errorf("invalid filter value")
		}
		verb, id := s[0], s[1]
		exec, err := executables.FindByVerbAndID(executable.Verb(verb), id)
		if err != nil {
			return fmt.Errorf("executable not found")
		}
		return ctx.SetView(NewExecutableView(ctx, exec, runFunc))
	}

	return views.NewCollectionView(container.RenderState(), executables, types.CollectionFormatList, selectFunc)
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
	if len(templates.Items()) == 0 {
		container.HandleError(fmt.Errorf("no templates found"))
	}

	selectFunc := func(filterVal string) error {
		template := templates.Find(filterVal)
		if template == nil {
			return fmt.Errorf("template %s not found", filterVal)
		}

		return ctx.SetView(NewTemplateView(ctx, template, runFunc))
	}

	return views.NewCollectionView(container.RenderState(), templates, types.CollectionFormatList, selectFunc)
}
