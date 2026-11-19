---
title: flow worktree add
description: "Create a new worktree and branch."
---

# flow worktree add

Create a new worktree and branch.

## Synopsis

Create a new git worktree for the given branch. If the branch does not exist,
it is created from HEAD (or the ref given by --base). If PATH is omitted, the
worktree is placed under the workspace's worktreeDir (default .worktrees/).

Examples:
  flow worktree add feature/auth
  flow worktree add feature/auth --base develop
  flow worktree add bugfix/login ./tmp/login-fix
  flow worktree add release/v2 --no-create-branch

```shell
flow worktree add BRANCH [PATH] [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-b, --base` | `string` | Base branch or commit to create the new branch from. Defaults to HEAD. |
| `--dir` | `string` | Parent directory for the new worktree. Defaults to the workspace's worktreeDir or .worktrees/. |
| `-h, --help` |  | help for add |
| `--no-create-branch` |  | Use an existing branch instead of creating a new one. |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow worktree](flow_worktree.md) — Manage git worktrees for the current repository.
