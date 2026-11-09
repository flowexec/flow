package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/v2/cmd/internal/errors"
	"github.com/flowexec/flow/v2/cmd/internal/flags"
	"github.com/flowexec/flow/v2/cmd/internal/response"
	"github.com/flowexec/flow/v2/internal/services/git"
	"github.com/flowexec/flow/v2/pkg/cache"
	"github.com/flowexec/flow/v2/pkg/context"
	"github.com/flowexec/flow/v2/pkg/logger"
)

func RegisterSyncCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "sync",
		Short:   "Refresh workspace cache and discover new executables.",
		Example: syncExamples,
		Long: "Pull git-sourced workspaces from their remotes, then refresh the workspace cache and " +
			"discover new executables. A pull only fast-forwards; a workspace that can't be pulled is " +
			"named and skipped. Use --no-git to rescan without touching the network, and --force to " +
			"discard local changes and hard reset git workspaces to their remotes.",
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			printContext(ctx, cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			syncFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.GitPullFlag)
	RegisterFlag(ctx, subCmd, *flags.NoGitPullFlag)
	RegisterFlag(ctx, subCmd, *flags.ForceFlag)
	if err := subCmd.Flags().MarkDeprecated(flags.GitPullFlag.Name, "sync pulls git workspaces by default"); err != nil {
		logger.Log().FatalErr(err)
	}
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func syncFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	noGit := flags.ValueFor[bool](cmd, *flags.NoGitPullFlag, false)
	force := flags.ValueFor[bool](cmd, *flags.ForceFlag, false)

	switch {
	case noGit && flags.ValueFor[bool](cmd, *flags.GitPullFlag, false):
		errhandler.HandleUsage(ctx, cmd, "--git and --no-git are mutually exclusive")
	case noGit && force:
		errhandler.HandleUsage(ctx, cmd, "--force resets git workspaces, so it can't be used with --no-git")
	}

	start := time.Now()

	var pullFailures []string
	if !noGit && hasGitWorkspaces(ctx) {
		pullFailures = pullGitWorkspaces(ctx, force)
	}

	if err := cache.UpdateAll(ctx.DataStore); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	duration := time.Since(start)
	msg := fmt.Sprintf("Synced flow cache (%s)", duration.Round(time.Second))
	if len(pullFailures) > 0 {
		msg += fmt.Sprintf(" — git pull failed for: %s", strings.Join(pullFailures, ", "))
	}
	response.HandleSuccess(ctx, cmd, msg, map[string]any{
		"duration":     duration.Round(time.Second).String(),
		"pullFailures": pullFailures,
	})
}

// PullGitWorkspacesForSync pulls git-sourced workspaces for the global --sync flag. It never resets
// local changes, and reports failures as warnings: a run asked to sync should still run, but must not
// look current when it isn't.
func PullGitWorkspacesForSync(ctx *context.Context) {
	if !hasGitWorkspaces(ctx) {
		return
	}
	if failed := pullGitWorkspaces(ctx, false); len(failed) > 0 {
		logger.Log().Warnf(
			"git workspaces not updated, so they may run stale code: %s (see 'workspace update <name>')",
			strings.Join(failed, ", "),
		)
	}
}

func hasGitWorkspaces(ctx *context.Context) bool {
	if len(ctx.Config.GitWorkspaces) > 0 {
		return true
	}
	for _, path := range ctx.Config.Workspaces {
		if git.IsManagedClone(path) {
			return true
		}
	}
	return false
}

func pullGitWorkspaces(ctx *context.Context, force bool) []string {
	if err := git.EnsureInstalled(); err != nil {
		logger.Log().Errorf("Cannot pull git workspaces: %v", err)
		return []string{"<all>"}
	}

	var failed []string
	cfg := ctx.Config
	for name, path := range cfg.Workspaces {
		wsCfg, err := prepareGitWorkspacePull(ctx, name, path)
		if err != nil {
			logger.Log().Errorf("Skipping workspace '%s': %v", name, err)
			failed = append(failed, name)
			continue
		}
		if wsCfg.GitRemote == "" {
			continue
		}

		logger.Log().Infof("Pulling workspace '%s' from %s...", name, wsCfg.GitRemote)
		pullStart := time.Now()

		var pullErr error
		if force {
			pullErr = git.ResetPull(path, wsCfg.GitRef, string(wsCfg.GitRefType))
		} else {
			pullErr = git.Pull(path, wsCfg.GitRef, string(wsCfg.GitRefType))
		}

		if pullErr != nil {
			logger.Log().Errorf("Failed to pull workspace '%s': %v", name, pullErr)
			if !force {
				logger.Log().Warnf("Hint: if '%s' has local changes blocking the pull, use --force to hard reset to remote", name)
			}
			failed = append(failed, name)
			continue
		}

		pullDuration := time.Since(pullStart)
		logger.Log().Infof("Workspace '%s' updated (%s)", name, pullDuration.Round(time.Millisecond))
	}
	return failed
}

const syncExamples = `
  flow sync               # pull git-sourced workspaces, then rescan for executables
  flow sync --no-git      # rescan only, without touching the network
  flow sync --force       # hard-reset git workspaces to their remotes, then rescan
`
