# Core Concepts

flow is built around three key concepts that work together to organize and run your automation workflows.

## Workspaces

A **workspace** is a directory containing your flow files and executables. Think of it as a project folder that flow knows about.

```shell
# Register any directory as a workspace
flow workspace add my-project ~/code/my-project --set

# Switch between workspaces
flow workspace switch my-project
```

Each workspace has a `flow.yaml` config file that defines workspace-specific settings like which directories to search for executables.

**Key points:**
- Workspaces can be anywhere on your system
- You can have multiple workspaces for different projects
- flow automatically discovers executables within registered workspaces
- Workspaces can be configured to customize behavior and discovery

> **Learn more**: See the [Workspaces guide](workspaces.md) for complete workspace management, configuration, and organization patterns.

## Executables

An **executable** is a task or workflow defined in a flow file (`.flow`, `.flow.yaml`, or `.flow.yml`). Executables are the building blocks of your automation.

```yaml
# hello.flow
executables:
  - verb: run
    name: hello
    exec:
      cmd: echo "Hello, world!"
  
  - verb: deploy
    name: app
    serial:
      execs:
        - cmd: npm run build
        - cmd: docker build -t app .
        - cmd: kubectl apply -f deployment.yaml
```

**Key points:**
- Executables are defined in YAML files within your workspaces
- They can be simple commands or complex multi-step workflows
- Each executable has a verb (like `run`, `build`, `deploy`) and optional name
- You can compose executables by referencing other executables

> **Learn more**: See the [Executables guide](executables.md) for complete configuration details and all executable types.

## Secrets Vault

The **vault** securely stores sensitive information like API keys, passwords, and tokens that your executables need.

```shell
# Create a vault
flow vault create my-vault

# Add secrets
flow secret set api-key
flow secret set database-url
```

```yaml
# Use in executables
executables:
  - verb: deploy
    name: app
    exec:
      params:
        - secretRef: api-key
          envKey: API_KEY
        - secretRef: database-url
          envKey: DATABASE_URL
      cmd: ./deploy.sh
```

**Key points:**
- Secrets are encrypted and stored locally
- Multiple vault types supported (AES, Age, external tools _coming soon_)
- Secrets are passed to executables as environment variables
- Vaults can be switched for different environments or projects

> **Learn more**: See the [Working with secrets guide](secrets.md) for complete vault setup and secret management.

## How They Work Together

Here's how these concepts connect:

1. **Workspaces** contain your flow files and organize your projects
2. **Executables** defined in those files automate your tasks
3. **Secrets** from the vault provide secure configuration for those executables

```
Workspace (my-project/)
├── flow.yaml              # Workspace config
├── api.flow               # API-related executables
├── deploy.flow            # Deployment executables
└── scripts/
    └── deploy.sh

Vault (encrypted)
├── api-key
├── database-url
└── ssl-cert
```

## Running Executables

### Basic Execution

The main command for running executables is `flow exec`, but you can use any verb:

```shell
# These are equivalent
flow exec my-task
flow run my-task

# Use the configured verb
flow build my-app
flow test integration
```

### Executable References

Executables are identified by their unique verb and ID using the format `workspace/namespace:name`:

```shell
# Full reference
flow send my-workspace/api:request

# Current workspace assumed
flow send api:request

# Current workspace and namespace assumed  
flow send request

# Nameless executable (verb only)
flow build
```

**Verbs**

The "verb" is a single word that describes the operation being executed. It can be configured in the flowfile for
each executable.

When running in the CLI, the configured verb can be replaced with any synonym/alias that describes the operation.

For instance, `flow test my-app` is equivalent to `flow validate my-app`. This allows for a more natural language-like
interaction with the CLI, making it easier to remember and use.
*See the [verb reference](../types/flowfile.md#verb-groups) for a list the default verbs and their synonym mappings.*

> [!TIP]
> Create shell aliases for commonly used verbs to make running executables easier. For example:
> ```shell
> alias build="flow build"
> ```
> This allows you to run `build my-app` instead of `flow build my-app` or the alias `flow package my-app`.


### Command-Line Overrides

Override or set additional environment variables using the `--param` flag:

```shell
flow deploy app --param API_KEY=override-value --param DRY_RUN=true --param VERBOSE=true
```

### Custom Verb Aliases

Workspaces can customize which verb aliases are available. This allows you to:

- **Use custom aliases**: Define your own preferred aliases for verbs
- **Disable default aliases**: Set an empty map `{}` to disable all verb aliases
- **Selective aliases**: Only enable specific aliases for certain verbs

```yaml
# In workspace flow.yaml
verbAliases:
  run: ["exec", "start"]    # `run` executables can use `exec` or `start`
  build: ["compile"]        # `build` executables can use `compile`
  # No entry for `test` means no aliases for test executables

# To disable all verb aliases:
verbAliases: {}
```

You can also define verb aliases on individual executables:

```yaml
# In flow file
executables:
  - verb: deploy
    name: app
    verbAliases: ["release", "apply"]
    exec:
      cmd: ./deploy.sh
```

### Discovery and Sync

When you create, move, or delete flow files, update the executable index:

```shell
# Sync executables
flow sync

# Or sync automatically before running
flow exec my-task --sync
```

## Organization Features

flow provides several ways to organize and find your executables:

**Workspaces** - Organize projects and configure discovery:
```yaml
# In workspace flow.yaml  
displayName: "My API Project"
description: "REST API and deployment tools"
tags: ["api", "production"]
```

**Namespaces** - Group related executables within a flow file:
```yaml
namespace: api
executables:
  - name: start
  - name: stop
  - name: restart
```

**Tags** - Label executables for easy filtering:
```yaml
executables:
  - name: deploy
    tags: [production, critical]
```

**Verbs** - Describe what an executable does (run, build, test, deploy, etc.)

**Visibility** - Control who can see and run executables (public, private, internal, hidden)

> **Learn more**: See the [Executables guide](executables.md) for complete configuration options and the [Interactive UI guide](interactive.md) for filtering and search features.

## Discovery and Execution

Once you understand these concepts, using flow becomes straightforward:

```shell
# Discover executables
flow browse                    # Interactive browser
flow browse --list             # Interactive simple list
flow browse --workspace api    # Filter by workspace
flow browse --tag production   # Filter by tags

# Run executables
flow run hello                 # By name
flow build my-project/app      # By full reference
```

## What's Next? <!-- {docsify-ignore} -->

Now that you understand the core concepts:

- **Build something** → [Your first workflow](first-workflow.md)
- **Secure your automation** → [Working with secrets](secrets.md)
- **Explore advanced features** → [Advanced workflows](advanced.md)
- **Customize your experience** → [Interactive UI](interactive.md)
