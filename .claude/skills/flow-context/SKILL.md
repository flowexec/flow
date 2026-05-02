---
name: flow-context
description: This project uses flow for automation. When asked to build, test, run, deploy, lint, generate, or perform any dev task, check for a flow executable before falling back to raw shell commands.
user-invocable: false
---

This repository uses **flow** for all development automation. The `mcp__flow__*` MCP tools are available.

## Default behavior for dev tasks

When the user asks to build, test, deploy, lint, generate, or run anything:

1. Call `mcp__flow__list_executables` first — check whether a flow executable already handles the task
2. If found, use `mcp__flow__execute` with the matching verb and name — do not run a raw shell command
3. Only fall back to `Bash` if no flow executable exists for the task

## Key executables in this repo

Common refs: `test unit`, `test e2e`, `lint`, `validate`, `generate`, `build binary`

Use `mcp__flow__list_executables` to discover the full list and current names — don't assume.

## When to call get_info

Call `mcp__flow__get_info` at the start of a session or when you need schema URLs to author `.flow` files.
