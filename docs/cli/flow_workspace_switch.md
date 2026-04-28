## flow workspace switch

Switch the current workspace.

```
flow workspace switch NAME [flags]
```

### Examples

```

  flow workspace switch myproject
  flow workspace switch myproject --fixed

```

### Options

```
  -f, --fixed           Set the workspace mode to fixed
  -h, --help            help for switch
  -o, --output string   Output format. One of: yaml, json, or tui.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow workspace](flow_workspace.md)	 - Manage development workspaces.

