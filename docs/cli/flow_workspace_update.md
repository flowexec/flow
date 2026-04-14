## flow workspace update

Pull latest changes for a git-sourced workspace.

### Synopsis

Pull the latest changes from the git remote for a workspace that was added from a Git URL. If NAME is omitted, the current workspace is used.

This respects the branch or tag that was originally specified when the workspace was added.
Use --force to discard local changes and hard reset to the remote.

```
flow workspace update [NAME] [flags]
```

### Options

```
      --force   Force update by discarding local changes (hard reset to remote)
  -h, --help    help for update
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow workspace](flow_workspace.md)	 - Manage development workspaces.

