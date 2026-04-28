## flow secret

Manage secrets stored in a vault.

### Synopsis

Manage secrets stored in the active vault. Secrets are encrypted key-value pairs that
can be referenced inside flowfiles using the secret reference syntax (e.g. ${secret:MY_KEY}).

The active vault is used by default; pass --vault to target a different one. Use
'vault' subcommands to create and manage vaults.

### Options

```
  -h, --help   help for secret
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow secret get](flow_secret_get.md)	 - Get the value of a secret in the current vault.
* [flow secret list](flow_secret_list.md)	 - List secrets stored in the current vault.
* [flow secret remove](flow_secret_remove.md)	 - Remove a secret from the vault.
* [flow secret set](flow_secret_set.md)	 - Set a secret in the current vault. If no value is provided, you will be prompted to enter one.

