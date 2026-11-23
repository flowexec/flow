---
title: Execution History & Logs
description: "View, filter, and manage flow's execution history and logs, including runs launched by AI assistants."
---

# Execution History & Logs

flow automatically records every execution, building a searchable history with associated log output.

## Viewing History

Open the interactive log viewer:

```shell
flow logs
```

This shows a table of recent executions with the executable reference, time, duration, status, and origin (which client launched the run). Press `Enter` to view full details and log output for any entry.

Each record is **lifecycle-aware**: it appears as `running` the moment a run starts and updates to `completed` or `failed` when it finishes, or `cancelled` when flow is interrupted or terminated (Ctrl-C, `SIGTERM`, `flow logs kill`). An in-progress foreground run is visible from another terminal, not just after it exits. (A run whose process dies without recording its outcome, such as on `SIGKILL`, is reconciled to `failed` the next time you view history.)

**TUI keyboard shortcuts:**

| Key     | Context     | Action                    |
|---------|-------------|---------------------------|
| `Enter` | List view   | Open detail view          |
| `d`     | Detail view | Delete the current record |
| `x`     | List view   | Delete all records        |
| `c`     | Any view    | Copy the selected ref, or the log in the detail view |

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
flow logs --status failed             # running, completed, failed, or cancelled
flow logs --status running            # only in-progress runs
flow logs --since 1h                  # last hour (supports d, h, m, s)
flow logs --limit 5                   # at most 5 records
flow logs -w api --status completed --since 7d
```

`--status` accepts the lifecycle values `running`, `completed`, `failed`, and `cancelled` (`success`/`failure` still work as aliases).

Filters work with all output modes (`--last`, `-o yaml`, TUI, etc.).

### By Origin (Provenance)

Every run records **where it came from**: `cli` for a run you started, or `mcp` for one launched by an AI assistant through the [MCP server](ai-tools.md), along with the client name and session ID for MCP runs. This turns history into an audit trail of what an assistant did on your behalf.

```shell
flow logs --source mcp               # only runs launched by an AI client
flow logs --session <id>             # everything one agent session ran
flow logs --client cursor            # runs launched by a specific client
flow logs --source mcp --status failed
```

An assistant connected over MCP can review its own runs the same way. See [AI Tools → Observability](ai-tools.md#observability).

## Background Execution

Run any executable in the background to free up your terminal for other work. The process is detached and tracked
automatically. You can check on it, read its output, or terminate it at any time.

### Starting a Background Run

Add the `--background` (or `-b`) flag to any `exec` command:

```shell
flow exec my-task --background
# Started background run a1b2c3d4 (PID 54321) for exec flow/:my-task
```

The command returns a short **run ID** immediately. The executable runs in a detached process with its output captured in the log archive.

### Listing Active Runs

See what's currently running in the background:

```shell
flow logs --running
# a1b2c3d4  PID 54321    exec flow/:my-task                  running 5m30s
```

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

## Digests

`flow digest` gathers runs, executables and files into one document, with a `## ` section per
source. Use it to read a failure next to the task that produced it, or to hand project state to a
script or agent.

```shell
flow digest run:<id>                             # metadata and the last 400 log lines
flow digest runs:"test status:failed ran:7d"     # a table of matching runs
flow digest exec:test/unit                       # definition, script source and last 5 runs
flow digest file:flow.yaml exec:build/binary     # several sources in one document
```

Each source is `type:value`:

| Source | Resolves to |
|--------|-------------|
| `run:<id>` | One run's metadata and the tail of its log. A unique ID prefix (6+ characters, as shown in tables) works |
| `runs:<filter>` | A summary table of runs. Refs whose latest run failed, and that failed more than once recently, are listed under *Repeated failures*. Terms are `ref:<text>`, `status:completed\|failed\|cancelled\|running` (`passed` works too), `ran:<age>` (or `since:`) and `limit:<n>` (default 20); bare words match the ref |
| `exec:<verb>/<id>` | The executable's definition, its script file when it has one, and its last 5 runs. `<verb> <id>` works too |
| `file:<path>` | The file's contents, resolved from the workspace root. Paths outside the workspace are rejected |

`--max-chars` (default 40000) is shared by all sections. Metadata, definitions and tables are
always kept; file and script text shrinks next, and log output gives way first. `--max-lines`
caps each run's log (default 400), and `--no-history` drops recent runs from `exec:` sections.
A section that was cut is marked `truncated`, and its heading says how much was kept, e.g.
`Output (last 120 of 2340 lines)`. Logs and files are fenced so their own Markdown can't break the
document's structure.

`-o json` returns `{"sections": [{"type", "ref", "label", "metadata", "content", "truncated"}],
"totalChars", "maxChars"}`. Run tables and repeated failures are in `metadata` (`runs`, `recentRuns`,
`repeatedFailures`) and not repeated in `content`. Cancelled and running runs don't count toward a
failure streak. The MCP server exposes the same output as the `get_digest` tool, with a smaller
default budget (12000).

## Clearing History

```shell
# Clear all history and logs
flow logs clear

# Clear history for a specific executable
flow logs clear verb ws/ns:name
```

> [!NOTE]
> Clearing history also removes the associated log archive files.

<MochiRef placement="history" title="Search your history, and ask why a run failed">

Mochi, the desktop app built on flow, reads this same history. It adds search ranked by what you run most, stats on what fails and where the time goes, and an explanation of a failed run from its log.

</MochiRef>

## What's Next?

- **Customize your interface** → [Interactive UI](interactive.md)
- **Run flow in CI** → [GitHub Actions](github-actions.md) or [Containers](containers.md)
