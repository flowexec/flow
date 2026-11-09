---
title: flow workspace checkout
description: "Bind a git workspace to the local clone its executables operate on."
---

# flow workspace checkout

Bind a git workspace to the local clone its executables operate on.

## Synopsis

A git workspace can hold only flow files for another repository. Bind it to your local clone of that repository, and an executable's dir starting with '@/' resolves there, with the clone's path in FLOW_CHECKOUT_DIR. When a run starts inside that repository or one of its worktrees, the caller's own working tree is used instead. The binding is stored in your user config.

With only NAME, prints the current binding.

Examples:
  flow workspace checkout ops ~/dev/app
  flow workspace checkout ops
  flow workspace checkout ops --clear

```shell
flow workspace checkout NAME [PATH] [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--clear` |  | Remove the workspace's checkout binding. |
| `-h, --help` |  | help for checkout |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow workspace](flow_workspace.md) — Manage development workspaces.
