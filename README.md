<p align="center"><a href="https://flowexec.io"><img src="docs/public/icon.png" alt="flow" width="100"/></a></p>

<p align="center">
    <a href="https://img.shields.io/github/v/release/flowexec/flow"><img src="https://img.shields.io/github/v/release/flowexec/flow" alt="GitHub release"></a>
    <a href="https://pkg.go.dev/github.com/flowexec/flow"><img src="https://pkg.go.dev/badge/github.com/flowexec/flow.svg" alt="Go Reference"></a>
    <a href="https://discord.gg/CtByNKNMxM"><img src="https://img.shields.io/badge/discord-join%20community-7289da?logo=discord&logoColor=white" alt="Join Discord"></a>
    <a href="https://github.com/flowexec/flow"><img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/flowexec/flow"></a>
</p>

<p align="center">
    <code>flow</code> is a workflow manager for developers who work across many projects. One interface for all your scripts, secrets, and automation — no matter the stack.
</p>

---

## Why Flow

Most projects come with their own Makefile, npm scripts, or shell scripts — each with different conventions, flags and undocumented quirks. AI-generated tools and side projects have made this worse. The real cost isn't running the scripts; it's remembering how everything works every time you switch contexts.

Flow sits above your projects, not inside them: one place to see, run, and compose everything — regardless of what's underneath.

## Quick Start

```bash
# Install
curl -sSL https://install.flowexec.io | bash

# Add the examples workspace
flow workspace add examples https://github.com/flowexec/examples --set

# Explore all workflows in the TUI
flow browse
```

You'll have a real workspace with runnable examples immediately. No configuration needed.

See the [full quickstart guide](https://flowexec.io/quickstart) to define your own workflows.

## Key Features

flow complements existing CLI tools by adding multi-project organization, built-in security, and visual discovery to your automation toolkit.

- **Works across all your projects** — Register any repo as a workspace; run its workflows from anywhere on your machine
- **Find anything instantly** — Browse, search, and filter all workflows across all projects from one TUI
- **Secrets as first-class citizens** — Encrypted local vaults with multiple backends; secrets inject at runtime, never hardcoded
- **Any execution pattern** — Serial, parallel, conditional, HTTP requests, interactive prompts, background jobs
- **Reusable templates** — Bootstrap new projects with flow-ready scaffolding from your own templates
- **AI-native** — MCP server lets Claude Code, Cursor, and other AI tools run your workflows directly

<p align="center"><img src="docs/demo/hero.gif" alt="flow" width="1600"/></p>

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

- [Installation](https://flowexec.io/installation) - Multiple installation methods
- [Quick Start](https://flowexec.io/quickstart) - Get up and running in 5 minutes
- [Core Concepts](https://flowexec.io/guides/concepts) - Understand workspaces, executables, and vaults
- [User Guides](https://flowexec.io/guides) - Comprehensive guides for all features

## Community

- [Discord Community](https://discord.gg/CtByNKNMxM) - Get help and share workflows
- [Issue Tracker](https://github.com/flowexec/flow/issues) - Report bugs and request features
- [Examples Repository](https://github.com/flowexec/examples) - Real-world workflow patterns
- [Contributing Guide](https://flowexec.io/development) - Help make flow better
