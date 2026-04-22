## flow sync

Refresh workspace cache and discover new executables.

### Synopsis

Refresh the workspace cache and discover new executables. Use --git to also pull latest changes for all git-sourced workspaces before syncing. Use --force with --git to discard local changes and hard reset to the remote.

```
flow sync [flags]
```

### Options

```
      --force           Force update by discarding local changes (hard reset to remote)
  -g, --git             Pull latest changes for all git-sourced workspaces before syncing
  -h, --help            help for sync
  -o, --output string   Output format. One of: yaml, json, or tui.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.

