---
title: Flow Cache Get
---

## flow cache get

Get cached data by key.

### Synopsis

The data store is a key-value store that can be used to persist data across executions. Values that are set outside of an executable will persist across all executions until they are cleared. When set within an executable, the data will only persist across serial or parallel sub-executables but all values will be cleared when the parent executable completes.

This will retrieve the value for the given key.

```
flow cache get KEY [flags]
```

### Options

```
  -h, --help   help for get
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow cache](flow_cache.md)	 - Manage temporary key-value data.

