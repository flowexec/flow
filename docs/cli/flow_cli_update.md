---
title: flow cli update
description: "Update flow to the latest version."
---

# flow cli update

Update flow to the latest version.

## Synopsis

Check GitHub for a newer version of flow and install it if available.

```shell
flow cli update [flags]
```

## Examples

```shell
flow cli update                    # check for an update and prompt before installing
flow cli update --yes              # install the latest version without confirmation
flow cli update --version v2.1.0   # install a specific version
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for update |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |
| `--version` | `string` | Target version to install (e.g. v2.1.0). Defaults to the latest release. |
| `-y, --yes` |  | Skip confirmation prompts |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow cli](flow_cli.md) — Manage the flow CLI itself.
