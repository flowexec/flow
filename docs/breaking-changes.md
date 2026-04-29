---
title: Breaking Changes
---

# Breaking Changes

This page documents changes that require you to update existing flow files, config, or scripts when upgrading between major versions.

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
