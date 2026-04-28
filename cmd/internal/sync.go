package internal

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	errhandler "github.com/flowexec/flow/cmd/internal/errors"
	"github.com/flowexec/flow/cmd/internal/flags"
	"github.com/flowexec/flow/cmd/internal/response"
	"github.com/flowexec/flow/internal/services/git"
	"github.com/flowexec/flow/pkg/cache"
	"github.com/flowexec/flow/pkg/context"
	"github.com/flowexec/flow/pkg/filesystem"
	"github.com/flowexec/flow/pkg/logger"
)

func RegisterSyncCmd(ctx *context.Context, rootCmd *cobra.Command) {
	subCmd := &cobra.Command{
		Use:     "sync",
		Short:   "Refresh workspace cache and discover new executables.",
		Example: syncExamples,
		Long: "Refresh the workspace cache and discover new executables. " +
			"Use --git to also pull latest changes for all git-sourced workspaces before syncing. " +
			"Use --force with --git to discard local changes and hard reset to the remote.",
		Args: cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			printContext(ctx, cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			syncFunc(ctx, cmd, args)
		},
	}
	RegisterFlag(ctx, subCmd, *flags.GitPullFlag)
	RegisterFlag(ctx, subCmd, *flags.ForceFlag)
	RegisterFlag(ctx, subCmd, *flags.OutputFormatFlag)
	rootCmd.AddCommand(subCmd)
}

func syncFunc(ctx *context.Context, cmd *cobra.Command, _ []string) {
	pullGit := flags.ValueFor[bool](cmd, *flags.GitPullFlag, false)
	force := flags.ValueFor[bool](cmd, *flags.ForceFlag, false)

	if force && !pullGit {
		errhandler.HandleUsage(ctx, cmd, "--force can only be used with --git")
	}

	start := time.Now()

	var pullFailures []string
	if pullGit {
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

func pullGitWorkspaces(ctx *context.Context, force bool) []string {
	if err := git.EnsureInstalled(); err != nil {
		logger.Log().Errorf("Cannot pull git workspaces: %v", err)
		return []string{"<all>"}
	}

	var failed []string
	cfg := ctx.Config
	for name, path := range cfg.Workspaces {
		wsCfg, err := filesystem.LoadWorkspaceConfig(name, path)
		if err != nil {
			logger.Log().Warnf("Skipping workspace '%s': %v", name, err)
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
  flow sync               # rescan all workspaces for new executables
  flow sync --git         # pull all git-sourced workspaces, then rescan
  flow sync --git --force # hard-reset git workspaces before rescan
`
