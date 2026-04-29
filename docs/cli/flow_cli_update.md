## flow cli update

Update flow to the latest version.

### Synopsis

Check GitHub for a newer version of flow and install it if available.

```
flow cli update [flags]
```

### Examples

```

  flow cli update                    # check for an update and prompt before installing
  flow cli update --yes              # install the latest version without confirmation
  flow cli update --version v2.1.0   # install a specific version

```

### Options

```
  -h, --help             help for update
  -o, --output string    Output format. One of: yaml, json, or tui.
      --version string   Target version to install (e.g. v2.1.0). Defaults to the latest release.
  -y, --yes              Skip confirmation prompts
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow cli](flow_cli.md)	 - Manage the flow CLI itself.

