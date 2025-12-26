## flow vault remove

Remove an existing vault.

### Synopsis

Remove an existing vault by its name. The vault data will remain in it's original location, but the vault will be unlinked from the global configuration.
Note: You cannot remove the current vault.

```
flow vault remove NAME [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow vault](flow_vault.md)	 - Manage sensitive secret stores.

