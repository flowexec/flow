<p align="center"><a href="https://flowexec.io"><img src="docs/_media/logo.png" alt="flow" width="200"/></a></p>

<p align="center">
    <a href="https://img.shields.io/github/v/release/flowexec/flow"><img src="https://img.shields.io/github/v/release/flowexec/flow" alt="GitHub release"></a>
    <a href="https://pkg.go.dev/github.com/flowexec/flow"><img src="https://pkg.go.dev/badge/github.com/flowexec/flow.svg" alt="Go Reference"></a>
    <a href="https://discord.gg/CtByNKNMxM"><img src="https://img.shields.io/badge/discord-join%20community-7289da?logo=discord&logoColor=white" alt="Join Discord"></a>
    <a href="https://github.com/flowexec/flow"><img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/flowexec/flow"></a>
</p>

<p align="center">
    <b>flow</b> is your personal workflow hub - organize automation across all your projects with built-in secrets, templates, and cross-project composition.
    Define workflows in YAML, discover them visually, and run them anywhere.
</p>

---

## Quick Start

```bash
# Install
curl -sSL https://raw.githubusercontent.com/flowexec/flow/main/scripts/install.sh | bash

# Create your first workflow
flow workspace add my-project . --set
echo 'executables:
  - verb: run
    name: hello
    exec:
      cmd: echo "Hello from flow!"' > hello.flow

# Run it
flow sync
flow run hello
```

## Key Features

flow complements existing CLI tools by adding multi-project organization, built-in security, and visual discovery to your automation toolkit.

- **Workspace organization** - Group and manage workflows across multiple projects
- **Encrypted secret vaults** - Multiple backends (AES, Age, external tools)
- **Interactive discovery** - Browse, search, and filter workflows visually
- **Flexible execution** - Serial, parallel, conditional, and interactive workflows
- **Workflow generation** - Create projects and workflows from reusable templates
- **Composable workflows** - Reference and chain workflows within and across projects
- **Platform integrations** - GitHub Actions, AI assistants (MCP), and more

<p align="center"><img src="docs/_media/demo.gif" alt="flow" width="1600"/></p>

## Example Workflows

```yaml
# api.flow
executables:
  - verb: deploy
    name: staging
    serial:
      execs:
        - cmd: npm run build
        - cmd: docker build -t api:staging .
        - ref: shared-tools/k8s:deploy-staging
        - cmd: curl -f https://api-staging.example.com/health

  - verb: backup
    name: database
    exec:
      params:
        - secretRef: database-url
          envKey: DATABASE_URL
      cmd: pg_dump $DATABASE_URL > backup-$(date +%Y%m%d).sql
```

```bash
# Run workflows
flow deploy staging
flow backup database

# Visual discovery
flow browse
```

## Documentation

**Complete documentation at [flowexec.io](https://flowexec.io)**

- [Installation](https://flowexec.io/#/installation) - Multiple installation methods
- [Quick Start](https://flowexec.io/#/quickstart) - Get up and running in 5 minutes
- [Core Concepts](https://flowexec.io/#/guide/concepts) - Understand workspaces, executables, and vaults
- [User Guides](https://flowexec.io/#/guide/README) - Comprehensive guides for all features

## Community

- [Discord Community](https://discord.gg/CtByNKNMxM) - Get help and share workflows
- [Issue Tracker](https://github.com/flowexec/flow/issues) - Report bugs and request features
- [Examples Repository](https://github.com/flowexec/examples) - Real-world workflow patterns
- [Contributing Guide](https://flowexec.io/#/development) - Help make flow better
