---
title: Quick Start
---

# Quick Start

> [!NOTE]
> Before getting started, install the latest `flow` version using one of the methods described in the
> [installation guide](installation.md).

This guide will walk you through exploring real examples and creating your first custom workflow with `flow` in about 5 minutes.

## 1. Explore Real Examples

The fastest way to get a feel for flow is to add the examples workspace — a real collection of runnable workflows you can browse and run immediately:

```shell
flow workspace add examples https://github.com/flowexec/examples.git --set
flow browse
```

Use arrow keys to navigate, press <kbd>R</kbd> to run a selected executable. When you're ready to define your own workflows, continue below.

## 2. Create Your First Workspace

A workspace is where flow looks for your executables. Create one in any directory:

```shell
flow workspace add my-workspace . --set
```

This registers the workspace and creates a `flow.yaml` config file. The `--set` flag makes it your current workspace.

## 3. Create Your First Executable

Executables are defined in flow files (`.flow`, `.flow.yaml`, or `.flow.yml`). Let's create one:

```shell
touch hello.flow
```

Open the file and add this content:

```yaml
executables:
  - verb: run
    name: hello
    exec:
      params:
      - prompt: What is your name?
        envKey: NAME
      cmd: echo "Hello, $NAME! Welcome to flow 🎉"
```

This creates an executable that prompts for your name and greets you.

## 4. Sync and Run

Update flow's index of executables:

```shell
flow sync
```

Now run your executable:

```shell
flow run hello
```

You'll be prompted for your name, then see your personalized greeting!

## 5. Try the Interactive Browser

flow's TUI makes it easy to discover and run executables across all your workspaces:

```shell
flow browse
```

Use arrow keys to navigate, press <kbd>R</kbd> to run a selected executable.

## 6. Add More Executables

Try adding different types of executables to your `hello.flow` file:

```yaml
executables:
  - verb: run
    name: hello
    exec:
      params:
      - prompt: What is your name?
        envKey: NAME
      cmd: echo "Hello, $NAME! Welcome to flow 🎉"

  - verb: open
    name: docs
    launch:
      uri: https://flowexec.io

  - verb: test
    name: system
    exec:
      cmd: |
        echo "Testing system info..."
        echo "OS: $(uname -s)"
        echo "User: $(whoami)"
        echo "Date: $(date)"
```

Run `flow sync` then try:
- `flow open docs` - Opens the flow documentation
- `flow test system` - Shows system information

## What's Next?

Now that you've got the basics:

- **Learn the fundamentals** → [Core concepts](guides/concepts.md)
- **Secure your workflows** → [Working with secrets](guides/secrets.md)
- **Build complex automations** → [Advanced workflows](guides/advanced.md)
- **Customize your experience** → [Interactive UI](guides/interactive.md)

## Getting Help

- **Browse the docs** → Explore the guides and reference sections
- **Join the community** → [Discord server](https://discord.gg/CtByNKNMxM)
- **Report issues** → [GitHub issues](https://github.com/flowexec/flow/issues)
