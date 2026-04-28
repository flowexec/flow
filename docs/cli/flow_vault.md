## flow vault

Manage sensitive secret stores.

### Synopsis

Manage secret stores (vaults). A vault is an encrypted key-value store that holds secrets
referenced by your executables. Multiple vault types are supported (e.g. age encryption,
AES-256, system keyring, or environment-variable passthrough).

One vault is active at a time; use 'vault switch' to change the active vault. Secrets
within a vault are managed with the 'secret' subcommands.

### Options

```
  -h, --help   help for vault
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow vault create](flow_vault_create.md)	 - Create a new vault.
* [flow vault edit](flow_vault_edit.md)	 - Edit the configuration of an existing vault.
* [flow vault get](flow_vault_get.md)	 - Get the details of a vault.
* [flow vault list](flow_vault_list.md)	 - List all available vaults.
* [flow vault remove](flow_vault_remove.md)	 - Remove an existing vault.
* [flow vault switch](flow_vault_switch.md)	 - Switch the active vault.

