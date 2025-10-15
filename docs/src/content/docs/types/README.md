---
title: flow YAML Configurations
---

# flow YAML Configurations

- [FlowFile](flowfile.md)
- [Template](template.md)
- [Workspace](workspace.md)
- [Config](config.md)

## IDE Integration <!-- {docsify-ignore} -->

All flow configuration files have YAML schemas available for intelligent suggestions and validation in your IDE.

### Enable Schema Validation <!-- {docsify-ignore} -->

Add this comment to the top of your flow files:

```yaml
# yaml-language-server: $schema=https://flowexec.io/schemas/flowfile_schema.json

executables:
  - verb: run
    name: my-task
    exec:
      cmd: echo "Hello, world!"
```

### Available Schemas <!-- {docsify-ignore} -->

- **FlowFile**: `https://flowexec.io/schemas/flowfile_schema.json`
- **Template**: `https://flowexec.io/schemas/template_schema.json`
- **Workspace**: `https://flowexec.io/schemas/workspace_schema.json`
- **Config**: `https://flowexec.io/schemas/config_schema.json`

### IDE Setup <!-- {docsify-ignore} -->

**VS Code**: Install the YAML extension and configure file associations:

```json
// settings.json
{
  "files.associations": {
    "*.flow": "yaml",
    "*.flow.yaml": "yaml",
    "*.flow.yml": "yaml"
  }
}
```

**Other IDEs**: Configure your IDE to treat `*.flow`, `*.flow.yaml`, and `*.flow.yml` files as YAML files.
