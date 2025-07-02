package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jahvon/tuikit/views"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/exp/maps"

	"github.com/jahvon/flow/cmd/internal/flags"
	"github.com/jahvon/flow/internal/cache"
	"github.com/jahvon/flow/internal/context"
	"github.com/jahvon/flow/internal/filesystem"
	"github.com/jahvon/flow/internal/io"
	workspaceIO "github.com/jahvon/flow/internal/io/workspace"
	"github.com/jahvon/flow/types/common"
	"github.com/jahvon/flow/types/config"
	"github.com/jahvon/flow/types/workspace"
)

func RegisterWorkspaceCmd(ctx *context.Context, rootCmd *cobra.Command) {
	wsCmd := &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws", "workspaces"},
		Short:   "Manage development workspaces.",
	}
	registerAddWorkspaceCmd(ctx, wsCmd)
	registerSwitchWorkspaceCmd(ctx, wsCmd)
	registerRemoveWorkspaceCmd(ctx, wsCmd)
	registerListWorkspaceCmd(ctx, wsCmd)
	registerGetWorkspaceCmd(ctx, wsCmd)
	rootCmd.AddCommand(wsCmd)
}

func registerAddWorkspaceCmd(ctx *context.Context, wsCmd *cobra.Command) {
	createCmd := &cobra.Command{
		Use:     "add NAME PATH",
		Aliases: []string{"init", "create", "new"},
		Short:   "Initialize a new workspace.",
		Args:    cobra.ExactArgs(2),
		Run:     func(cmd *cobra.Command, args []string) { addWorkspaceFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, createCmd, *flags.SetAfterCreateFlag)
	wsCmd.AddCommand(createCmd)
}

func addWorkspaceFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	logger := ctx.Logger
	name := args[0]
	path := args[1]

	userConfig := ctx.Config
	if _, found := userConfig.Workspaces[name]; found {
		logger.Fatalf("workspace %s already exists at %s", name, userConfig.Workspaces[name])
	}

	switch {
	case path == "":
		path = filepath.Join(filesystem.CachedDataDirPath(), name)
	case path == "." || strings.HasPrefix(path, "./"):
		wd, err := os.Getwd()
		if err != nil {
			logger.FatalErr(err)
		}
		if path == "." {
			path = wd
		} else {
			path = fmt.Sprintf("%s/%s", wd, path[2:])
		}
	case path == "~" || strings.HasPrefix(path, "~/"):
		hd, err := os.UserHomeDir()
		if err != nil {
			logger.FatalErr(err)
		}
		if path == "~" {
			path = hd
		} else {
			path = fmt.Sprintf("%s/%s", hd, path[2:])
		}
	case !filepath.IsAbs(path):
		wd, err := os.Getwd()
		if err != nil {
			logger.FatalErr(err)
		}
		path = fmt.Sprintf("%s/%s", wd, path)
	}

	if !filesystem.WorkspaceConfigExists(path) {
		if err := filesystem.InitWorkspaceConfig(name, path); err != nil {
			logger.FatalErr(err)
		}
	}
	userConfig.Workspaces[name] = path

	set := flags.ValueFor[bool](ctx, cmd, *flags.SetAfterCreateFlag, false)
	if set {
		userConfig.CurrentWorkspace = name
		logger.Infof("Workspace '%s' set as current workspace", name)
	}

	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.FatalErr(err)
	}

	if err := cache.UpdateAll(logger); err != nil {
		logger.FatalErr(errors.Wrap(err, "failure updating cache"))
	}

	logger.PlainTextSuccess(fmt.Sprintf("Workspace '%s' created in %s", name, path))
}

func registerSwitchWorkspaceCmd(ctx *context.Context, setCmd *cobra.Command) {
	workspaceCmd := &cobra.Command{
		Use:     "switch NAME",
		Aliases: []string{"set", "use"},
		Short:   "Switch the current workspace.",
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { switchWorkspaceFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, workspaceCmd, *flags.FixedWsModeFlag)
	setCmd.AddCommand(workspaceCmd)
}

func switchWorkspaceFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	logger := ctx.Logger
	ws := args[0]
	userConfig := ctx.Config
	if _, found := userConfig.Workspaces[ws]; !found {
		logger.Fatalf("workspace %s not found", ws)
	}
	userConfig.CurrentWorkspace = ws
	fixedMode := flags.ValueFor[bool](ctx, cmd, *flags.FixedWsModeFlag, false)
	if fixedMode {
		userConfig.WorkspaceMode = config.ConfigWorkspaceModeFixed
	}

	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.FatalErr(err)
	}
	logger.PlainTextSuccess("Workspace set to " + ws)
}

