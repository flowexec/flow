package internal

import (
	"os"
	"strconv"

	"github.com/flowexec/tuikit"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/context"
	"github.com/flowexec/flow/internal/filesystem"
	flowIO "github.com/flowexec/flow/internal/io"
	"github.com/flowexec/flow/internal/logger"
	"github.com/flowexec/flow/types/executable"
	"github.com/flowexec/flow/types/workspace"
)

func RegisterFlag(ctx *context.Context, cmd *cobra.Command, flag flags.Metadata) {
	flagSet, err := flags.ToPflag(cmd, flag, false)
	if err != nil {
		logger.Log().FatalErr(err)
	}
	cmd.Flags().AddFlagSet(flagSet)
	if flag.Required {
		MarkFlagRequired(ctx, cmd, flag.Name)
	}
}

func RegisterPersistentFlag(_ *context.Context, cmd *cobra.Command, flag flags.Metadata) {
	flagSet, err := flags.ToPflag(cmd, flag, true)
	if err != nil {
		logger.Log().FatalErr(err)
	}
	cmd.PersistentFlags().AddFlagSet(flagSet)
}

func MarkFlagRequired(_ *context.Context, cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagRequired(name); err != nil {
		logger.Log().FatalErr(err)
	}
}

func MarkFlagMutuallyExclusive(cmd *cobra.Command, names ...string) {
	cmd.MarkFlagsMutuallyExclusive(names...)
}

func MarkOneFlagRequired(cmd *cobra.Command, names ...string) {
	cmd.MarkFlagsOneRequired(names...)
}

func MarkFlagFilename(_ *context.Context, cmd *cobra.Command, name string) {
	if err := cmd.MarkFlagFilename(name); err != nil {
		logger.Log().FatalErr(err)
	}
}

func TUIEnabled(ctx *context.Context, cmd *cobra.Command) bool {
	formatDisabled := false
	if flags.HasFlag(cmd, *flags.OutputFormatFlag) {
		format := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
		formatDisabled = format == "yaml" || format == "yml" || format == "json"
	}
	envDisabled, _ := strconv.ParseBool(os.Getenv(flowIO.DisableInteractiveEnvKey))
	return !formatDisabled && !envDisabled && ctx.Config.ShowTUI()
}

func SetView(ctx *context.Context, cmd *cobra.Command, view tuikit.View) {
	if TUIEnabled(ctx, cmd) {
		if err := ctx.SetView(view); err != nil {
			logger.Log().Fatalx("unable to set view", "view", view.Type(), "error", err)
		}
	} else {
		logger.Log().Errorx("interactive mode is disabled", "view", view.Type())
	}
}

func StartTUI(ctx *context.Context, cmd *cobra.Command) {
	if !TUIEnabled(ctx, cmd) {
		return
	}
	if err := ctx.TUIContainer.Start(); err != nil {
		logger.Log().FatalErr(err)
	}
}

func WaitForTUI(ctx *context.Context, cmd *cobra.Command) {
	if !TUIEnabled(ctx, cmd) {
		return
	}
	ctx.TUIContainer.WaitForExit()
}

func printContext(ctx *context.Context, cmd *cobra.Command) {
	if TUIEnabled(ctx, cmd) {
		logger.Log().Println(flowIO.Theme(ctx.Config.Theme.String()).
			RenderHeader(context.AppName, context.HeaderCtxKey, ctx.String(), 0))
	}
}

func workspaceOrCurrent(ctx *context.Context, workspaceName string) *workspace.Workspace {
	var ws *workspace.Workspace
	if workspaceName == "" {
		ws = ctx.CurrentWorkspace
		workspaceName = ws.AssignedName()
	} else {
		wsPath, wsFound := ctx.Config.Workspaces[workspaceName]
		if !wsFound {
			return nil
		}
		var err error
		ws, err = filesystem.LoadWorkspaceConfig(workspaceName, wsPath)
		if err != nil {
			logger.Log().Error(err, "unable to load workspace config")
		}
		ws.SetContext(workspaceName, wsPath)
	}
	logger.Log().Debugf("'%s' workspace set", workspaceName)
	return ws
}

func loadFlowfileTemplate(ctx *context.Context, name, path string) *executable.Template {
	if name != "" {
		if ctx.Config.Templates == nil {
			logger.Log().Errorf("template %s not found", name)
			return nil
		}
		var found bool
		if path, found = ctx.Config.Templates[name]; !found {
			logger.Log().Errorf("template %s not found", name)
			return nil
		}
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Log().Errorf("flowfile template at %s not found", path)
			return nil
		}
	}
	tmpl, err := filesystem.LoadFlowFileTemplate(name, path)
	if err != nil {
		logger.Log().Error(err, "unable to load flowfile template")
		return nil
	}
	return tmpl
}
