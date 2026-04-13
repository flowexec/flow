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
