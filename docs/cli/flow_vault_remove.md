## flow vault remove

Remove an existing vault.

### Synopsis

Remove an existing vault by its name. The vault's encrypted secret data remains on disk at its storage path, but its configuration is deleted so flow no longer tracks it.
Note: You cannot remove the current vault.

```
flow vault remove NAME [flags]
```

### Options

```
  -h, --help            help for remove
  -o, --output string   Output format. One of: yaml, json, or tui.
  -y, --yes             Skip confirmation prompts
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow vault](flow_vault.md)	 - Manage sensitive secret stores.

