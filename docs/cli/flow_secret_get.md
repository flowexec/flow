## flow secret get

Get the value of a secret in the current vault.

```
flow secret get REFERENCE [flags]
```

### Options

```
      --copy            Copy the secret value to the clipboard
  -h, --help            help for get
  -o, --output string   Output format. One of: yaml, json, or tui.
  -p, --plaintext       Output the secret value as plain text instead of an obfuscated string
  -V, --vault string    Vault name to use instead of the current vault.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow secret](flow_secret.md)	 - Manage secrets stored in a vault.

