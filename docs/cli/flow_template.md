## flow template

Manage flowfile templates.

### Synopsis

Manage flowfile templates. A template is a reusable flowfile scaffold that can generate
executables, directory structures, and configuration files via 'template generate'.

Templates are registered by name for easy reuse. Use 'template add' to register a
template, 'template generate' to scaffold from one, and 'template list' to see what's available.

Templates (*.flow.tmpl files) are also auto-discovered within your workspaces during
'flow sync' — dropping one into a workspace makes it available without registering it.
Registered templates take precedence over discovered ones when names collide; use a
'workspace/name' reference to target a specific discovered template.

### Options

```
  -h, --help   help for template
```

### Options inherited from parent commands

```
  -L, --log-level string   Log verbosity level (debug, info, fatal) (default "info")
      --sync               Sync flow cache and workspaces
```

### SEE ALSO

* [flow](flow.md)	 - flow is a command line interface designed to make managing and running development workflows easier.
* [flow template add](flow_template_add.md)	 - Register a flowfile template by name.
* [flow template generate](flow_template_generate.md)	 - Generate workspace executables and scaffolding from a flowfile template.
* [flow template get](flow_template_get.md)	 - Get a flowfile template's details. Either it's registered name or file path can be used.
* [flow template list](flow_template_list.md)	 - List registered and workspace-discovered flowfile templates.
* [flow template remove](flow_template_remove.md)	 - Unregister a flowfile template by name.

