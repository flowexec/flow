# Flow MCP Server Instructions

Flow is a local automation platform. **Executables** (tasks of any kind) are declared in `*.flow` YAML files; **workspaces** group them by project and are rooted at a `flow.yaml`. Secrets live in **vaults**. Templates (`*.flow.tmpl`) generate new workflows.

## Start here

Call `get_info` once at the start of a session (unless the context is already provided). It returns the current workspace/namespace/vault, a short summary, and URLs for docs (`llmsTxtUrl`), JSON schemas (`schemaUrls.*`), and guides (`guideUrls.*`). If the workspace has `agentInstructions` (or an `agentInstructionsFile`), they appear in `currentContext.agentInstructions` — read and follow them: they carry the workspace author's project conventions, required workflows, and environment constraints. In a worktree without its own instructions, the parent repo's registered workspace instructions apply. Fetch docs URLs when you need deeper detail instead of asking the user. `write_flowfile` validates against the flowfile schema server-side, so you don't need the schema in context to author files — fetch `schemaUrls.flowFile` only as the authoritative reference for non-trivial executables.

After `get_info`, call `get_project_status` to see the git branch, dirty state, ahead/behind, and last few run results in one shot — it replaces separate git-status and logs queries.

Re-read these instructions mid-session via the `flow://server/instructions` resource. Read `flow://failures/recent` to see what failed recently — useful for picking up where a previous session left off.

## Flow as your shell

Prefer running work **through flow** over a raw shell tool — you get the workspace's environment and secrets, captured logs, and a durable, attributable history entry (visible via `get_execution_logs` / `flow logs`) instead of an untracked shell command.

- `execute` — run a **named** executable (build, test, lint, deploy, …) by verb + optional ID. Try this first; discover names with `list_executables`.
- `run_command` — run an **arbitrary** shell command when no named executable fits (a one-off `git status`, `npm ci`, a script). Pass a short `label` so the history entry is self-documenting; `dir` sets the working directory. To run several commands in one call, pass `commands` (array) with `mode: serial` (default) or `parallel`.
- `run_python` — run **Python** code through flow instead of `python -c` or a scratch file. Pass `code` (multi-line is fine — it runs from a file, so tracebacks report real line numbers) and a short `label`. flow picks the workspace's virtualenv when there is one, so imports resolve against the project's installed dependencies; read parameters and secrets from `os.environ`.
- `run_executable` — run a **transient executable of any type** from an inline `spec` when a single command isn't enough: a `serial`/`parallel` batch, an HTTP `request`, a `render`, or a `launch`. The `spec` is one executable definition (same shape as an entry under a flowfile's `executables:`); author non-trivial ones against `schemaUrls.flowFile`.

`run_command`, `run_python`, and `run_executable` take an optional `workspace` to scope a run to another workspace's environment **without** changing the current workspace; otherwise the workspace is inferred from the run directory, then the current one.

Every run you launch is attributed to this session. `get_execution_logs` with `mine: true` returns only what *this* session has run — use it to review your own recent work; `source`/`session`/`status` filter more broadly.

By default `get_execution_logs` returns only run **metadata** (ref, status, exitCode, error). To read the captured **output**, set `tail` (last N lines — best for debugging, since errors surface at the end), `grep` (a regex to pull just the matching lines out of a large log), or `content: true` (the whole log). Output is always byte-capped to protect the context window; `contentTruncated` in the response tells you when part of a log was omitted (widen with `tail`/`grep`/`max_bytes`, or open `logFile` directly). This reads the log live, so it also works on a **still-running** execution — you get a snapshot of what has been written so far.

## Worktrees

`list_worktrees`, `add_worktree`, `remove_worktree`, and `prune_worktrees` manage git worktrees through the workspace's configured worktree directory. Use worktrees to isolate parallel tasks on separate branches without stashing or switching.

- `prune_worktrees` cleans up stale entries (directories that no longer exist on disk). Pass `merged: true` to also remove worktrees whose branches are fully merged into the default branch. Always call with `dry_run: true` first to preview what would be removed.

## Templates

`list_templates`, `get_template`, and `generate_template` discover and scaffold from flowfile templates (`*.flow.tmpl`). Templates produce ready-to-run flowfiles with optional artifacts (config files, scripts, etc.).

- Call `list_templates` to see what's available — each entry includes form fields the template expects.
- Call `get_template` before generating to see the full definition: artifacts, pre/post-run hooks, and the template body.
- `generate_template` writes the flowfile and artifacts into the workspace. Pass form values in `fields` as key→value pairs; unanswered required fields fail before writing.

## Secrets

`list_secrets` returns the names of secrets in the current vault — **never values**. Use it to discover which secrets are available for executables that reference vault entries via `$s(NAME)` or environment variables. Pass `vault` to query a specific vault instead of the current one.

## Notes

- `execute` runs a defined executable, not a raw command — use `run_command` for arbitrary commands.
- `list_executables` doesn't return an executable's `args` or `params`. Call `get_executable` to find out what it accepts.
- If the executable defines `args`, pass them in `args` as you would on a command line: `--node workbox 'uname -a'`. Shell-style quotes group words and nothing is expanded. Everything goes to the executable, never to flow itself. A leading `--` is optional; a later one makes the words after it positional. Anything the executable doesn't define fails the call before it runs, and the error lists what it accepts.
- If the executable defines `params`, pass values in `params` as an EnvKey→value map of strings. A prompt param without a value fails the call before it runs. Keys are set whether or not the executable declares them, and beat `args` with the same EnvKey. Don't pass secrets this way.
- A failed call returns `isError: true`, with `output`, `exitCode`, and `started` in the structured result. `started: false` means no command ran, so fix the call rather than checking for side effects. Otherwise `exitCode` is the failing command's own status, or 1 if flow ended the run, as on a timeout.
- Confirm before running anything destructive.
- The current workspace affects results. Filter `list_executables` (workspace/verb/keyword/tag) rather than listing everything. Prefer summarizing tool JSON over dumping it raw.
