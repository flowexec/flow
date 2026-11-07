---
title: What is flow?
description: "flow is an open-source workflow engine. Write tasks down as YAML flow files, run them from any project on your machine, and keep a record of every run."
---

# What is flow?

flow is an open-source workflow engine. You write a project's tasks down once and run them
from anywhere on your machine.

## Why flow

Every project has commands someone has to remember: how to build it, deploy it, reset its
database, rotate its keys. They live in shell history, a README, or one person's head, and the
secrets they need sit in a `.env` file that was never meant to be committed. flow gives those
commands a home in the repo.

## What it does

<NumberedSection index="01" label="Write" claim="Tasks are YAML that lives in the repo.">
A flow file says what a task runs and what it needs: a command, steps in order or in
parallel, an HTTP request, a condition, a prompt. Every flow file validates against a
published JSON schema, so your editor can check it as you type.
<a href="/guides/executables">Executables →</a>
</NumberedSection>

<NumberedSection index="02" label="Run" claim="Every project, reachable from every directory.">
Register a repo as a workspace, or add one from a Git URL, and its tasks run by name from
any shell or from the terminal UI. Secrets come from a vault at the moment a task needs them.
<a href="/guides/workspaces">Workspaces →</a>
</NumberedSection>

<NumberedSection index="03" label="Record" claim="Every run leaves a trace.">
Each run is kept with its status, duration, output, and where it came from: your terminal
or a coding agent.
<a href="/guides/execution-history">History & logs →</a>
</NumberedSection>

## Built for working with agents

`flow mcp` gives Claude Code, Cursor and any other MCP client the tasks you run by hand, so
an agent can see what a project can do instead of guessing commands from a README.

flow provides the tools and nothing more. It makes no model calls and needs no vendor keys.
[AI tools & MCP →](/guides/ai-tools)

## Local-first

flow is a single binary. Your flow files, history and vaults stay on your machine. There's
no account, and flow sends no telemetry.

## Mochi

Mochi is a desktop app built on flow by the same author. It opens your flow workspaces as
they are and adds task discovery, run search, and controls over what agents may run. flow
stays free and open source either way.

<MochiRef placement="what-is-flow" title="Try Mochi">

Free for 7 days, no card. Every `flow` command also works as `mochi`.

</MochiRef>

## Next

- [Install flow](/installation)
- [Quickstart](/quickstart)
- [Coming from Make, npm, just or Task](/guides/coming-from)
