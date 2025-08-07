# Flow File Types Guide

Flow users use three distinct types of configuration files that serve different purposes:

## 1. Workspace Configuration: `flow.yaml`
- **Location**: Root of each workspace directory
- **Purpose**: Configure workspace-level settings
- **Contains**: Workspace metadata, executable filters, display settings
- **Quantity**: One per workspace
- **Example path**: `/my-project/flow.yaml`

**Example content:**
```yaml
displayName: "My Project"
description: "A web application project"
tags: ["web", "typescript"]
```

**Key fields**:
- `displayName`: Human-readable workspace name
- `description`: Workspace description (markdown supported)
- `tags`: Workspace-level tags for organization

## 2. Executable Definitions: *.flow files

- Extensions: *.flow, *.flow.yaml, *.flow.yml
- Purpose: Define executable tasks and workflows
- Contains: Executable definitions with verbs, commands, parameters
- Quantity: Multiple per workspace
- Example paths: /my-project/build.flow, /my-project/deploy.flow.yaml, /my-project/.execs/test.flow.yml

**Note**: The `.flow` extension is reserved for flow files, while `flow.yaml` is the workspace configuration file. 
Do not try to create executables in a `.flow` directory, if grouping of flow files is desired use the `.execs` directory 
if a name is needed.

**Example content:**
```yaml
namespace: backend
executables:
  - verb: build
    name: api
    exec:
      cmd: npm run build
  - verb: test
    name: unit
    exec:
      cmd: npm test
```

**Key fields:**
- `namespace`: Optional logical grouping within workspace
- `executables`: Array of executable definitions
- Each executable has: `verb`, execution type (`exec`, `serial`, `parallel`, etc.)

## 3. Flow File Templates: `.flow.tmpl`
- **Extensions**: `.flow.tmpl`, `.flow.tmpl.yaml`, `.flow.tmpl.yml`
- **Purpose**: Generate new flow files and workspace scaffolding
- **Contains**: Template configuration with forms, artifacts, and flow file template
- **Quantity**: Multiple templates can be registered
- **Schema**: Template schema
- **Example paths**: `/templates/k8s-app.flow.tmpl`, `/scaffolds/web-project.flow.template`

**Example content:**
```yaml
form:
  - key: "AppName"
    prompt: "What is the application name?"
    required: true
  - key: "Namespace" 
    prompt: "What namespace should be used?"
    default: "default"
  - key: "Deploy"
    prompt: "Deploy immediately after creation?"
    type: "confirm"

artifacts:
  - srcName: "deployment.yaml"
    asTemplate: true
    dstName: "k8s-deployment.yaml"
    if: form["Deploy"]

template: |
  namespace: {{ .form.Namespace }}
  executables:
    - verb: build
      name: {{ .form.AppName }}
      exec:
        cmd: docker build -t {{ .form.AppName }} .
    - verb: deploy
      name: {{ .form.AppName }}
      exec:
        cmd: kubectl apply -f k8s-deployment.yaml
```

**Key fields:**
- `form`: Interactive form fields for user input
- `artifacts`: Files to copy/generate alongside the flow file
- `template`: Go template string that generates the actual flow file
- `preRun`/`postRun`: Optional executables to run during generation

## Key Differences

| Aspect | flow.yaml | .flow files               | .flow.tmpl files |
|--------|-----------|---------------------------|------------------|
| **Purpose** | Workspace configuration | Executable definitions    | Template generation |
| **Scope** | Entire workspace | Individual tasks          | Scaffolding/generation |
| **Location** | Workspace root only | Anywhere in workspace     | Template directories |
| **Schema** | Workspace schema | FlowFile schema           | Template schema |
| **Contains** | Settings, filters, metadata | Executables, verbs, commands | Forms, templates, artifacts |
| **Executable** | No | Yes (defines executables) | No (generates executables) |
| **Usage** | Automatic (workspace config) | `flow <verb> <id>`        | `flow template generate` |

## Common Confusion Points

### ❌ Don't Mix These Up:
- **flow.yaml is NOT executable** - it only configures the workspace
- **.flow files define executables** - they contain the actual tasks you run
- **.flow.tmpl files generate other files** - they're not executed directly

### ✅ Remember:
- **flow.yaml** = "How should this workspace behave?"
- **.flow** = "What tasks can I run?"
- **.flow.tmpl** = "How do I generate new flow files?"

## Examples Project Structure:
```
my-web-app/
├── flow.yaml                    # Workspace config
├── backend.flow                 # Backend executables
├── frontend.flow.yaml           # Frontend executables
├── deployment/
│   └── k8s.flow                # Deployment executables
└── templates/
    └── microservice.flow.tmpl   # Service template
```

## Typical Workflow:
1. **Create workspace**: Add `flow.yaml` to configure workspace
2. **Define executables**: Create `.flow` files with tasks
3. **Use templates**: (optional) Generate new `.flow` files from `.flow.tmpl` templates
4. **Sync workspace state**: Update flow's cache with `flow sync`
5. **Execute tasks**: Run executables with `flow <verb> <id>`
