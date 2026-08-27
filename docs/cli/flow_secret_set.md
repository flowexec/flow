---
title: flow secret set
description: "Set a secret in the current vault. If no value is provided, you will be prompted to enter one."
---

# flow secret set

Set a secret in the current vault. If no value is provided, you will be prompted to enter one.

```shell
flow secret set NAME [VALUE] [flags]
```

## Examples

```shell
flow secret set MY_TOKEN              # prompted securely
flow secret set MY_TOKEN s3cr3t       # inline value
flow secret set MY_TOKEN --from-file ./token.txt
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `--file` | `string` | File to read the secret's value from |
| `-h, --help` |  | help for set |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `-V, --vault` | `string` | Vault name to use instead of the current vault. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow secret](flow_secret.md) — Manage secrets stored in a vault.
