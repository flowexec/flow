---
title: flow template add
description: "Register a flowfile template by name."
---

# flow template add

Register a flowfile template by name.

```shell
flow template add NAME DEFINITION_TEMPLATE_PATH [flags]
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for add |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow template](flow_template.md) — Manage flowfile templates.
