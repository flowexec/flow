---
title: flow config set container-runtime
description: "Set the container runtime that executables set to auto use on this machine."
---

# flow config set container-runtime

Set the container runtime that executables set to auto use on this machine.

## Synopsis

Set the runtime a container executable with `runtime: auto` (the default) runs with on this machine. `auto` detects one: docker, else podman, preferring an engine that is running when both are installed. An executable that names docker or podman keeps that runtime.

```shell
flow config set container-runtime [auto|docker|podman] [flags]
```

## Examples

```shell
flow config set container-runtime podman
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for container-runtime |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync the executable cache first, pulling git-sourced workspaces from their remotes |

## See also

- [flow config set](flow_config_set.md) — Set a global configuration value.
