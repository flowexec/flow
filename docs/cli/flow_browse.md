## flow browse

Discover and explore available executables.

### Synopsis

Browse executables across workspaces.

  flow browse                # Interactive multi-pane executable browser
  flow browse --list         # Simple list view of executables
  flow browse VERB [ID]      # Detailed view of specific executable

See https://flowexec.io/types/flowfile?id=executableverb for more information on executable verbs and https://flowexec.io/types/flowfile?id=executableref for more information on executable references.

```
flow browse [EXECUTABLE-REFERENCE] [flags]
```

### Options

```
  -a, --all                List from all namespaces.
  -f, --filter string      Filter executable by reference substring.
  -h, --help               help for browse
  -l, --list               Show a simple list view of executables instead of interactive discovery.
  -n, --namespace string   Filter executables by namespace.
  -o, --output string      Output format. One of: yaml, json, or tui. (default "tui")
  -t, --tag stringArray    Filter by tags.
  -v, --verb string        Filter executables by verb.
  -w, --workspace string   Filter executables by workspace.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.

