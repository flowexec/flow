## flow schema validate

Validate flow files and workspace configs against their schemas.

### Synopsis

Validate one or more flow files or workspace configuration files against their JSON schemas. File type is auto-detected from the filename (*.flow for flow files, flow.yaml for workspace configs). Use --type to override auto-detection. Use --strict to also check for unknown keys.

```
flow schema validate FILE... [flags]
```

### Options

```
  -h, --help            help for validate
  -o, --output string   Output format. One of: yaml, json, or tui.
      --strict          Also check for unknown keys not defined in the schema
      --type string     File type to validate as (flowfile, workspace, config, template). Auto-detected if omitted.
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow schema](flow_schema.md)	 - Schema utilities for flow files.

