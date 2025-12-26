## flow workspace remove

Remove an existing workspace.

### Synopsis

Remove an existing workspace. File contents will remain in the corresponding directory but the workspace will be unlinked from the flow global configurations.
Note: You cannot remove the current workspace.

```
flow workspace remove NAME [flags]
```

### Options

```
  -h, --help   help for remove
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow workspace](flow_workspace.md)	 - Manage development workspaces.

