---
title: Flow Secret List
---

## flow secret list

List secrets stored in the current vault.

```
flow secret list [flags]
```

### Options

```
  -h, --help            help for list
  -o, --output string   Output format. One of: yaml, json, or tui. (default "tui")
  -p, --plaintext       Output the secret value as plain text instead of an obfuscated string
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow secret](flow_secret.md)	 - Manage secrets stored in a vault.

