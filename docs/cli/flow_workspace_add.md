## flow workspace add

Initialize a new workspace from a local path or Git URL.

### Synopsis

Initialize a new workspace. PATH_OR_GIT_URL can be a local directory path or a Git repository URL (HTTPS or SSH). When a Git URL is provided, the repository is cloned to the flow cache directory and registered as a workspace.

Examples:
  flow workspace add my-ws ./path/to/dir
  flow workspace add shared https://github.com/org/flows.git
  flow workspace add tools git@github.com:org/tools.git --branch main
  flow workspace add stable https://github.com/org/flows.git --tag v1.0.0

```
flow workspace add NAME PATH_OR_GIT_URL [flags]
```

### Options

```
  -b, --branch string   Git branch to checkout when cloning a git workspace
      --depth int       Git clone depth (0 for full history)
  -h, --help            help for add
  -o, --output string   Output format. One of: yaml, json, or tui.
  -s, --set             Set the newly created workspace as the current workspace
      --tag string      Git tag to checkout when cloning a git workspace
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow workspace](flow_workspace.md)	 - Manage development workspaces.

