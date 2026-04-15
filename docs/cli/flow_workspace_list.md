## flow workspace list

List all registered workspaces.

```
flow workspace list [flags]
```

### Options

```
      --annotation stringArray   Filter by annotations. Format: 'key=value' for exact value match, or 'key' for presence regardless of value. Repeat the flag for multiple selectors; all selectors must match (AND).
  -h, --help                     help for list
  -o, --output string            Output format. One of: yaml, json, or tui.
  -t, --tag stringArray          Filter by tags.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow workspace](flow_workspace.md)	 - Manage development workspaces.

