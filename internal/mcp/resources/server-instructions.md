# Flow MCP Server Instructions

Flow is a local automation platform. **Executables** (tasks of any kind) are declared in `*.flow` YAML files; **workspaces** group them by project and are rooted at a `flow.yaml`. Secrets live in **vaults**. Templates (`*.flow.tmpl`) generate new workflows.

## Start here

Call `get_info` once at the start of a session (unless the context is already provided). It returns the current workspace/namespace/vault, a short summary, and URLs for docs (`llmsTxtUrl`), JSON schemas (`schemaUrls.*`), and guides (`guideUrls.*`). Fetch those URLs when you need deeper detail instead of asking the user. `write_flowfile` validates against the flowfile schema server-side, so you don't need the schema in context to author files — fetch `schemaUrls.flowFile` only as the authoritative reference for non-trivial executables.

## Flow as your shell

Prefer running work **through flow** over a raw shell tool — you get the workspace's environment and secrets, captured logs, and a durable, attributable history entry (visible via `get_execution_logs` / `flow logs`) instead of an untracked shell command.

- `execute` — run a **named** executable (build, test, lint, deploy, …) by verb + optional ID. Try this first; discover names with `list_executables`.
- `run_command` — run an **arbitrary** shell command when no named executable fits (a one-off `git status`, `npm ci`, a script). Pass a short `label` so the history entry is self-documenting; `dir` sets the working directory. To run several commands in one call, pass `commands` (array) with `mode: serial` (default) or `parallel`.
- `run_executable` — run a **transient executable of any type** from an inline `spec` when a single command isn't enough: a `serial`/`parallel` batch, an HTTP `request`, a `render`, or a `launch`. The `spec` is one executable definition (same shape as an entry under a flowfile's `executables:`); author non-trivial ones against `schemaUrls.flowFile`.

`run_command` and `run_executable` take an optional `workspace` to scope a run to another workspace's environment **without** changing the current workspace; otherwise the workspace is inferred from the run directory, then the current one.

## Notes

- `execute` runs a defined executable, not a raw command — use `run_command` for arbitrary commands. If the executable defines `args`, pass them in `args`; if it defines prompt `params`, pass them in `params` as an EnvKey→value map.
- Confirm before running anything destructive.
- The current workspace affects results. Filter `list_executables` (workspace/verb/keyword/tag) rather than listing everything. Prefer summarizing tool JSON over dumping it raw.
