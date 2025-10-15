---
title: Workspaces
---

# Workspaces

Workspaces organize your flow files and executables into logical projects. Think of them as containers for related automation.

## Workspace Management

### Adding Workspaces <!-- {docsify-ignore} -->

Register any directory as a workspace:

```shell
# Create workspace in current directory
flow workspace add my-project . --set

# Basic registration in specific directory
flow workspace add my-project /path/to/project

# Register and switch to it
flow workspace add my-project /path/to/project --set
```

When you add a workspace, flow creates a `flow.yaml` configuration file in the root directory if one doesn't exist.

### Switching Workspaces <!-- {docsify-ignore} -->

Change your current workspace:

```shell
# Switch to a workspace
flow workspace switch my-project

# Switch with fixed mode (see workspace modes below)
flow workspace switch my-project --fixed
```

### Listing and Viewing <!-- {docsify-ignore} -->

Explore your registered workspaces:

```shell
# List all workspaces
flow workspace list

# List workspaces with specific tags
flow workspace list --tag production

# View current workspace details
flow workspace get

# View specific workspace
flow workspace get my-project
```

### Removing Workspaces <!-- {docsify-ignore} -->

Unregister a workspace:

```shell
# Remove workspace registration
flow workspace remove old-project
```

> [!NOTE]
> Removing a workspace only unlinks it from flow - your files and directories remain unchanged.

## Workspace Configuration

Configure workspace behavior in the `flow.yaml` file:

```yaml
# flow.yaml
displayName: "API Service"
description: "REST API and deployment automation"
descriptionFile: README.md
tags: ["api", "production", "backend"]

# Customize verb aliases
verbAliases:
  run: ["start", "exec"]
  build: ["compile", "make"]
  # Set to {} to disable all aliases
  
# Environment variables to load for all executables in this workspace
envFiles:
  - .env
  - .env.local

# Control executable discovery
executables:
  included: ["api/", "scripts/", "deploy/"]
  excluded: ["node_modules/", ".git/", "tmp/"]
```

### Configuration Options <!-- {docsify-ignore} -->

**Display and Documentation:**
- `displayName`: Human-readable name for the workspace
- `description`: Markdown description shown in the UI
- `descriptionFile`: Path to markdown file with workspace documentation
- `tags`: Labels for filtering and categorization

**Executable Discovery:**
- `included`: Directories to search for flow files
- `excluded`: Directories to skip during discovery

**Behavior Customization:**
- `verbAliases`: Customize which verb synonyms are available
- `envFiles`: List of environment files to load for all executables (the root `.env` is loaded by default)

> **Complete reference**: See the [workspace configuration schema](../types/workspace.md) for all available options.

## Workspace Modes

Control how flow determines your current workspace:

### Dynamic Mode (Default) <!-- {docsify-ignore} -->
flow automatically switches to the workspace containing your current directory:

```shell
# Configure dynamic mode
flow config set workspace-mode dynamic

# Now flow automatically uses the right workspace
cd ~/code/api-service    # Uses api-service workspace
cd ~/code/frontend       # Uses frontend workspace
```

### Fixed Mode <!-- {docsify-ignore} -->
flow always uses the workspace you've explicitly set:

```shell
# Configure fixed mode
flow config set workspace-mode fixed

# Set the fixed workspace
flow workspace switch my-project

# Now flow always uses my-project, regardless of directory
```

## Multi-Workspace Workflows

### Cross-Workspace References <!-- {docsify-ignore} -->

Reference executables from other workspaces (requires `visibility: public`):

```yaml
executables:
  - verb: deploy
    name: full-stack
    serial:
      execs:
        - ref: build frontend/app
        - ref: build backend/api
        - ref: deploy infrastructure/k8s:services
```

### Shared Workspaces <!-- {docsify-ignore} -->

Create workspaces for shared tools and utilities:

```shell
# Create shared workspace
flow workspace add shared-tools ~/shared

# Reference from other workspaces
flow send shared-tools/slack:notification "Deployment complete"
```

## What's Next? <!-- {docsify-ignore} -->

Now that you can organize your automation with workspaces:

- **Define your tasks** → [Executables](executables.md)
- **Build sophisticated workflows** → [Advanced workflows](advanced.md)
- **Customize your interface** → [Interactive UI](interactive.md)