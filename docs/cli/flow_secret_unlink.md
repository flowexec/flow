## flow secret unlink

Remove a link from the current vault, leaving the secret itself untouched.

```
flow secret unlink NAME [flags]
```

### Options

```
  -h, --help            help for unlink
  -o, --output string   Output format. One of: yaml, json, or tui.
  -V, --vault string    Vault name to use instead of the current vault.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow secret](flow_secret.md)	 - Manage secrets stored in a vault.

