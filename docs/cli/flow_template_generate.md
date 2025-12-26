## flow template generate

Generate workspace executables and scaffolding from a flowfile template.

### Synopsis

Add rendered executables from a flowfile template to a workspace.

The WORKSPACE_NAME is the name of the workspace to initialize the flowfile template in.
The FLOWFILE_NAME is the name to give the flowfile (if applicable) when rendering its template.

One one of -f or -t must be provided and must point to a valid flowfile template.
The -o flag can be used to specify an output path within the workspace to create the flowfile and its artifacts in.

```
flow template generate FLOWFILE_NAME [-w WORKSPACE ] [-o OUTPUT_DIR] [-f FILE | -t TEMPLATE] [flags]
```

### Options

```
  -f, --file string                  Path to the template file. It must be a valid flow file template.
  -h, --help                         help for generate
  -o, --output string                Output directory (within the workspace) to create the flow file and its artifacts. If the directory does not exist, it will be created.
  -t, --template flow set template   Registered template name. Templates can be registered in the flow configuration file or with flow set template.
  -w, --workspace string             Workspace to create the flow file and its artifacts. Defaults to the current workspace.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow template](flow_template.md)	 - Manage flowfile templates.

