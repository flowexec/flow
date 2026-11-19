---
title: flow worktree remove
description: "Remove a worktree."
---

# flow worktree remove

Remove a worktree.

## Synopsis

Remove a git worktree by branch name or path. The main worktree cannot be removed.
Use --force to remove a worktree with uncommitted changes.

Examples:
  flow worktree remove feature/auth
  flow worktree remove .worktrees/feature--auth
  flow worktree remove feature/auth --force

```shell
flow worktree remove BRANCH_OR_PATH [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--force` |  | Force update by discarding local changes (hard reset to remote) |
| `-h, --help` |  | help for remove |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-y, --yes` |  | Skip confirmation prompts |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow worktree](flow_worktree.md) — Manage git worktrees for the current repository.
