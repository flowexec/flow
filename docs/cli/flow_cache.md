## flow cache

Manage temporary key-value data.

### Synopsis

Manage temporary key-value data. Values set outside executables runs persist globally, while values set within executables persist only for that execution scope.

### Options

```
  -h, --help   help for cache
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow cache clear](flow_cache_clear.md)	 - Clear cache data. Use --all to remove data across all scopes.
* [flow cache get](flow_cache_get.md)	 - Get cached data by key.
* [flow cache list](flow_cache_list.md)	 - List all keys in the store.
* [flow cache remove](flow_cache_remove.md)	 - Remove a key from the cached data store.
* [flow cache set](flow_cache_set.md)	 - Set cached data by key.

