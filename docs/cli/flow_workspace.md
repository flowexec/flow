## flow workspace

Manage development workspaces.

### Synopsis

Manage flow workspaces. A workspace is a directory (local or git-sourced) that contains
flow files defining your executables. One workspace is active at a time.

Workspaces are registered globally so flow can find executables across projects.
Use 'workspace add' to register a new workspace, 'workspace switch' to change the
active workspace, and 'workspace update' to pull the latest changes from a remote source.

### Options

```
  -h, --help   help for workspace
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow workspace add](flow_workspace_add.md)	 - Initialize a new workspace from a local path or Git URL.
* [flow workspace get](flow_workspace_get.md)	 - Get workspace details. If the name is omitted, the current workspace is used.
* [flow workspace list](flow_workspace_list.md)	 - List all registered workspaces.
* [flow workspace remove](flow_workspace_remove.md)	 - Remove an existing workspace.
* [flow workspace switch](flow_workspace_switch.md)	 - Switch the current workspace.
* [flow workspace update](flow_workspace_update.md)	 - Pull latest changes for a git-sourced workspace.

