package internal

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/response"
	"github.com/flowexec/flow/internal/updater"
	"github.com/flowexec/flow/internal/version"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/logger"
)

func RegisterCliCmd(ctx *context.Context, rootCmd *cobra.Command) {
	cliCmd := &cobra.Command{
		Use:   "cli",
		Short: "Manage the flow CLI itself.",
		Long:  "Commands for managing the flow CLI tool (updates, version info, etc.).",
	}

	updateCmd := &cobra.Command{
		Use:     "update",
		Short:   "Update flow to the latest version.",
		Long:    "Check GitHub for a newer version of flow and install it if available.",
		Example: updateExamples,
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			updateFunc(ctx, cmd)
		},
	}
	RegisterFlag(ctx, updateCmd, *flags.CLIVersionFlag)
	RegisterFlag(ctx, updateCmd, *flags.YesFlag)
	RegisterFlag(ctx, updateCmd, *flags.OutputFormatFlag)

	cliCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(cliCmd)
}

func updateFunc(ctx *context.Context, cmd *cobra.Command) {
	current := version.SemVer()
	if current == "" {
		errhandler.HandleFatal(ctx, cmd,
			fmt.Errorf("current version is unknown (dev build); cannot update"))
		return
	}

	targetVersion := flags.ValueFor[string](cmd, *flags.CLIVersionFlag, false)
	if targetVersion != "" {
		logger.Log().Infof("Fetching release info for %s...", targetVersion)
	} else {
		logger.Log().Infof("Checking for updates (current: %s)...", current)
	}

	var info *updater.ReleaseInfo
	var err error
	if targetVersion != "" {
		info, err = updater.ReleaseByTag(targetVersion)
	} else {
		info, err = updater.LatestRelease()
	}
	if err != nil {
		errhandler.HandleFatal(ctx, cmd,
			fmt.Errorf("failed to fetch release info: %w", err))
		return
	}

	// Refresh the cache so the update notice clears on the next invocation.
	_ = updater.RefreshCache(ctx.DataStore, info)

	if strings.TrimPrefix(info.TagName, "v") == current || info.TagName == current {
		response.HandleSuccess(ctx, cmd,
			fmt.Sprintf("Already on %s.", current),
			map[string]any{
				"currentVersion": current,
				"latestVersion":  strings.TrimPrefix(info.TagName, "v"),
			},
		)
		return
	}

	// For latest-release flow, skip if already up to date.
	if targetVersion == "" && !updater.IsNewer(current, info.TagName) {
		response.HandleSuccess(ctx, cmd,
			fmt.Sprintf("Already up to date (%s).", current),
			map[string]any{
				"currentVersion": current,
				"latestVersion":  strings.TrimPrefix(info.TagName, "v"),
			},
		)
		return
	}

	yes := flags.ValueFor[bool](cmd, *flags.YesFlag, false)
	if !yes {
		logger.Log().Infof("Version to install: %s (current: %s)", info.TagName, current)
		if !confirmPrompt(ctx, "Upgrade now? [y/N]: ") {
			logger.Log().Info("Update cancelled.")
			return
		}
	}

	logger.Log().Infof("Upgrading flow to %s...", info.TagName)
	if err := updater.Upgrade(info); err != nil {
		errhandler.HandleFatal(ctx, cmd,
			fmt.Errorf("upgrade failed: %w", err))
		return
	}

	response.HandleSuccess(ctx, cmd,
		fmt.Sprintf("Successfully upgraded flow to %s.", info.TagName),
		map[string]any{
			"previousVersion": current,
			"newVersion":      info.TagName,
		},
	)
}

// confirmPrompt prints prompt and reads a y/Y/yes response from stdin.
func confirmPrompt(ctx *context.Context, prompt string) bool {
	fmt.Fprint(ctx.StdOut(), prompt)
	scanner := bufio.NewScanner(ctx.StdIn())
	if !scanner.Scan() {
		return false
	}
	answer := strings.TrimSpace(strings.ToLower(scanner.Text()))
	return answer == "y" || answer == "yes"
}

const updateExamples = `
  flow cli update                    # check for an update and prompt before installing
  flow cli update --yes              # install the latest version without confirmation
  flow cli update --version v2.1.0   # install a specific version
`
