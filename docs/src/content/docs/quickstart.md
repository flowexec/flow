---
title: Quick Start
---

# Quick Start

> [!NOTE]
> Before getting started, install the latest `flow` version using one of the methods described in the
> [installation guide](installation.md).

This guide will walk you through creating your first workspace and executable with `flow` in about 5 minutes.

## 1. Create Your First Workspace

A workspace is where flow looks for your executables. Create one in any directory:

```shell
flow workspace add my-workspace . --set
```

This registers the workspace and creates a `flow.yaml` config file. The `--set` flag makes it your current workspace.

## 2. Create Your First Executable

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

## 3. Sync and Run

Update flow's index of executables:

```shell
flow sync
```

Now run your executable:

```shell
flow run hello
```

You'll be prompted for your name, then see your personalized greeting!

## 4. Try the Interactive Browser

flow's TUI makes it easy to discover and run executables:

```shell
flow browse
```

Use arrow keys to navigate press <kbd>R</kbd> to run an executable that you have selected.

## 5. Add More Executables

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

## 6. Explore a Real Workspace

Want to see more examples? Add the flow project itself as a workspace:

```shell
git clone https://github.com/flowexec/flow.git
flow workspace add flow flow
flow workspace switch flow
```

Then browse the executables:

```shell
flow browse
```

You'll see real-world examples of builds, tests, and development workflows used for developing flow.

## What's Next?

Now that you've got the basics:

- **Learn the fundamentals** → [Core concepts](guide/concepts.md)
- **Secure your workflows** → [Working with secrets](guide/secrets.md)
- **Build complex automations** → [Advanced workflows](guide/advanced.md)
- **Customize your experience** → [Interactive UI](guide/interactive.md)

## Getting Help

- **Browse the docs** → Explore the guides and reference sections
- **Join the community** → [Discord server](https://discord.gg/CtByNKNMxM)
- **Report issues** → [GitHub issues](https://github.com/flowexec/flow/issues)
