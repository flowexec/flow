---
title: flow logs attach
description: "Stream log output from a running background process by run ID."
---

# flow logs attach

Stream log output from a running background process by run ID.

## Synopsis

Stream the log output of a background process identified by its run ID.

```shell
flow logs attach RUN_ID [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for attach |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow logs](flow_logs.md) — View execution history and logs.
