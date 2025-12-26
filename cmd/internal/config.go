package internal

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	tuikitIO "github.com/flowexec/tuikit/io"
	"github.com/flowexec/tuikit/views"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/internal/io"
	configIO "github.com/flowexec/flow/internal/io/config"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
	"github.com/flowexec/flow/types/config"
)

func RegisterConfigCmd(ctx *context.Context, rootCmd *cobra.Command) {
	setCmd := &cobra.Command{
		Use:     "config",
		Aliases: []string{"cfg"},
		Short:   "Update flow configuration values.",
	}
	registerConfigResetCmd(ctx, setCmd)
	registerConfigGetCmd(ctx, setCmd)
	registerSetConfigCmd(ctx, setCmd)
	rootCmd.AddCommand(setCmd)
}

func registerConfigResetCmd(ctx *context.Context, configCmd *cobra.Command) {
	resetCmd := &cobra.Command{
		Use:   "reset",
		Short: "Restore the default flow configuration values. This will overwrite the current configuration.",
		Args:  cobra.NoArgs,
		Run:   func(cmd *cobra.Command, args []string) { resetConfigFunc(ctx, cmd, args) },
	}
	configCmd.AddCommand(resetCmd)
}

func resetConfigFunc(ctx *context.Context, _ *cobra.Command, _ []string) {
	form, err := views.NewForm(
		io.Theme(ctx.Config.Theme.String()),
		ctx.StdIn(),
		ctx.StdOut(),
		&views.FormField{
			Key:   "confirm",
			Type:  views.PromptTypeConfirm,
			Title: "This will overwrite your current flow configurations. Are you sure you want to continue?",
		})
	if err != nil {
		logger.Log().FatalErr(err)
	}
	if err := form.Run(ctx); err != nil {
		logger.Log().FatalErr(err)
	}
	resp := form.FindByKey("confirm").Value()
	if truthy, _ := strconv.ParseBool(resp); !truthy {
		logger.Log().Warnf("Aborting")
		return
	}

	if err := filesystem.InitConfig(); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess("Restored flow configurations")
}

func registerSetConfigCmd(ctx *context.Context, configCmd *cobra.Command) {
	setCmd := &cobra.Command{
		Use:     "set",
		Aliases: []string{"update"},
		Short:   "Set a global configuration value.",
	}
	registerSetNamespaceCmd(ctx, setCmd)
	registerSetWorkspaceModeCmd(ctx, setCmd)
	registerSetLogModeCmd(ctx, setCmd)
	registerSetTUICmd(ctx, setCmd)
	registerSetNotificationsCmd(ctx, setCmd)
	registerSetThemeCmd(ctx, setCmd)
	registerSetTimeoutCmd(ctx, setCmd)
	configCmd.AddCommand(setCmd)
}

