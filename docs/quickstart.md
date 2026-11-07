---
title: Quickstart
description: "From a fresh install to running your own flow executable in about five minutes."
---

# Quickstart

This takes you from a fresh install to running a task you wrote, in about five minutes.
You'll need flow [installed](installation.md).

## 1. Try the examples

Add the examples repo as a workspace and open the terminal UI:

```shell
flow workspace add examples https://github.com/flowexec/examples.git --set
flow browse
```

Use the arrow keys to pick an executable and press <kbd>R</kbd> to run it.

## 2. Create a workspace

From the root of a project of your own:

```shell
flow workspace add my-project . --set
```

flow registers the folder and writes a `flow.yaml` to it. `--set` makes it your current
workspace.

## 3. Write an executable

Create `hello.flow` in the same folder:

```yaml
executables:
  - verb: run
    name: hello
    exec:
      params:
        - prompt: What is your name?
          envKey: NAME
      cmd: echo "Hello, $NAME"
```

## 4. Run it

```shell
flow sync
flow run hello
```

`flow sync` picks up new and changed flow files. After that, `hello` also shows up in
`flow browse`.

## 5. Add another

Executables come in more types than `exec`. Add these under `executables:` in the same file:

```yaml
  - verb: open
    name: docs
    launch:
      uri: https://flowexec.io

  - verb: test
    name: system
    exec:
      cmd: |
        echo "OS: $(uname -s)"
        echo "User: $(whoami)"
```

Then `flow sync`, and try `flow open docs` and `flow test system`.

## Next

- [Concepts](guides/concepts.md): workspaces, executables and vaults
- [Secrets](guides/secrets.md): keep keys out of your repo
- [Executables](guides/executables.md): every type and option
- [Coming from Make, npm, just or Task](guides/coming-from.md)
