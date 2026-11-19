---
title: flow worktree list
description: "List all worktrees for the current repository."
---

# flow worktree list

List all worktrees for the current repository.

```shell
flow worktree list [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for list |
| `--no-size` |  | Skip measuring each worktree's disk usage, which walks every file and is slow on large checkouts. |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow worktree](flow_worktree.md) — Manage git worktrees for the current repository.
