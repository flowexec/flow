---
title: Installation
---

# Installation

> [!NOTE] System requirements
> Flow supports Linux, macOS, and Windows systems.
> - Linux: requires `xclip` for clipboard features and `notify-send` (from `libnotify`) for desktop notifications.
> - Optional: [Git](https://git-scm.com/) is required for [git workspace features](guides/workspaces.md#git-workspaces).

## Quick Install

::: code-group
```shell [macOS / Linux]
curl -sSL https://install.flowexec.io | bash
```

```powershell [Windows (PowerShell)]
irm https://install.flowexec.io/win | iex
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

### Manual Download

Download the latest release from the [releases page](https://github.com/flowexec/flow/releases) and add the binary to your `PATH`.
Each release includes checksums for verification.

::: tip Windows
Download the `.zip` archive for your architecture, extract `flow.exe`, and add its directory to your `PATH` environment variable.
:::

## Upgrading

### Using the built-in update command

flow can check for and install updates itself:

```shell
# Check for an update and prompt before installing
flow cli update

# Install the latest version without confirmation
flow cli update --yes

# Install a specific version
flow cli update --version v2.1.0
```

`flow cli update` replaces the running binary in-place — no additional steps needed.

#### Automatic update notifications

flow can check for new releases in the background and print a notice after your next command completes. This is disabled by default and opt-in:

```shell
# Enable background update checks
flow config set update-check true

# Disable them again
flow config set update-check false
```

When a newer version is available, you'll see a one-line notice at the end of your next command's output. Run `flow cli update` to act on it.

### Using your original install method

Re-run the same command you used to install — see [Quick Install](#quick-install) and [Alternative Install Methods](#alternative-install-methods) above. Homebrew uses `brew upgrade flowexec/tap/flow`.

> [!NOTE]
> Review the [Breaking Changes guide](breaking-changes) before upgrading.

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