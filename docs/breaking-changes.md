---
title: Breaking Changes
description: "Breaking changes between flow releases, and what to do about each one when upgrading."
---

# Breaking Changes

This page documents changes that require you to update existing flow files, config, or scripts when upgrading between releases.

## v2.5.0

### MCP Runs That Finish Are Results

**What changed:** `execute`, `run_command`, `run_python`, and `run_executable` return a run that
started and finished as a normal result, with `output` and `exitCode`, whatever the exit code.
Only a run that never started is a tool error (`NOT_STARTED`). Output has CRLF normalized, and a
line redrawn with carriage returns keeps only its last frame.

**How to migrate:** Check `exitCode` instead of `isError` to tell a failed run from a passing one.

### Unset Variables in Argument Defaults Expand to Empty

**What changed:** A `$VAR` in an argument's `default` that isn't set now expands to an empty
string, as in a shell. It used to be left as written.

**How to migrate:** Write `$$` where a default needs a literal `$`.

### Missing Env Files Are Skipped

**What changed:** A workspace `envFiles` entry or param `envFile` that doesn't exist is skipped
with a warning. A missing param `envFile` used to fail the run, and one missing workspace file
dropped every file in the list.

### Sync Pulls Git Workspaces

**What changed:** `flow sync`, the global `--sync` flag, and `sync` on the MCP tools pull
git-sourced workspaces (fast-forward only) before rescanning. A workspace that can't be pulled is
named and skipped. `--git` is deprecated, since pulling is the default, and `--force` no longer
needs it.

**How to migrate:** Use `flow sync --no-git` where a sync must not touch the network.

### Listing a Workspace Lists Every Namespace

**What changed:** `flow browse --list --workspace NAME` lists executables in all of that
workspace's namespaces unless `--namespace` narrows it. It used to list only the root namespace.

### Confirmations Need an Answer

**What changed:** Without a terminal, `workspace remove`, `vault remove`, `secret remove`, and
`config reset` read the confirmation as a line from stdin. No answer at all is a usage error
(exit 2) asking for `--yes`; it used to abort with exit 0.

**How to migrate:** Pass `--yes` in scripts, or pipe `yes`.

## v2.4.0

### Undefined Executable Input Is an Error

**What changed:** Input an executable doesn't define used to be dropped, and the executable ran
anyway on its defaults. `flow build app -- --relase` now fails before anything runs, as does a
positional past those it defines, a non-boolean flag without a value, and any of these for an
executable that defines no arguments, or any arguments given with `--cmd` or `--spec`. A step's
`args:` in a serial or parallel executable log a warning instead, so existing flows keep working.

`--dry false` on a boolean flag now sets it to `false`, where it used to set `true` and drop the
word. Other spellings such as `--dry 0` are errors, since they could be a positional; write
`--dry=0`. An empty value such as `--target=` still falls back to the default. A `--` after the first one makes the words after it positional. An omitted optional argument
with a `type` no longer fails type validation.

**How to migrate:** Remove the input, correct it, or define it under the executable's `args`.

### Exit Codes

**What changed:** A failed run used to exit 1. flow now exits with the failing command's own status.
Input flow rejects before looking at values exits 2 in every output mode, where it used to exit 2
only with `--output json` or `yaml`. Other failures inside flow, such as a missing required argument
or secret, still exit 1.

**How to migrate:** Scripts testing for a non-zero exit need no change. Update any that test for
exactly `1`.

### Argument Defaults See the Shell Environment

**What changed:** A `$VAR` in an argument's `default` now also resolves against the environment
flow was started in, so `default: "$HOME/.kube/config"` is no longer passed as literal text.

**How to migrate:** Write `$$` where a default needs a literal `$`.

### Git Workspace Fields Moved Out of `flow.yaml`

**What changed:** `flow workspace add` with a Git URL no longer writes `gitRemote`, `gitRef`, or
`gitRefType` to the clone's `flow.yaml`. They are recorded under `gitWorkspaces` in your flow
config. Existing clones migrate on the next `flow sync --git` or `flow workspace update`, and a
`flow.yaml` that differs from upstream only by those fields is restored.

**How to migrate:** Read these fields from `flow workspace get` instead of the clone's `flow.yaml`.

### MCP Tool Behavior

**What changed:**

- A failed run from `execute`, `run_command`, `run_python`, or `run_executable` now returns
  `isError: true`, with `output` and `exitCode` in the structured result.
- `get_workspace` returns an error for an unknown name instead of an empty workspace.
- A failed call has a `started` field. When it is `false` no command ran and the error code is
  `NOT_STARTED`. `exitCode` is otherwise the failing command's own status, or 1 for a timeout.
- A prompt param without a value in `params` fails the call. It used to run with an empty value.
- `execute` `args` are split shell-style and always go to the executable, so they can no longer
  set flow's own flags such as `--param` or `--background`. A verb or ID starting with `-` is rejected.
- `execute` `args` without an `executable_id` target the nameless executable, rather than reading
  the first word as the ID.
- Run output is plain text, without timestamps, log levels, or color. Commands run with `NO_COLOR=1`
  and without `FORCE_COLOR`, `CLICOLOR_FORCE`, or `TTY_FORCE`.
- `list_workspaces` and `list_executables` no longer return `fullDescription`, and cut `description`
  to its first line, up to 160 characters.

**How to migrate:** Check `isError` rather than parsing output for failures, pass prompt values in
`execute`'s `params`, and call `get_workspace` or `get_executable` for full descriptions.

## v2.3.0

### Argument Values Are Literals

