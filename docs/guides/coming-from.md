---
title: Coming From…
description: "What changes, and what doesn't, when you bring a project that uses Make, npm scripts, just, Task, or shell scripts into flow."
---

# Coming From…

flow doesn't ask you to throw out the task runner you have. Register the repo as a
workspace, import what it already defines, and write flow executables only for what was
never written down.

```shell
flow workspace add my-project . --set
```

## Make

List the Makefile under `imports` in a flow file and each target becomes an executable,
with a verb and name taken from the target. `make` still runs it. What you gain is reach:
the target runs by name from any directory, sits in one list with every other project's
tasks, and leaves a record each time it runs.

```yaml
imports:
  - Makefile
```

Comments above a target override its name, verb, description and more. See
[Importing Executables](executables.md#importing-executables).

## npm scripts

Import `package.json` the same way and its scripts become executables: `build` stays
`build`, `dev` becomes `start dev`. npm still runs them.

## Docker Compose

Importing a Compose file turns its services into executables, so bringing a stack up reads
the same as any other task.

## just and Task

flow doesn't read Justfiles or Taskfiles. Wrap the recipes you run often in an executable
that calls `just` or `task`, and they gain flow's history, secrets and cross-project
references:

```yaml
executables:
  - verb: test
    exec:
      cmd: just test
```

## Shell scripts and READMEs

Scripts (`.sh`, `.ps1`, `.bat`, `.cmd`, `.py`) import as single executables, with comment
headers for their name, arguments and secrets. Commands that only live in a README or your
shell history are worth writing down as an executable once. From then on they are
documented, searchable and recorded. [Your First Workflow](first-workflow.md) walks
through it.

## Secrets in `.env` files

Move them into a [vault](secrets.md) and reference them from the executable that needs
them. The values are decrypted at run time and never written to disk.

<MochiRef placement="coming-from" title="Want them found for you?">

Mochi, the desktop app built on flow, scans a repo's Makefiles, npm scripts, Justfiles,
Taskfiles, Compose files and scripts without anything written into it.

</MochiRef>
