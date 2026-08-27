---
title: flow secret remove
description: "Remove a secret from the vault."
---

# flow secret remove

Remove a secret from the vault.

```shell
flow secret remove NAME [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for remove |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-V, --vault` | `string` | Vault name to use instead of the current vault. |
| `-y, --yes` |  | Skip confirmation prompts |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow secret](flow_secret.md) — Manage secrets stored in a vault.
