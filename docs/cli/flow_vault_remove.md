---
title: flow vault remove
description: "Remove an existing vault."
---

# flow vault remove

Remove an existing vault.

## Synopsis

Remove an existing vault by its name. The vault's encrypted secret data remains on disk at its storage path, but its configuration is deleted so flow no longer tracks it.
Note: You cannot remove the current vault.

```shell
flow vault remove NAME [flags]
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

- [flow vault](flow_vault.md) — Manage sensitive secret stores.
