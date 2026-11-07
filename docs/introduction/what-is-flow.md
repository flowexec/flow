---
title: What is flow?
description: "flow is an open-source workflow engine: write tasks down as YAML flow files, run them from any project on your machine with the right secrets, and keep a record of every run, including your coding agents'."
---

# What is flow?

flow is an open-source workflow engine. You write your tasks down once, and run them from
anywhere on your machine, with the right secrets and a record of what happened.

## Why flow

Every project has commands you need to remember: how to build it, deploy it, reset its
database, rotate its keys. They live in shell history, a README, or someone's head, and the
secrets they need sit in a `.env` file that shouldn't have been committed.

flow gives those commands a home that travels with the code.

## What it does

<NumberedSection index="01" label="Write" claim="Tasks are YAML that lives in the repo.">
A flow file says what a task runs, what it needs, and how: in order, in parallel, behind a
condition, as an HTTP request, or in the background. A published JSON schema means your
editor checks it as you type.
<a href="/guides/executables">Executables →</a>
</NumberedSection>

<NumberedSection index="02" label="Run" claim="Every project, reachable from every directory.">
Register a repo as a workspace, or add one straight from a Git URL, and its tasks run by
name from any shell or from the terminal UI. Secrets come from an encrypted local vault at
the moment they are needed.
<a href="/guides/workspaces">Workspaces →</a>
</NumberedSection>

<NumberedSection index="03" label="Record" claim="Every run leaves a trace.">
Each run is kept with its status, duration, output, and where it came from: your terminal,
or a coding agent over MCP.
<a href="/guides/execution-history">History & logs →</a>
</NumberedSection>

## Built for working with agents

`flow mcp` gives Claude Code, Cursor and any other MCP client the same tasks you run by
hand, so an agent can find out what a project can do instead of guessing commands from a
README. Every run it starts is recorded with the client and session that started it.

flow provides tools and stops there: no model calls, no vendor keys, nothing
non-deterministic in the path of a task.
[AI tools & MCP →](/guides/ai-tools)

## Local-first

flow is a single binary. Your flow files, history and vaults stay on your machine, there's no
account, and flow sends no telemetry.

## Mochi, built on flow

Mochi is a desktop app built on flow by the same author. It opens your flow workspaces as
they are, finds the Makefiles, npm scripts and other tasks in your repos, and adds run
search, failure explanations and controls over what agents may run. flow stays free and
open source, with or without it.

<MochiRef placement="what-is-flow" title="Try Mochi">

Free 7-day trial, no card. Every `flow` command also works as `mochi`.

</MochiRef>

## Start here

- [Install flow](/installation)
- [Quickstart](/quickstart): a working task in about five minutes
- [Coming from Make, npm, just or Task](/guides/coming-from)
