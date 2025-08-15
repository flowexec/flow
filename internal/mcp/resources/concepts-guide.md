# Flow Concepts Guide

## What is flow?

flow is a local-first, customizable CLI automation platform designed to streamline development and operations workflows. 
It helps developers organize, discover, and execute tasks across projects through a unified interface.

More comprehensive details can be fetched from the [flow documentation](https://flowexec.io).

## Core Philosophy

- **Local-First**: All data and execution happens on your machine
- **Declarative**: Define what you want to happen, not how
- **Discoverable**: Find and explore workflows through interactive interfaces
- **Composable**: Build complex workflows from simple building blocks
- **Workspace-Centric**: Organize tasks by project/workspace context

## Key Concepts

### Executables
**Executables** are the core building blocks of flow - they define actions that can be performed. 
Think of them as "smart scripts" with metadata, parameters, and rich configuration options.

**Types of Executables:**
- **exec**: Run shell commands or scripts
- **serial**: Execute multiple steps sequentially
- **parallel**: Execute multiple steps concurrently
- **request**: Make HTTP API calls
- **launch**: Open applications or URIs
- **render**: Generate and display markdown content

**Example executable:**
```yaml
executables:
  - verb: build
    name: web-app
    description: Build the web application
    exec:
      cmd: npm run build
      params:
        - envKey: NODE_ENV
          text: production
```

#### Executable References
Executables are identified by its reference: combination of **Verb** and **ID**. The  ID is in the form `<workspace>/<namespace>:<name>`.
A full reference must be unique across all registered workspaces.

For instance:
- `build app` - current workspace, root namespace
- `build backend/app` - current workspace, backend namespace
- `build my-project/backend:app` - specific workspace and namespace
- `build` - current workspace, root namespace, no name set

### Verbs
**Verbs** describe what action an executable performs. Flow groups related verbs together, allowing natural language-like commands.
By default, flow provide the following verb alias groups:

- **Execution Group**: exec, run, execute
- **Retrieval Group**: get, fetch, retrieve
- **Display Group**: show, view, list
- **Configuration Group**: configure, setup
- **Update Group**: update, upgrade

Additional aliases can be defined in the workspace or executable configuration to customize verb behavior.

Users can invoke executables using any verb using the CLI (e.g. `flow build app` and `flow deploy app`)

Run `flow exec --help` for more information on available verbs and execution details.

### Workspaces
**Workspaces** are project containers that organize related executables. Each workspace:
- Has a root directory containing the workspace configuration file (`flow.yaml`)
- Can contain multiple namespaces (defined in flow files)
- Provides isolation between projects

#### Workspace Modes
**Dynamic**: The flow workspace is determined by the current working directory that the flow CLI is executed from
**Fixed**: The flow workspace much be switched explicitly

Only executables with `public` visibility can be executed across workspaces without switching. If the current workspace
is not the workspace under review, you should switch into that workspace to avoid execution errors.

### Namespaces
**Namespaces** provide logical grouping within workspaces. They help organize executables by:
- Feature area (auth, payments, notifications)
- Environment (dev, staging, production)
- Technology stack (backend, frontend, mobile)
- Team ownership (platform, product, data)


Executable names are optional. When not specified, the verb is used as the identifier, 
e.g. `flow build` refers to the executable with the verb "build" in the current workspace.

### Executable Definitions (Flow Files)
**Flow files** (`.flow`, `.flow.yaml`, `.flow.yml`) are YAML configuration files that define executables. They support:
- Multiple executables per file
- Shared metadata (namespace, tags, descriptions)
- Environment variable management
- Conditional execution logic
- Can be located anywhere within the workspace directory

### Secrets and Vaults
**Vaults** provide secure secret storage with multiple encryption backends:
- **AES256**: Symmetric encryption with generated keys
- **Age**: Asymmetric encryption for team sharing

Executables reference secrets using `secretRef` parameters, keeping sensitive data separate from configuration.

### Templates
**Templates** enable scaffolding new workspaces and executables. They support:
- Interactive form collection
- Go / Expr template rendering
- File artifact copying
- Pre/post-run hooks
- Conditional generation

## Execution Model

### Sync

Anytime a new workspace is registered, an executable is added, or it's identifier changes, flow state will need to be synchronized.
This ensures that the latest workspace and executable definitions are available in the cache for execution.
This can be done using the `flow sync` command or with the `--sync` flag on any command.

### Management Commands

Flow provides a set of management commands to interact with workspaces, executables, and configurations:
```bash
flow config # Manage flow's user configuration (get, set, reset)
flow workspace # Manage workspaces (add, get, switch, list, remove)
flow browse # Discover executables (interactive TUI)
flow vault # Manage secrets and vaults (create, edit, get, list, remove, switch)
flow cache # Manage cache data (get, set, remove, list, clear)
flow secret # Manage secrets (get, list, remove, set)
flow template # Manage templates (add, generate, get, list)
flow logs # View execution logs
```

### Execution Command Structure
```bash
flow <verb> <reference> [arguments] [flags]
```

**Examples:**
- `flow build app` - Build the app executable
- `flow test backend/unit` - Run unit tests in backend namespace
- `flow deploy prod-project/k8s:webapp` - Deploy webapp in specific workspace/namespace

### Conditional Execution
flow supports runtime conditions using the Expr language for serial and parallel executables:
```yaml
serial:
  execs:
    - if: os == "darwin"
      cmd: brew install mytool
    - if: env["CI"] == "true"  
      cmd: run-ci-specific-setup
    - if: len(store["build-id"]) > 0 # checks the cache for a build ID key
      cmd: use-cached-build
```

### State Management
flow provides state persistence through:
- **Cache**: Key-value store for sharing data between executables. State is only persisted during execution 
  - Use `flow cache set KEY VALUE` and `flow cache get KEY` to manage cache entries within executable scripts
  - Users can also manage global keys outside the execution context using `flow cache` commands
- **Context variables**: OS, architecture, workspace, and path information
- **Environment inheritance**: Variables (`params` and `args`) flow from parent to child executables (for serial and parallel)
- **Temporary directories**: Isolated scratch space for executable runs
  - A temporary directory is created for the execution when the executable's `dir` is set to `f:tmp`)

## Workflow Patterns

### Simple Task
```yaml
executables:
  - verb: test
    name: unit
    exec:
      cmd: npm test
```

### Multi-Step Workflow
```yaml
executables:
  - verb: deploy
    name: application
    serial:
      execs:
        - ref: build app
        - ref: test unit
        - cmd: docker build -t myapp .
        - cmd: kubectl apply -f k8s/
```

### Parallel Execution
```yaml
executables:
  - verb: test
    name: all
    parallel:
      maxThreads: 3
      execs:
        - ref: test unit
        - ref: test integration  
        - ref: test e2e
```

## Integration Points

### CLI Interface
The primary interface is the `flow` CLI command:
- Interactive TUI for browsing and discovery
- Direct command execution
- Workspace and configuration management
- Secret and vault operations

### Desktop Application (Upcoming)
A companion GUI application providing:
- Visual workflow browsing
- Execution monitoring
- Configuration editing
- Documentation viewing

## Common Use Cases

Flow is NOT just a DevOps tool. It's a general-purpose automation platform for ANY repetitive task:

### Development & Operations
- Build, test, deploy applications
- Manage infrastructure and environments
- Automate CI/CD workflows
- Code generation and scaffolding

### Personal and Team Productivity
- Standardized development setup
- Shared workflow templates
- Manage todo lists via APIs (Todoist, Notion, etc.)
- Organize notes and knowledge bases
- Automate file organization and backups
- Schedule and run maintenance tasks

### Content & Media
- Process images, videos, documents
- Generate reports and documentation
- Manage blog posts and publications
- Sync content between platforms

### System Administration
- Monitor system health and resources
- Manage configurations and settings
- Automate routine maintenance
- Handle log analysis and cleanup

### Custom Integrations
- Reusable libraries of executables
- Integration with / wrapper for existing CLI tools and APIs
- Connect different APIs and services
- Build personal dashboards and tools
- Automate data synchronization
- Create custom workflows for unique needs

**Key Insight**: If someone has a repetitive task involving commands, APIs, or file operations, Flow can likely automate it.

## Best Practices

### Executable Design
- Use descriptive verbs and names
- Include clear descriptions and documentation
- Make executables idempotent when possible
- Handle errors gracefully with meaningful messages

### Workspace Organization
- Group related functionality in the same flow file (and namespace if it makes sense)
- Use consistent naming conventions
- Document workspace purpose and setup
- Share common patterns through templates

### Secret Management
- Never commit secrets to flow files
- Use descriptive secret references
- Include information on required secrets in executable documentation

### Workflow Composition
- Break complex tasks into smaller, reusable executables
- Use conditional logic for environment differences
- Leverage parallel execution for independent tasks
