---
title: flow secret link
description: "Link a name in the current vault to a secret in an external provider."
---

# flow secret link

Link a name in the current vault to a secret in an external provider.

## Synopsis

Point NAME at REFERENCE, a path the vault's provider understands -- an
op:// URI, a pass entry path, an SSM parameter name. Reading NAME reads
through to that secret; nothing is copied and nothing is written back.

Only external vaults hold links.

```shell
flow secret link NAME REFERENCE [flags]
```

## Examples

```shell
flow secret link aws-access-key 'op://Team/AWS/access_key_id'
flow secret link db-password 'team/db/password'
flow secret link api-token '/prod/service-a/api-token'
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for link |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-V, --vault` | `string` | Vault name to use instead of the current vault. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow secret](flow_secret.md) — Manage secrets stored in a vault.
