---
title: flow config set container-command
description: "Set the command and global flags a container runtime runs with on this machine."
---

# flow config set container-command

Set the command and global flags a container runtime runs with on this machine.

## Synopsis

Run COMMAND in place of docker or podman for container executables, with any flags after `--` placed before every subcommand. COMMAND must accept that runtime's `run` and `rm` flags, such as a Docker-compatible CLI for docker. Set a runtime's own name with no flags to go back to the default.

```shell
flow config set container-command [docker|podman] COMMAND [-- FLAG...] [flags]
```

## Examples

```shell
flow config set container-command docker docker -- --context colima
flow config set container-command podman podman -- --connection build-vm
flow config set container-command docker /opt/homebrew/bin/docker
flow config set container-command docker docker   # back to the default
```

## Options

| Flag | Type | Description |
|------|------|-------------|
| `-h, --help` |  | help for container-command |

## Options inherited from parent commands

| Flag | Type | Description |
|------|------|-------------|
| `-L, --log-level` | `string` | Log verbosity level (debug, info, fatal) (default "info") |
| `--sync` |  | Sync flow cache and workspaces |

## See also

- [flow config set](flow_config_set.md) — Set a global configuration value.
