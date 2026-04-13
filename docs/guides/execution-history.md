---
title: Execution History & Logs
---

# Execution History & Logs

flow automatically records every execution, building a searchable history with associated log output.

## Viewing History

Open the interactive log viewer:

```shell
flow logs
```

This shows a table of recent executions with the executable reference, time, duration, and status. Press `Enter` to view full details and log output for any entry.

**TUI keyboard shortcuts:**

| Key     | Context     | Action                    |
|---------|-------------|---------------------------|
| `Enter` | List view   | Open detail view          |
| `d`     | Detail view | Delete the current record |
| `x`     | List view   | Delete all records        |

### Structured Output

Export history for scripting or CI:

```shell
flow logs -o yaml
flow logs -o json
```

### Last Execution

Review the most recent execution's metadata and full log output:

```shell
flow logs --last
```

## Filtering

### By Executable Reference

Pass a ref argument to scope history to a single executable:

```shell
flow logs verb ws/ns:name
```

### By Workspace, Status, or Time

Use flags to narrow results:

```shell
flow logs -w my-workspace            # filter by workspace
flow logs --status failure            # only failed executions
flow logs --since 1h                  # last hour (supports d, h, m, s)
flow logs --limit 5                   # at most 5 records
flow logs -w api --status success --since 7d
```

Filters work with all output modes (`--last`, `-o yaml`, TUI, etc.).

## Background Execution

Run any executable in the background to free up your terminal for other work. The process is detached and tracked
automatically — you can check on it, read its output, or terminate it at any time.

### Starting a Background Run

Add the `--background` (or `-b`) flag to any `exec` command:

```shell
flow exec my-task --background
# Started background run a1b2c3d4 (PID 54321) for exec flow/:my-task
```

The command returns a short **run ID** immediately. The executable — including deeply nested serial and parallel
workflows — runs in a detached process with its output captured in the log archive.

### Listing Active Runs

See what's currently running in the background:

```shell
flow logs --running
# a1b2c3d4  PID 54321    exec flow/:my-task                  running 5m30s
```

This uses the same output format as regular execution history — it supports `-o yaml`, `-o json`, and the
interactive TUI. Stale entries (processes that exited unexpectedly) are automatically cleaned up when you list.

### Streaming Output

Attach to a background run to stream its log output in real time:

```shell
flow logs attach a1b2c3d4
```

This tail-follows the log file, printing new output as it appears. Press `Ctrl-C` to detach without
stopping the process. When the background process exits, the stream ends automatically.

### Terminating a Run

Stop a running background process by its run ID:

```shell
flow logs kill a1b2c3d4
# Terminated background run a1b2c3d4 (PID 54321).
```

> [!NOTE]
> Background runs cannot prompt for interactive input (`reviewRequired` gates, parameter prompts).
> Make sure all required parameters are provided via `--param` flags or environment variables when
> using `--background`.

## Clearing History

```shell
# Clear all history and logs
flow logs clear

# Clear history for a specific executable
flow logs clear verb ws/ns:name
```

> [!NOTE]
> Clearing history also removes the associated log archive files.

## What's Next?

- **Customize your interface** → [Interactive UI](interactive.md)
- **Integrate with external tools** → [Integrations](integrations.md)
