---
title: Installation
---

# Installation

> **System Requirements:** flow supports Linux, macOS, and Windows systems. On Linux, you'll need `xclip` installed for clipboard features and `notify-send` (from `libnotify`) for desktop notifications.
>
> **Optional:** [Git](https://git-scm.com/) is required for [git workspace](guides/workspaces.md#git-workspaces) features (`workspace add` from URLs, `workspace update`, `sync --git`).

## Quick Install

::: code-group
```shell [macOS / Linux]
curl -sSL https://raw.githubusercontent.com/flowexec/flow/main/scripts/install.sh | bash
```

```powershell [Windows (PowerShell)]
irm https://raw.githubusercontent.com/flowexec/flow/main/scripts/install.ps1 | iex
```
:::

## Alternative Install Methods

### Homebrew (macOS/Linux)

```shell
brew install flowexec/tap/flow
```

### Go Install

```bash
go install github.com/flowexec/flow@latest
```

### Scoop (Windows)

```powershell
scoop bucket add flowexec https://github.com/flowexec/scoop-bucket
scoop install flow
```

### Manual Download

Download the latest release from the [releases page](https://github.com/flowexec/flow/releases) and add the binary to your `PATH`.
Each release includes checksums for verification.

::: tip Windows
Download the `.zip` archive for your architecture, extract `flow.exe`, and add its directory to your `PATH` environment variable.
:::

## Verify Installation

Check that flow is installed correctly:

```shell
flow --version
```

## Shell Completion

Enable tab completion for your shell:

```bash
# Bash
flow completion bash > /etc/bash_completion.d/flow

# Zsh (oh-my-zsh)
flow completion zsh > ~/.oh-my-zsh/completions/_flow

# Fish
flow completion fish > ~/.config/fish/completions/flow.fish

# PowerShell
flow completion powershell | Out-String | Invoke-Expression
```

## Next Steps

Ready to start automating? → [Quick start guide](quickstart.md)

## CI/CD & Containers

For GitHub Actions, Docker, and other integrations, see the [integrations guide](guides/integrations.md).