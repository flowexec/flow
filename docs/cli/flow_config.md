## flow config

View and update global flow configuration.

### Synopsis

Manage global flow configuration. Settings are stored in the flow config file and apply
across all workspaces unless overridden.

Use 'config get' to view current values and 'config set <setting>' subcommands to change them.

### Options

```
  -h, --help   help for config
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow config get](flow_config_get.md)	 - Get the current global configuration values.
* [flow config reset](flow_config_reset.md)	 - Restore the default flow configuration values. This will overwrite the current configuration.
* [flow config set](flow_config_set.md)	 - Set a global configuration value.

