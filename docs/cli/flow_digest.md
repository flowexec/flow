---
title: flow digest
description: "Gather runs, executables and files into one document."
---

# flow digest

Gather runs, executables and files into one document.

## Synopsis

Resolve each source against run history, the executable cache and the workspace, and print one labeled section per source. Sources are type:value pointers:

  run:<id>         one run's metadata and log tail
  runs:<filter>    a table of matching runs; filter terms are ref:<text>, status:<passed|failed|cancelled|running>, ran:<age> (or since:) and limit:<n>, and bare words match the ref
  exec:<verb>/<id> an executable's definition, script source and last 5 runs
  file:<path>      a file's contents, relative to the workspace root and confined to it

```shell
flow digest SOURCE [SOURCE...] [flags]
```

## Examples

```shell
flow digest run:3f2a9c1e-...                     # one run's metadata and log tail
flow digest runs:"test status:failed ran:7d"     # failed test runs from the last week
flow digest exec:test/unit                       # definition, script and recent runs
flow digest file:./flow.yaml exec:build/binary   # several sources in one document
flow digest runs:status:failed -o json           # structured sections for scripts and agents
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for digest |
| `--max-chars` | `int` | Character budget shared by all sections. Log, file and script text shrink to fit; logs first. (default 40000) |
| `--max-lines` | `int` | Maximum log lines kept per run section. (default 400) |
| `--no-history` |  | Omit recent runs from exec sections. |
| `-o, --output` | `string` | Output format. One of: yaml, json, or tui. |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow](flow.md) — flow is a command line interface designed to make managing and running development workflows easier.
