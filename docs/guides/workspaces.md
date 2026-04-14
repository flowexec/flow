---
title: Workspaces
---

# Workspaces

Workspaces organize your flow files and executables into logical projects. Think of them as containers for related automation.

## Workspace Management

### Adding Workspaces

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

#### Git Workspaces

You can also add workspaces directly from Git repositories. Flow clones the repository to its cache directory and registers it as a workspace:

```shell
# Clone from HTTPS URL
flow workspace add shared-tools https://github.com/myorg/tools.git

# Clone from SSH URL
flow workspace add k8s-flows git@github.com:platform/k8s.git --set

# Clone a specific branch
flow workspace add dev-tools https://github.com/myorg/tools.git --branch develop

# Clone a specific tag
flow workspace add stable https://github.com/myorg/tools.git --tag v1.0.0

# Clone from a local bare repo (useful for testing or air-gapped environments)
flow workspace add local-tools file:///path/to/bare/repo
```

Flow supports HTTPS, SSH, and `file://` Git URLs. The `file://` protocol is useful for local testing, air-gapped environments, or pointing at bare repos on a shared filesystem.

Git workspaces are stored in `~/.cache/flow/git-workspaces/` following Go module conventions (e.g., `github.com/myorg/tools/`). The git remote URL and branch/tag information are saved in the workspace's `flow.yaml` so they can be used for updates.

### Updating Git Workspaces

Pull the latest changes for a git-sourced workspace:

```shell
# Update a specific workspace
flow workspace update shared-tools

# Update the current workspace
flow workspace update

# Force update, discarding any local changes
flow workspace update shared-tools --force
```

This respects the branch or tag originally specified when the workspace was added. For branch-based workspaces, it performs a `git pull`. For tag-based workspaces, it fetches the latest tags and checks out the specified tag.

If a pull fails due to merge conflicts or local changes, the error output from git is shown directly. Use `--force` to discard local changes and hard reset to the remote state.

You can also update all git workspaces at once during a cache sync:

```shell
# Sync cache and pull all git workspaces
flow sync --git

# Force pull all git workspaces (discards local changes)
flow sync --git --force
```

### Switching Workspaces

Change your current workspace:

```shell
# Switch to a workspace
flow workspace switch my-project

# Switch with fixed mode (see workspace modes below)
flow workspace switch my-project --fixed
```

### Listing and Viewing

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

### Removing Workspaces

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

### Configuration Options

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

**Git Workspace Fields** (set automatically when adding from a Git URL):
- `gitRemote`: The git remote URL for the workspace
- `gitRef`: The branch or tag name specified at registration
- `gitRefType`: Either `branch` or `tag`

> **Complete reference**: See the [workspace configuration schema](../types/workspace.md) for all available options.

## Workspace Modes

Control how flow determines your current workspace:

### Dynamic Mode (Default)
flow automatically switches to the workspace containing your current directory:

```shell
# Configure dynamic mode
flow config set workspace-mode dynamic

# Now flow automatically uses the right workspace
cd ~/code/api-service    # Uses api-service workspace
cd ~/code/frontend       # Uses frontend workspace
```

### Fixed Mode
flow always uses the workspace you've explicitly set:

```shell
# Configure fixed mode
flow config set workspace-mode fixed

# Set the fixed workspace
flow workspace switch my-project

# Now flow always uses my-project, regardless of directory
```

## Multi-Workspace Workflows

### Cross-Workspace References

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

### Shared Workspaces

Share workspaces across teams using Git repositories:

```shell
# Add shared workspace from git
flow workspace add team-tools https://github.com/myorg/flow-workflows.git

# Keep it up to date
flow workspace update team-tools

# Reference from other workspaces
flow send team-tools/slack:notification "Deployment complete"
```

## What's Next?

Now that you can organize your automation with workspaces:

- **Define your tasks** → [Executables](executables.md)
- **Build sophisticated workflows** → [Advanced workflows](advanced.md)
- **Customize your interface** → [Interactive UI](interactive.md)