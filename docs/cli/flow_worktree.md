---
title: flow worktree
description: "Manage git worktrees for the current repository."
---

# flow worktree

Manage git worktrees for the current repository.

## Synopsis

Manage git worktrees for the current repository. Worktrees are lightweight
working copies that share the same repository but can have different branches
checked out simultaneously. Each worktree is auto-discovered as a flow workspace.

Use 'worktree add' to create a new branch and worktree, 'worktree list' to see
all worktrees with their status, and 'worktree prune' to clean up stale ones.

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for worktree |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow](flow.md) — flow is a command line interface designed to make managing and running development workflows easier.
- [flow worktree add](flow_worktree_add.md) — Create a new worktree and branch.
- [flow worktree list](flow_worktree_list.md) — List all worktrees for the current repository.
- [flow worktree prune](flow_worktree_prune.md) — Remove stale worktree entries and optionally clean up merged branches.
- [flow worktree remove](flow_worktree_remove.md) — Remove a worktree.
