---
title: Installation
---

# Installation

> **System Requirements:** flow supports Linux and macOS systems. On Linux, you'll need `xclip` installed to use clipboard features.

## Quick Install

Run this command to install the latest version of flow:

```shell
curl -sSL https://raw.githubusercontent.com/flowexec/flow/main/scripts/install.sh | bash
```

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

Download the latest release from the [releases page](https://github.com/flowexec/flow/releases) and add the binary to your `$PATH`.
Each release includes checksums for verification.

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
```

## Next Steps

Ready to start automating? → [Quick start guide](quickstart.md)

## CI/CD & Containers

For GitHub Actions, Docker, and other integrations, see the [integrations guide](guides/integrations.md).