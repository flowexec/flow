---
title: flow sync
description: "Refresh workspace cache and discover new executables."
---

# flow sync

Refresh workspace cache and discover new executables.

## Synopsis

Pull git-sourced workspaces from their remotes, then refresh the workspace cache and discover new executables. A pull only fast-forwards; a workspace that can't be pulled is named and skipped. Use --no-git to rescan without touching the network, and --force to discard local changes and hard reset git workspaces to their remotes.

```shell
flow sync [flags]
```

## Examples

```shell
flow sync               # pull git-sourced workspaces, then rescan for executables
flow sync --no-git      # rescan only, without touching the network
flow sync --force       # hard-reset git workspaces to their remotes, then rescan
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--force` |  | Force update by discarding local changes (hard reset to remote) |
| `-h, --help` |  | help for sync |
| `--no-git` |  | Rescan without pulling git-sourced workspaces |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow](flow.md) — flow is a command line interface designed to make managing and running development workflows easier.
