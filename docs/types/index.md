---
title: Configuration Reference
---

# Configuration Reference

flow reads four kinds of YAML file. Every one of them has a published JSON schema, so your editor
can validate and autocomplete it as you type — see [Editor setup](#editor-setup) below.

<CardGrid>
  <Card icon="file" title="Flow File" :stack="['*.flow', '*.flow.yaml', '*.flow.yml']" href="/types/flowfile" href-text="Reference" alt-href="/guides/executables" alt-href-text="Guide">
    A group of executables with shared metadata. This is the file you write most often — a workspace can hold as many as you like, anywhere in its tree.
  </Card>
  <Card icon="folder" title="Workspace" :stack="['flow.yaml']" href="/types/workspace" href-text="Reference" alt-href="/guides/workspaces" alt-href-text="Guide">
    Marks a directory as a workspace and sets what applies across it: display name, tags, environment files, and which paths are searched for executables.
  </Card>
  <Card icon="box" title="Template" :stack="['*.flow.tmpl']" href="/types/template" href-text="Reference" alt-href="/guides/templating" alt-href-text="Guide">
    A form plus a flow file to render from it, with files to copy and executables to run before and after generating.
  </Card>
  <Card icon="terminal" title="Config" :stack="['~/.flow/config.yaml']" href="/types/config" href-text="Reference" alt-href="/cli/flow_config" alt-href-text="CLI">
    Your machine-level settings: the current workspace and namespace, registered workspaces, theme, log mode, and timeouts.
  </Card>
</CardGrid>

## Editor setup

Point a schema-aware editor at the published schema with a `yaml-language-server` comment on the
first line, and you get completion, hover docs, and inline validation with no plugin beyond a YAML
extension:

```yaml
# yaml-language-server: $schema=https://flowexec.io/schemas/flowfile_schema.json

executables:
  - verb: run
    name: my-task
    exec:
      cmd: echo "Hello, world!"
```

| File | Schema |
|------|--------|
| `*.flow`, `*.flow.yaml`, `*.flow.yml` | [flowfile_schema.json](https://flowexec.io/schemas/flowfile_schema.json) |
| `flow.yaml` | [workspace_schema.json](https://flowexec.io/schemas/workspace_schema.json) |
| `*.flow.tmpl` | [template_schema.json](https://flowexec.io/schemas/template_schema.json) |
| `config.yaml` | [config_schema.json](https://flowexec.io/schemas/config_schema.json) |

### Associating the file extensions

`*.flow` files are YAML, but editors do not know that from the extension alone. In VS Code:

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

Other editors need the equivalent mapping of `*.flow`, `*.flow.yaml`, and `*.flow.yml` to YAML.

> [!TIP]
> `flow schema validate` checks files from the command line, so the same validation runs in CI
> without an editor. See [flow schema](/cli/flow_schema).
