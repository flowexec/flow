---
title: Installation
description: "Install the flow CLI on macOS, Linux, or Windows with a one-line script, Homebrew, go install, or a release binary."
---

# Installation

flow is a single binary for macOS, Linux and Windows.

## Install

::: code-group
```shell [macOS / Linux]
curl -sSL https://install.flowexec.io | bash
```

```powershell [Windows (PowerShell)]
irm https://install.flowexec.io/win | iex
```
:::

Then check it's on your `PATH`:

```shell
flow --version
```

On Linux, clipboard support needs `xclip` and desktop notifications need `notify-send` (from
`libnotify`). [Git workspaces](guides/workspaces.md#git-workspaces) need
[Git](https://git-scm.com/).

### Other ways to install

```shell
brew install flowexec/tap/flow               # Homebrew
go install github.com/flowexec/flow/v2@latest # Go
```

Or download a binary from the [releases page](https://github.com/flowexec/flow/releases),
with checksums, and put it on your `PATH`. On Windows, extract `flow.exe` from the `.zip`
for your architecture.

## Shell completion

```bash
flow completion bash > /etc/bash_completion.d/flow                  # Bash
flow completion zsh > ~/.oh-my-zsh/completions/_flow                # Zsh (oh-my-zsh)
flow completion fish > ~/.config/fish/completions/flow.fish         # Fish
flow completion powershell | Out-String | Invoke-Expression         # PowerShell
```

## Updating

```shell
flow cli update                     # check, then confirm before installing
flow cli update --yes               # install the latest without asking
flow cli update --version v2.1.0    # install a specific version
```

The update replaces the binary in place. If you installed with Homebrew, run
`brew upgrade flowexec/tap/flow` instead.

To get a one-line notice after a command when a new release is out, turn on update checks.
They're off by default.

```shell
flow config set update-check true
```

Read [Breaking Changes](breaking-changes) before a major upgrade.

## Next

- [Quickstart](quickstart.md)
- [GitHub Actions](guides/github-actions.md) and [Containers](guides/containers.md), to run flow in CI
