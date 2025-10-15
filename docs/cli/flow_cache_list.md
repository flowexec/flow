## flow cache list

List all keys in the store.

### Synopsis

The data store is a key-value store that can be used to persist data across executions. Values that are set outside of an executable will persist across all executions until they are cleared. When set within an executable, the data will only persist across serial or parallel sub-executables but all values will be cleared when the parent executable completes. Use the --global flag to force use of the global cache scope, even when called from within an executable.

This will list all keys currently stored in the data store.

```
flow cache list [flags]
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format. One of: yaml, json, or tui. (default "tui")
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow cache](flow_cache.md)	 - Manage temporary key-value data.