func registerRemoveWorkspaceCmd(ctx *context.Context, wsCmd *cobra.Command) {
	deleteCmd := &cobra.Command{
		Use:     "remove NAME",
		Aliases: []string{"delete", "rm"},
		Short:   "Remove an existing workspace.",
		Long: "Remove an existing workspace. File contents will remain in the corresponding directory but the " +
			"workspace will be unlinked from the flow global configurations.\nNote: You cannot remove the current workspace.",
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return maps.Keys(ctx.Config.Workspaces), cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) { removeWorkspaceFunc(ctx, cmd, args) },
	}
	wsCmd.AddCommand(deleteCmd)
}

func removeWorkspaceFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	logger := ctx.Logger
	name := args[0]

	form, err := views.NewForm(
		io.Theme(ctx.Config.Theme.String()),
		ctx.StdIn(),
		ctx.StdOut(),
		&views.FormField{
			Key:   "confirm",
			Type:  views.PromptTypeConfirm,
			Title: fmt.Sprintf("Are you sure you want to remove the workspace '%s'?", name),
		})
	if err != nil {
		logger.FatalErr(err)
	}
	if err := form.Run(ctx.Ctx); err != nil {
		logger.FatalErr(err)
	}
	resp := form.FindByKey("confirm").Value()
	if truthy, _ := strconv.ParseBool(resp); !truthy {
		logger.Warnf("Aborting")
		return
	}

	userConfig := ctx.Config
	if name == userConfig.CurrentWorkspace {
		logger.Fatalf("cannot remove the current workspace")
	}
	if _, found := userConfig.Workspaces[name]; !found {
		logger.Fatalf("workspace %s was not found", name)
	}

	delete(userConfig.Workspaces, name)
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.FatalErr(err)
	}

	logger.Warnf("Workspace '%s' deleted", name)

	if err := cache.UpdateAll(logger); err != nil {
		logger.FatalErr(errors.Wrap(err, "unable to update cache"))
	}
}

func registerListWorkspaceCmd(ctx *context.Context, wsCmd *cobra.Command) {
	listCmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all registered workspaces.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { listWorkspaceFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, listCmd, *flags.OutputFormatFlag)
	RegisterFlag(ctx, listCmd, *flags.FilterTagFlag)
	wsCmd.AddCommand(listCmd)
}

func listWorkspaceFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	logger := ctx.Logger
	outputFormat := flags.ValueFor[string](ctx, cmd, *flags.OutputFormatFlag, false)
	tagsFilter := flags.ValueFor[[]string](ctx, cmd, *flags.FilterTagFlag, false)

	workspaceCache, err := ctx.WorkspacesCache.GetLatestData(logger)
	if err != nil {
		logger.Fatalx("failure loading workspace configs from cache", "err", err)
	}

	filteredWorkspaces := make([]*workspace.Workspace, 0)
	for name, ws := range workspaceCache.Workspaces {
		location := workspaceCache.WorkspaceLocations[name]
		ws.SetContext(name, location)
		if !common.Tags(ws.Tags).HasAnyTag(tagsFilter) {
			continue
		}
		filteredWorkspaces = append(filteredWorkspaces, ws)
	}

	if len(filteredWorkspaces) == 0 {
		logger.Fatalf("no workspaces found")
	}

	if TUIEnabled(ctx, cmd) {
		view := workspaceIO.NewWorkspaceListView(ctx, filteredWorkspaces)
		SetView(ctx, cmd, view)
	} else {
		workspaceIO.PrintWorkspaceList(logger, outputFormat, filteredWorkspaces)
	}
}

func registerGetWorkspaceCmd(ctx *context.Context, wsCmd *cobra.Command) {
	viewCmd := &cobra.Command{
		Use:     "get NAME",
		Aliases: []string{"show", "view", "info"},
		Short:   "Get workspace details. If the name is omitted, the current workspace is used.",
		Args:    cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return maps.Keys(ctx.Config.Workspaces), cobra.ShellCompDirectiveNoFileComp
		},
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { getWorkspaceFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, viewCmd, *flags.OutputFormatFlag)
	wsCmd.AddCommand(viewCmd)
}

func getWorkspaceFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	logger := ctx.Logger
	var workspaceName, wsPath string
	if len(args) == 1 {
		workspaceName = args[0]
		wsPath = ctx.Config.Workspaces[workspaceName]
	} else {
		workspaceName = ctx.CurrentWorkspace.AssignedName()
		wsPath = ctx.CurrentWorkspace.Location()
	}

	wsCfg, err := filesystem.LoadWorkspaceConfig(workspaceName, wsPath)
	if err != nil {
		logger.FatalErr(errors.Wrap(err, "failure loading workspace config"))
	} else if wsCfg == nil {
		logger.Fatalf("config not found for workspace %s", workspaceName)
	}

	outputFormat := flags.ValueFor[string](ctx, cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		view := workspaceIO.NewWorkspaceView(ctx, wsCfg)
		SetView(ctx, cmd, view)
	} else {
		workspaceIO.PrintWorkspaceConfig(logger, outputFormat, wsCfg)
	}
}
