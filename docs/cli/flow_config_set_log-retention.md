---
title: flow config set log-retention
description: "Set how many run logs are kept, by age, total size, and count."
---

# flow config set log-retention

Set how many run logs are kept, by age, total size, and count.

## Synopsis

Set how many run logs flow keeps. Each time an executable runs, the oldest logs are removed until every limit is met. Only the flags passed are changed; 0 removes that limit.

Defaults: --max-age 720h0m0s, --max-size-mb 500, --max-count 1000.

```shell
flow config set log-retention [flags]
```

## Examples

```shell
flow config set log-retention --max-age 168h --max-count 200
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for log-retention |
| `--max-age` | `string` | Remove run logs older than this duration (for example 720h). 0 removes the limit. |
| `--max-count` | `int` | Keep at most this many run logs. 0 removes the limit. |
| `--max-size-mb` | `int` | Keep the total size of run logs under this many megabytes. 0 removes the limit. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow config set](flow_config_set.md) — Set a global configuration value.
