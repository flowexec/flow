## flow logs

View execution history and logs.

### Synopsis

View execution history recorded in the data store, with associated log output. Optionally filter by executable reference.

```
flow logs [ref] [flags]
```

### Examples

```

  flow logs                          # all history
  flow logs --last                   # most recent entry with full output
  flow logs --status failed          # only failed runs
  flow logs --status running         # only in-progress runs
  flow logs --source mcp             # only runs launched by an AI/MCP client
  flow logs --session <id>           # everything one agent session ran
  flow logs run build                # history for 'run build' executable
  flow logs --running                # list active background processes

```

### Options

```
      --client string      Filter history by the client that launched the run (e.g. 'claude', 'cursor').
  -h, --help               help for logs
      --last               Print the last execution's logs
      --limit int          Maximum number of records to display.
  -o, --output string      Output format. One of: yaml, json, or tui.
      --running            Show only active background processes.
      --session string     Filter history to a single provenance session ID (e.g. an AI agent session).
      --since string       Filter history to entries after a duration (e.g. 1h, 30m, 7d).
      --source string      Filter history by run origin: 'cli' or 'mcp'.
      --status string      Filter history by status (running, completed, or failed; success/failure accepted as aliases).
  -w, --workspace string   Filter history by workspace name.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow logs attach](flow_logs_attach.md)	 - Stream log output from a running background process by run ID.
* [flow logs clear](flow_logs_clear.md)	 - Clear execution history and logs.
* [flow logs kill](flow_logs_kill.md)	 - Terminate a running background process by run ID.