func registerSetNamespaceCmd(ctx *context.Context, setCmd *cobra.Command) {
	namespaceCmd := &cobra.Command{
		Use:     "namespace NAME",
		Aliases: []string{"ns"},
		Short:   "Change the current namespace.",
		Args:    cobra.ExactArgs(1),
		Run:     func(cmd *cobra.Command, args []string) { setNamespaceFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(namespaceCmd)
}

func setNamespaceFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	namespace := args[0]
	userConfig := ctx.Config
	userConfig.CurrentNamespace = namespace
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess("Namespace set to " + namespace)
}

func registerSetWorkspaceModeCmd(ctx *context.Context, setCmd *cobra.Command) {
	workspaceModeCmd := &cobra.Command{
		Use:       "workspace-mode [fixed|dynamic]",
		Short:     "Switch between fixed and dynamic workspace modes.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"fixed", "dynamic"},
		Run:       func(cmd *cobra.Command, args []string) { setWorkspaceModeFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(workspaceModeCmd)
}

func setWorkspaceModeFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	mode := config.ConfigWorkspaceMode(strings.ToLower(args[0]))

	userConfig := ctx.Config
	if userConfig.Interactive == nil {
		userConfig.Interactive = &config.Interactive{}
	}
	userConfig.WorkspaceMode = mode
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess(fmt.Sprintf("Workspace mode set to '%s'", string(mode)))
}

func registerSetLogModeCmd(ctx *context.Context, setCmd *cobra.Command) {
	logModeCmd := &cobra.Command{
		Use:       "log-mode [logfmt|json|text|hidden]",
		Short:     "Set the default log mode.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"logfmt", "json", "text", "hidden"},
		Run:       func(cmd *cobra.Command, args []string) { setLogModeFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(logModeCmd)
}

func setLogModeFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	mode := tuikitIO.LogMode(strings.ToLower(args[0]))

	userConfig := ctx.Config
	userConfig.DefaultLogMode = mode
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess(fmt.Sprintf("Default log mode set to '%s'", mode))
}

func registerSetTUICmd(ctx *context.Context, setCmd *cobra.Command) {
	tuiCmd := &cobra.Command{
		Use:       "tui [true|false]",
		Short:     "Enable or disable the interactive terminal UI experience.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"true", "false"},
		Run:       func(cmd *cobra.Command, args []string) { setInteractiveFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(tuiCmd)
}

func setInteractiveFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	enabled, err := strconv.ParseBool(args[0])
	if err != nil {
		logger.Log().FatalErr(errors.Wrap(err, "invalid boolean value"))
	}

	userConfig := ctx.Config
	if userConfig.Interactive == nil {
		userConfig.Interactive = &config.Interactive{}
	}
	userConfig.Interactive.Enabled = enabled
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	strVal := "disabled"
	if enabled {
		strVal = "enabled"
	}
	logger.Log().PlainTextSuccess("Interactive UI " + strVal)
}

func registerSetNotificationsCmd(ctx *context.Context, setCmd *cobra.Command) {
	notificationsCmd := &cobra.Command{
		Use:       "notifications [true|false]",
		Short:     "Enable or disable notifications.",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"true", "false"},
		Run:       func(cmd *cobra.Command, args []string) { setNotificationsFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, notificationsCmd, *flags.SetSoundNotificationFlag)
	setCmd.AddCommand(notificationsCmd)
}

func setNotificationsFunc(ctx *context.Context, cmd *cobra.Command, args []string) {
	enabled, err := strconv.ParseBool(args[0])
	if err != nil {
		logger.Log().FatalErr(errors.Wrap(err, "invalid boolean value"))
	}
	sound := flags.ValueFor[bool](cmd, *flags.SetSoundNotificationFlag, false)

	userConfig := ctx.Config
	if userConfig.Interactive == nil {
		userConfig.Interactive = &config.Interactive{}
	}
	userConfig.Interactive.NotifyOnCompletion = &enabled
	if sound {
		userConfig.Interactive.SoundOnCompletion = &enabled
	}
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	strVal := "disabled"
	if enabled {
		strVal = "enabled"
	}
	logger.Log().PlainTextSuccess("Notifications " + strVal)
}

func registerSetThemeCmd(ctx *context.Context, setCmd *cobra.Command) {
	themeCmd := &cobra.Command{
		Use:   "theme [default|dark|light|dracula|tokyo-night]",
		Short: "Set the theme for the TUI views",
		Args:  cobra.ExactArgs(1),
		ValidArgs: []string{
			string(config.ConfigThemeDefault),
			string(config.ConfigThemeDark),
			string(config.ConfigThemeLight),
			string(config.ConfigThemeDracula),
			string(config.ConfigThemeTokyoNight),
		},
		Run: func(cmd *cobra.Command, args []string) { setThemeFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(themeCmd)
}

func setThemeFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	themeName := args[0]

	userConfig := ctx.Config
	if userConfig.Interactive == nil {
		userConfig.Interactive = &config.Interactive{}
	}
	userConfig.Theme = config.ConfigTheme(themeName)
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess("Theme set to " + themeName)
}

func registerSetTimeoutCmd(ctx *context.Context, setCmd *cobra.Command) {
	timeoutCmd := &cobra.Command{
		Use:   "timeout DURATION",
		Short: "Set the default timeout for executables.",
		Args:  cobra.ExactArgs(1),
		Run:   func(cmd *cobra.Command, args []string) { setTimeoutFunc(ctx, cmd, args) },
	}
	setCmd.AddCommand(timeoutCmd)
}

func setTimeoutFunc(ctx *context.Context, _ *cobra.Command, args []string) {
	timeoutStr := args[0]
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		logger.Log().FatalErr(errors.Wrap(err, "invalid duration"))
	}

	userConfig := ctx.Config
	userConfig.DefaultTimeout = timeout
	if err := filesystem.WriteConfig(userConfig); err != nil {
		logger.Log().FatalErr(err)
	}
	logger.Log().PlainTextSuccess("Default timeout set to " + timeoutStr)
}

func registerConfigGetCmd(ctx *context.Context, configCmd *cobra.Command) {
	getCmd := &cobra.Command{
		Use:     "get",
		Aliases: []string{"view", "show", "current"},
		Short:   "Get the current global configuration values.",
		Args:    cobra.NoArgs,
		PreRun:  func(cmd *cobra.Command, args []string) { StartTUI(ctx, cmd) },
		PostRun: func(cmd *cobra.Command, args []string) { WaitForTUI(ctx, cmd) },
		Run:     func(cmd *cobra.Command, args []string) { getConfigFunc(ctx, cmd, args) },
	}
	RegisterFlag(ctx, getCmd, *flags.OutputFormatFlag)
	configCmd.AddCommand(getCmd)
}

func getConfigFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	userConfig := ctx.Config
	outputFormat := flags.ValueFor[string](cmd, *flags.OutputFormatFlag, false)
	if TUIEnabled(ctx, cmd) {
		view := configIO.NewUserConfigView(ctx.TUIContainer, *userConfig)
		SetView(ctx, cmd, view)
	} else {
		configIO.PrintUserConfig(outputFormat, userConfig)
	}
}
