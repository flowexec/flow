---
title: flow config add verb
description: "Register custom executable verbs."
---

# flow config add verb

Register custom executable verbs.

## Synopsis

Register one or more custom executable verbs. Custom verbs are standalone (they do not join a
built-in alias group) and can be used in flow files and on the command line like any built-in verb.

Custom verbs are stored in your user config, so anyone running a flow file that uses one needs to
register it too.

```shell
flow config add verb NAME [NAME...] [flags]
```

## Examples

```shell
flow config add verb status
flow config add verb status health
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

- [flow config add](flow_config_add.md) — Add an entry to a global configuration list.
