package internal

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/flowexec/tuikit"
	"github.com/flowexec/tuikit/views"
	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	flowIO "github.com/flowexec/flow/v2/internal/io"
	"github.com/flowexec/flow/v2/internal/version"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/pkg/logger"
	"github.com/flowexec/flow/v2/types/executable"
	"github.com/flowexec/flow/v2/types/workspace"
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
	if !flowIO.TTYAttached(ctx.StdIn(), ctx.StdOut()) {
		return false
	}
	if flags.HasFlag(cmd, *flags.OutputFormatFlag) {
		format := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
		if format == "yaml" || format == "yml" || format == "json" {
			return false
		}
		if format == "tui" {
			envDisabled, _ := strconv.ParseBool(os.Getenv(flowIO.DisableInteractiveEnvKey))
			return !envDisabled
		}
	}
	envDisabled, _ := strconv.ParseBool(os.Getenv(flowIO.DisableInteractiveEnvKey))
	return !envDisabled && ctx.Config.ShowTUI()
}

func SetView(ctx *context.Context, cmd *cobra.Command, view tuikit.View) {
	if TUIEnabled(ctx, cmd) {
		if err := ctx.SetView(view); err != nil {
			logger.Log().Fatal("unable to set view", "view", view.Type(), "error", err)
		}
	} else {
		logger.Log().Error("interactive mode is disabled", "view", view.Type())
	}
}

func StartTUI(ctx *context.Context, cmd *cobra.Command) {
	if !TUIEnabled(ctx, cmd) {
		return
	}
	if err := ctx.TUIContainer().Start(); err != nil {
		logger.Log().FatalErr(err)
	}
}

func WaitForTUI(ctx *context.Context, cmd *cobra.Command) {
	if !TUIEnabled(ctx, cmd) {
		return
	}
	ctx.TUIContainer().WaitForExit()
}

func printContext(ctx *context.Context, cmd *cobra.Command) {
	if TUIEnabled(ctx, cmd) {
		logger.Log().Println(logger.Theme(ctx.Config.Theme.String()).
			RenderHeader(ctx.AppName(), version.SemVer(), context.HeaderCtxKey, ctx.String(), 0))
	}
}

func workspaceOrCurrent(ctx *context.Context, workspaceName string) *workspace.Workspace {
	// An empty name, or the name of the workspace already resolved for this command, is answered
	// from the context — which is the only place a workspace discovered from the working
	// directory exists, since it is in no config.
	if workspaceName == "" || ctx.CurrentWorkspaceName() == workspaceName {
		if ctx.CurrentWorkspace == nil {
			return nil
		}
		logger.Log().Debugf("'%s' workspace set", ctx.CurrentWorkspaceName())
		return ctx.CurrentWorkspace
	}

	wsPath, wsFound := ctx.Config.Workspaces[workspaceName]
	if !wsFound {
		return nil
	}
	ws, err := filesystem.LoadWorkspaceConfig(workspaceName, wsPath)
	if err != nil {
		logger.Log().WrapError(err, "unable to load workspace config")
		return nil
	}
	ws.SetContext(workspaceName, wsPath)
	logger.Log().Debugf("'%s' workspace set", workspaceName)
	return ws
}

func loadFlowfileTemplate(ctx *context.Context, name, path string) *executable.Template {
	if name == "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logger.Log().Errorf("flowfile template at %s not found", path)
			return nil
		}
		return loadTemplateFromPath(name, path)
	}

	// Registered templates take precedence over discovered ones for backward compatibility.
	if registeredPath, found := ctx.Config.Templates[name]; found {
		return loadTemplateFromPath(name, registeredPath)
	}

	// Fall back to a workspace-discovered template resolved via the cache.
	tmpl, err := ctx.TemplateCache.GetTemplate(name)
	if err != nil {
		logger.Log().Errorf("template %s not found", name)
		return nil
	}
	return tmpl
}

func loadTemplateFromPath(name, path string) *executable.Template {
	tmpl, err := filesystem.LoadFlowFileTemplate(name, path)
	if err != nil {
		logger.Log().WrapError(err, "unable to load flowfile template")
		return nil
	}
	return tmpl
}

// confirmed asks the user to confirm a destructive action unless --yes was passed. It returns false
// when the user declines. Without a terminal the answer is read as a line from stdin, so a script can
// pipe "yes"; getting no answer at all - stdin closed or empty, as when an agent runs the command - is
// a usage error, so a skipped action never looks like a success.
func confirmed(ctx *context.Context, cmd *cobra.Command, title string) bool {
	if flags.ValueFor[bool](cmd, *flags.YesFlag, false) {
		return true
	}
	var resp string
	if flowIO.TTYAttached(ctx.StdIn(), ctx.StdOut()) {
		form, err := views.NewForm(
			logger.Theme(ctx.Config.Theme.String()),
			ctx.StdIn(),
			ctx.StdOut(),
			&views.FormField{Key: "confirm", Type: views.PromptTypeConfirm, Title: title},
		)
		if err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		if err := form.Run(ctx); err != nil {
			errhandler.HandleFatal(ctx, cmd, err)
		}
		resp = form.FindByKey("confirm").Value()
	} else {
		logger.Log().PlainTextInfo(title)
		line, _ := bufio.NewReader(ctx.StdIn()).ReadString('\n')
		resp = strings.TrimSpace(line)
		if resp == "" {
			errhandler.HandleUsage(ctx, cmd, "no confirmation received; pass --yes to confirm")
			return false
		}
	}
	switch strings.ToLower(resp) {
	case "y", "yes", "true", "1":
		return true
	default:
		logger.Log().Warnf("Aborting")
		return false
	}
}
