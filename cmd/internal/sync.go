package internal

import (
	"fmt"
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
		Use:   "sync",
		Short: "Refresh workspace cache and discover new executables.",
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

	if pullGit {
		pullGitWorkspaces(ctx, force)
	}

	if err := cache.UpdateAll(ctx.DataStore); err != nil {
		errhandler.HandleFatal(ctx, cmd, err)
	}
	duration := time.Since(start)
	response.HandleSuccess(ctx, cmd, fmt.Sprintf("Synced flow cache (%s)", duration.Round(time.Second)), map[string]any{
		"duration": duration.Round(time.Second).String(),
	})
}

func pullGitWorkspaces(ctx *context.Context, force bool) {
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
				logger.Log().Warnf("Hint: use --force to discard local changes and hard reset to remote")
			}
			continue
		}

		pullDuration := time.Since(pullStart)
		logger.Log().Infof("Workspace '%s' updated (%s)", name, pullDuration.Round(time.Millisecond))
	}
}
