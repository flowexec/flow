## flow logs

View execution history and logs.

### Synopsis

View execution history recorded in the data store, with associated log output. Optionally filter by executable reference.

```
flow logs [ref] [flags]
```

### Options

```
  -h, --help            help for logs
      --last            Print the last execution's logs
  -o, --output string   Output format. One of: yaml, json, or tui.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow logs clear](flow_logs_clear.md)	 - Clear execution history and logs.

