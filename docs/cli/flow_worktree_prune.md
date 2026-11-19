---
title: flow worktree prune
description: "Remove stale worktree entries and optionally clean up merged branches."
---

# flow worktree prune

Remove stale worktree entries and optionally clean up merged branches.

## Synopsis

Clean up stale worktree entries whose directories no longer exist. With --merged,
also remove worktrees whose branches have been merged into the default branch.

Examples:
  flow worktree prune --dry-run
  flow worktree prune --merged
  flow worktree prune --merged --yes

```shell
flow worktree prune [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--dry-run` |  | Print what the run would do - resolved args, params, dir and command - without running it. Secrets and prompts are not resolved. |
| `-h, --help` |  | help for prune |
| `--merged` |  | Also remove worktrees whose branches have been merged into the default branch. |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-y, --yes` |  | Skip confirmation prompts |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow worktree](flow_worktree.md) — Manage git worktrees for the current repository.
