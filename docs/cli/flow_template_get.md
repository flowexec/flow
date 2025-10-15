## flow template get

Get a flowfile template's details. Either it's registered name or file path can be used.

```
flow template get [flags]
```

### Options

```
  -f, --file string                  Path to the template file. It must be a valid flow file template.
  -h, --help                         help for get
  -o, --output string                Output format. One of: yaml, json, or tui.
  -t, --template flow set template   Registered template name. Templates can be registered in the flow configuration file or with flow set template.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow template](flow_template.md)	 - Manage flowfile templates.

