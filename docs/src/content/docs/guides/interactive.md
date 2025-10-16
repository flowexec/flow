---
title: Interactive UI
---

# Interactive UI

flow provides both a powerful terminal user interface (TUI) and flexible command-line options to fit your workflow. 
This guide covers using the interactive browser, customizing the experience, and working with different output formats.

## Using the TUI Browser

The `flow browse` command launches an interactive browser for discovering and running executables.

### Basic Navigation

```shell
flow browse  # Launch the interactive browser
```

**Keyboard shortcuts:**
- <kbd>↑</kbd> / <kbd>↓</kbd> - Move up/down through the list
- <kbd>←</kbd> / <kbd>→</kbd> - Navigate between panels (workspaces, executables)
- <kbd>Enter</kbd> - Select the highlighted workspace or executable
- <kbd>Space</kbd> - Toggle the namespace list for the selected workspace
- <kbd>Tab</kbd> - Toggle the executable detail viewer
- <kbd>R</kbd> - Run the selected executable (when applicable)
- <kbd>H</kbd> - Show help menu with all shortcuts
- <kbd>Q</kbd> - Quit the browser

### Filtering and Search

Filter executables by various criteria:

```shell
# Filter by workspace
flow browse --workspace api-service

# Filter by namespace
flow browse --namespace deployment

# Filter by verb
flow browse --verb deploy

# Filter by tags
flow browse --tag production --tag critical

# Search by name or description
flow browse --filter "database backup"

# Show executables from all namespaces (not just current)
flow browse --all
```

**Combine filters for precise results:**
```shell
flow browse --workspace api --verb deploy --tag production
```

### Running Executables

**From the browser:**
- Select an executable and press <kbd>R</kbd> to run it
- Arguments and prompts will be handled interactively

**Direct execution:**
```shell
# View specific executable details
flow browse deploy api:production

# Run without browsing
flow deploy api:production
```

## Output Formats

Control how flow displays information with output format options.

### TUI vs Non-Interactive

```shell
# Interactive TUI (default)
flow browse
flow workspace list
flow secret list

# Simple list output
flow browse --output json 
flow workspace list --output json
flow secret list --output yaml
```

### Disabling the TUI

For scripts, CI/CD, or personal preference:

```shell
# Permanently disable TUI
flow config set tui false

# Temporarily disable with environment variable
DISABLE_FLOW_INTERACTIVE=true flow browse
```

## Customization

### Themes

Choose from several built-in themes:

```shell
# Available themes
flow config set theme default      # Everforest (default)
flow config set theme light        # Light theme
flow config set theme dark         # Dark theme  
flow config set theme dracula      # Dracula
flow config set theme tokyo-night  # Tokyo Night
```

### Custom Colors

Override theme colors by editing your config file:

```yaml
# In your flow config file
colorOverride:
  primary: "#83C092"
  secondary: "#D699B6"
  background: "#2D353B"
  border: "#7FBBB3"
  # See config reference for all options
```

> **Complete reference**: See the [config file reference](../types/config.md#ColorPalette) for all color options.

### Notifications

Get notified when long-running executables complete:

```shell
# Enable desktop notifications
flow config set notifications true

# Enable notification sound
flow config set notifications true --sound

# Disable notifications
flow config set notifications false
```

### Log Display

Control how command output is displayed:

```shell
# Set global log mode
flow config set log-mode logfmt    # Structured logs (default)
flow config set log-mode text      # Plain text output
flow config set log-mode json      # JSON format
flow config set log-mode hidden    # Hide output
```

**Per-executable log modes:**
```yaml
executables:
  - name: debug-task
    exec:
      logMode: text  # Override global setting
      cmd: echo "Debug output"
```

### Workspace Modes

Control how flow determines your current workspace:

```shell
# Dynamic mode - auto-switch based on directory
flow config set workspace-mode dynamic

# Fixed mode - always use set workspace
flow config set workspace-mode fixed
```

> **Learn more**: See the [Workspaces guide](workspaces.md) for detailed workspace mode explanations.

### Timeouts

Set default timeout for all executables:

```shell
# Set global timeout
flow config set timeout 45m

# Examples: 30s, 5m, 2h
flow config set timeout 10m
```

## Configuration Management

### View Current Settings

```shell
# View all settings
flow config get

# View specific setting
flow config get --output json | jq '.theme'
```

### Reset to Defaults

```shell
# Reset all configuration
flow config reset

# Warning: This overwrites all customizations
```

### Configuration File Location

Your config is stored in:
- **Linux**: `~/.config/flow/config.yaml`
- **macOS**: `~/Library/Application Support/flow/config.yaml`
