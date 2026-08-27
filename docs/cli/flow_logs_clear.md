---
title: flow logs clear
description: "Clear execution history and logs."
---

# flow logs clear

Clear execution history and logs.

## Synopsis

Remove execution history records and associated log files. If a ref is provided, only that executable's data is cleared.

```shell
flow logs clear [ref] [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for clear |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow logs](flow_logs.md) — View execution history and logs.
