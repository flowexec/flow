---
title: flow workspace list
description: "List all registered workspaces."
---

# flow workspace list

List all registered workspaces.

```shell
flow workspace list [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--annotation` | `stringArray` | Filter by annotations. Format: 'key=value' for exact value match, or 'key' for presence regardless of value. Repeat the flag for multiple selectors; all selectors must match (AND). |
| `-h, --help` |  | help for list |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-t, --tag` | `stringArray` | Filter by tags. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow workspace](flow_workspace.md) — Manage development workspaces.
