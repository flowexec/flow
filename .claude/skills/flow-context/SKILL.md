---
name: flow-context
description: This project uses flow for automation. When asked to build, test, run, deploy, lint, generate, or perform any dev task — or to run any one-off shell command — prefer the flow MCP tools over raw Bash so the work runs with workspace env/secrets and is recorded in flow's history.
user-invocable: false
---

This repository uses **flow** for all development automation. The `mcp__flow__*` MCP tools are available — prefer them over raw `Bash` for anything runnable, so it executes with the workspace's environment and secrets and is captured in flow's execution history.

## Running work — pick the closest tool

1. **Named task?** (build, test, lint, validate, generate, deploy, …) → `mcp__flow__list_executables` to find it, then `mcp__flow__execute` with its verb + name. Don't hand-roll a shell command a flow executable already covers.
2. **Arbitrary one-off command?** (a `git ...`, `go test ./...`, a script) → `mcp__flow__run_command` with the command and a short `label`. This is preferred over raw `Bash`: it runs with workspace env/secrets and lands in `flow logs` with provenance. Pass `commands` (array) + `mode` (`serial`/`parallel`) to run several in one call.
3. **Something richer than one command?** (a serial/parallel batch, an HTTP `request`) → `mcp__flow__run_executable` with an inline `spec`.
4. Only fall back to `Bash` when a command genuinely shouldn't be recorded or flow isn't the right tool (e.g. interactive/TTY programs).

## Key executables in this repo

Common refs: `test unit`, `test e2e`, `lint`, `validate`, `generate`, `build binary`. Use `mcp__flow__list_executables` to discover current names — don't assume.

## Context & authoring

- Call `mcp__flow__get_info` at the start of a session, or when you need schema URLs to author `.flow` files.
- Author or edit flow files with `mcp__flow__write_flowfile` (validates against the schema server-side) rather than writing YAML by hand.
- Runs are scoped to the current workspace by default; `run_command`/`run_executable` accept an optional `workspace` to target another without switching the global current workspace.
