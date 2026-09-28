---
title: flow config remove verb
description: "Unregister custom executable verbs."
---

# flow config remove verb

Unregister custom executable verbs.

```shell
flow config remove verb NAME [NAME...] [flags]
```

## Examples

```shell
flow config remove verb status
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for verb |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow config remove](flow_config_remove.md) — Remove an entry from a global configuration list.
