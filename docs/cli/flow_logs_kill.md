---
title: flow logs kill
description: "Terminate a running background process by run ID."
---

# flow logs kill

Terminate a running background process by run ID.

## Synopsis

Send a termination signal to a running background process identified by its run ID.

```shell
flow logs kill RUN_ID [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for kill |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow logs](flow_logs.md) — View execution history and logs.