**What changed:** A `$word` inside an argument value used to be expanded against flow's
environment and deleted when it did not resolve, so `flow run notify 'costs $5'` arrived as
`costs `. Argument values a user supplies are now literals, and pass through unchanged.

Expansion still applies to the strings you write in a flow file — an argument's `default`,
and a serial or parallel step's `args:` entries.

**How to migrate:** If you relied on a *caller* passing `$VAR` for flow to expand, expand it
in your shell instead (`flow run deploy "$TARGET"`), or move the reference into the
argument's `default`.

---

### Unresolved Variables Are Left As Written

**What changed:** A `$VAR` that does not resolve now stays in the string instead of becoming
an empty string. This affects every context flow expands: `dir` paths, `cmd` strings, an
argument's `default`, a step's `args:`, and a request's `url`, `headers` and `body`.

**Before:** `url: "https://$API_HOST/deploy"` with no `API_HOST` requested `https:///deploy`.

**After:** the same executable requests `https://$API_HOST/deploy`, which fails visibly.

`$$` now produces a literal `$`, which previously had no escape.

**How to migrate:** Nothing to change if your variables resolve. If you were depending on a
missing variable collapsing to nothing, give it an explicit empty default via a `param`.

---

### Request Bodies That Are JSON Are Sent As JSON

**What changed:** `body` was always evaluated as an Expr expression, and a JSON object is
valid Expr — it parses as a map literal and was stringified as
`map[environment:prod version:1.0]`. A body that is a JSON object or array is now sent as
written. Anything else is still an expression.

**How to migrate:** Nothing to change. A JSON body starts working; an expression body that
returns a string keeps working.

---

### Step `args:` Override Inherited Values

**What changed:** A serial or parallel step's `args:` list is now applied to the child, and
overrides values it would otherwise inherit from the parent. It was previously ignored when
the parent declared matching `args` — its mere presence, not its contents, was what enabled
inheritance.

**How to migrate:** Remove any placeholder `args:` entries added to force inheritance, such
as `args: ["-"]`. Inheritance no longer requires them, and a placeholder now reaches the
child as a real value.

---

## v2.0.0

### `fromFile` Import Field Removed

**What changed:** The deprecated `fromFile` field in flow files has been completely removed. The replacement field `imports` has been available since v1 and is the only supported way to import executables from other flow files.

**Before:**
```yaml
# my-workflows.flow
fromFile:
  - ./script.sh
```

**After:**
```yaml
# my-workflows.flow
imports:
  - ./script.sh
```

**How to migrate:**
Replace all occurrences of `fromFile:` with `imports:` in your `.flow` and `.flow.yaml` files. The value format is identical — no other changes are needed.

See the [Executables guide](./guides/executables) for full import documentation.

---

### Executable Argument Syntax

**What changed:** The `flag=value` positional argument format is no longer supported. Executable arguments must now use standard `--flag=value` syntax, separated from flow's own flags by `--`.

**Before:**
```sh
flow exec build flag1=value1 positional-arg
flow run deploy env=prod region=us-east-1
```

**After:**
```sh
flow exec build -- --flag1=value1 positional-arg
flow run deploy -- --env=prod --region=us-east-1
```

**Why this matters:**
- Boolean flags can now omit a value — `--verbose` is equivalent to `--verbose=true`
- Positional arguments work the same as before; only named flags changed
- Shell completion works correctly with the new syntax
- This aligns with POSIX conventions and eliminates ambiguity between flow's flags and executable flags

**How to migrate:**
1. Find all invocations of `flow exec`, `flow run`, `flow build`, etc. that pass `key=value` arguments
2. Add `--` before the first executable argument
3. Prefix each named flag with `--`

See the [Executables guide](./guides/executables) for full argument documentation.

---

### Default Workspace Removed

**What changed:** The concept of an implicit "default" workspace no longer exists. Flow requires an explicitly active workspace to be set at all times.

**Before (v1):** If no workspace was set, flow would fall back to a built-in default workspace.

**After (v2):** If no workspace is set, commands that require one will return an error. You must explicitly switch to or configure an active workspace.

**How to migrate:**

Set an active workspace with either of these commands:

```sh
# Interactively choose a workspace
flow workspace switch
```

If you're unsure which workspaces are registered:

```sh
flow workspace list
```

> [!TIP]
> Run `flow workspace add . --set` from any project directory to register and activate a local workspace in one step.

---

### Log File Path Change

**What changed:** Execution logs and archives are now stored in a new location managed by the persistent data store introduced in v2.

**Impact:**
- Log archives from v1 will not appear in `flow logs` or the execution history viewer
- Old log files remain on disk and are not deleted automatically

**How to migrate:**

Old logs are no longer surfaced by the CLI but remain on disk. The v1 log archive was stored inside the OS cache directory; v2 uses the OS state directory instead.

| Platform | v1 log path (cache) | v2 log path (state) |
|---|---|---|
| macOS | `~/Library/Caches/flow/logs` | `~/Library/Logs/flow/logs` |
| Linux | `~/.cache/flow/logs` | `~/.local/state/flow/logs` |
| Windows | `%LOCALAPPDATA%\flow\logs` | `%LOCALAPPDATA%\flow\logs` (unchanged) |

You can safely delete the old directory once you've confirmed you no longer need those archives:

::: code-group
```sh [macOS]
rm -rf ~/Library/Caches/flow/logs
```

```sh [Linux]
rm -rf ~/.cache/flow/logs
```

```powershell [Windows]
Remove-Item -Recurse -Force "$env:LOCALAPPDATA\flow\logs"
```
:::

New log storage is managed automatically — no manual setup is required.
