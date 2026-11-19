package mcp

import (
	"github.com/flowexec/flow/v2/internal/services/git"
	"github.com/flowexec/flow/v2/pkg/filesystem"
	"github.com/flowexec/flow/v2/types/config"
	"github.com/flowexec/flow/v2/types/workspace"
)

// resolveAgentInstructions returns the agent instructions for the workspace at wsPath.
// When the workspace has none and the working directory is inside a git worktree, it falls
// back to the parent repository's registered workspace instructions — the worktree shares
// the same repo, so the parent's conventions apply.
func resolveAgentInstructions(ws *workspace.Workspace, cfg *config.Config) string {
	if instructions := ws.ResolvedAgentInstructions(); instructions != "" {
		return instructions
	}
	return parentWorkspaceAgentInstructions(ws.Location(), cfg)
}

// parentWorkspaceAgentInstructions resolves instructions from the main worktree's workspace
// when dir is inside a git worktree whose parent repo has a registered workspace.
func parentWorkspaceAgentInstructions(dir string, cfg *config.Config) string {
	if dir == "" || cfg == nil {
		return ""
	}

	repoRoot, err := git.RepoCommonDir(dir)
	if err != nil {
		return ""
	}

	// If we're already at the repo root, there's no parent to fall back to.
	wtRoot, err := git.WorktreeTopLevel(dir)
	if err != nil || filesystem.NormalizePath(wtRoot) == filesystem.NormalizePath(repoRoot) {
		return ""
	}

	// Find a registered workspace at the parent repo root.
	for name, wsPath := range cfg.Workspaces {
		if filesystem.NormalizePath(wsPath) == filesystem.NormalizePath(repoRoot) {
			parentWs, err := filesystem.LoadWorkspaceConfig(name, wsPath)
			if err != nil {
				return ""
			}
			return parentWs.ResolvedAgentInstructions()
		}
	}
	return ""
}
