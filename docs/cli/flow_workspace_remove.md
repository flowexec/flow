---
title: flow workspace remove
description: "Remove an existing workspace."
---

# flow workspace remove

Remove an existing workspace.

## Synopsis

Remove an existing workspace. File contents will remain in the corresponding directory but the workspace will be unlinked from the flow global configurations.
Note: You cannot remove the current workspace.

```shell
flow workspace remove NAME [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for remove |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-y, --yes` |  | Skip confirmation prompts |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow workspace](flow_workspace.md) — Manage development workspaces.
