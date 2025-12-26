---
title: Advanced Workflows
---

# Advanced Workflows

Once you're comfortable with basic executables, flow's real power comes from building sophisticated workflows that adapt 
to conditions, maintain state, and compose multiple operations into powerful automations.

## Conditional Execution

Make your workflows smart by running different steps based on runtime conditions.

### Basic Operators

The expression language supports standard comparison and logical operators:

- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `and`, `or`, `not`
- String: `+` (concatenation), `matches` (regex matching)
- Length: `len()`

**See the [Expr Language Definition](https://expr-lang.org/docs/language-definition) for all available operators and functions.**

### Basic Conditions

Use the `if` field to control when executables run:

```yaml
executables:
  - verb: deploy
    name: app
    serial:
      execs:
        # Only run on macOS
        - if: os == "darwin"
          cmd: brew install kubectl

        # Only run on Linux
        - if: os == "linux"
          cmd: apt-get install kubectl

        # Always run the deployment
        - cmd: kubectl apply -f deployment.yaml
```

### Environment-Based Conditions

Check environment variables to customize behavior:

```yaml
executables:
  - verb: build
    name: app
    serial:
      execs:
        # Development build
        - if: env["NODE_ENV"] == "development"
          cmd: npm run build:dev
        
        # Production build  
        - if: env["NODE_ENV"] == "production"
          cmd: npm run build:prod
        
        # Run tests in CI
        - if: env["CI"] == "true"
          cmd: npm test
```

### File System Conditions
Check for files or directories to control execution:

```yaml
executables:
  - verb: deploy
    name: app
    serial:
      execs:
        # Only run if config file exists
        - if: fileExists("config.yaml")
          cmd: kubectl apply -f config.yaml
        
        # Abort deployment if no Dockerfile found
        - if: not fileExists("Dockerfile")
          cmd: echo "No Dockerfile found"; exit 1
        
        # Run deployment if Dockerfile exists
        - cmd: docker build -t myapp .
```

### Data-Driven Conditions

Use stored data to control execution flow:

```yaml
executables:
  - verb: setup
    name: feature
    serial:
      execs:
        # Enable a feature flag
        - cmd: flow cache set feature-x enabled
        
        # Later workflows can check this flag
        - if: store["feature-x"] == "enabled"
          cmd: echo "Feature X is enabled"
          
        # Complex conditions
        - if: len(store["build-id"]) > 0 and os == "linux"
          cmd: echo "Valid build on Linux"
```

### Available Context

Conditions have access to extensive runtime information:

**System Information:**
- `os` - Operating system ("linux", "darwin", "windows")
- `arch` - System architecture ("amd64", "arm64")

**Environment Variables:**
- `env["VAR_NAME"]` - Any environment variable

**Stored Data:**
- `store["key"]` - Values from the cache/data store

**Flow Context:**
- `ctx.workspace` - Current workspace name
- `ctx.namespace` - Current namespace
- `ctx.workspacePath` - Full path to workspace root
- `ctx.flowFilePath` - Path to current flow file
- `ctx.flowFileDir` - Directory containing current flow file

Additionally the following functions are provided alongside the Expr language definition:

- `fileExists(path)` - Check if file/directory exists
- `dirExists(path)` - Check if path is a directory
- `isFile(path)` - Check if path is a file
- `isDir(path)` - Check if path is a directory
- `basename(path)` - Get filename from path
- `dirname(path)` - Get directory from path
- `readFile(path)` - Read file contents as string
- `fileSize(path)` - Get file size in bytes
- `fileModTime(path)` - Get file modification time
- `fileAge(path)` - Get duration since last modified

## Managing State

Persist data across executions and share information between workflow steps.

### Cache Basics

The cache stores key-value data with different persistence scopes:

```yaml
executables:
  - verb: build
    name: with-cache
    serial:
      execs:
        # Set data in cache
        - cmd: |
            build_id=$(date +%Y%m%d_%H%M%S)
            flow cache set current-build $build_id
            echo "Build ID: $build_id"
        
        # Use cached data in later steps
        - cmd: |
            build_id=$(flow cache get current-build)
            echo "Deploying build: $build_id"
            docker tag myapp:latest myapp:$build_id
```

### Cache Persistence Scopes

Understanding cache persistence is crucial for complex workflows:

**Global Scope** - Set outside executables, persists until manually cleared:
```shell
# Set globally (persists across all executions)
flow cache set api-endpoint "https://api.prod.com"
flow cache set feature-enabled true

# Available in all executables until cleared
flow run any-executable  # Can access api-endpoint and feature-enabled
flow cache clear         # Removes all global data
```

**Execution Scope** - Set within executables, automatically cleared when parent completes:
```yaml
executables:
  - verb: test
    name: integration
    serial:
      execs:
        # This data persists across ALL sub-executables
        - cmd: flow cache set test-db-url "localhost:5432"
        - cmd: flow cache set test-session-id $(uuidgen)
        
        # All these steps can access the test data
        - ref: setup database    # Can use test-db-url
        - ref: run tests         # Can use test-session-id
        - ref: cleanup database  # Can use test-db-url
        
        # Data automatically cleared when 'test integration' completes
```

**Cache Management Commands:**
```shell
# View all cached data
flow cache list

# Get specific value
flow cache get key-name

# Set value globally
flow cache set key-name value

# Remove specific key
flow cache remove key-name

# Clear all data
flow cache clear

# Clear all data (including different scopes)
flow cache clear --all
```

See the [cache command reference](../cli/flow_cache.md) for detailed commands and options.

### Temporary Directories

Use isolated temporary directories for complex operations:

```yaml
executables:
  - verb: build
    name: container
    exec:
      dir: f:tmp  # Creates temporary directory
      cmd: |
        # All commands run in isolated temp directory
        git clone https://github.com/user/repo .
        docker build -t myapp .
        # Directory automatically cleaned up
```

When defined for a `serial` or `parallel` executable, the temporary directory is created at the start and cleaned up after all steps complete.

## Workflow Composition

Build complex automations by combining executables in sophisticated ways.

### Serial vs Parallel Execution

**Serial** - Steps run one after another:
```yaml
executables:
  - verb: deploy
    name: backend
    serial:
      failFast: true  # Stop on first failure
      execs:
        - cmd: npm run build
        - cmd: docker build -t api .
        - cmd: kubectl apply -f deployment.yaml
        - cmd: kubectl rollout status deployment/api
```

**Parallel** - Steps run simultaneously:
```yaml
executables:
  - verb: test
    name: all
    parallel:
      maxThreads: 3  # Limit concurrent operations
      failFast: false # Run all tests even if some fail
      execs:
        - cmd: npm run test:unit
        - cmd: npm run test:integration  
        - cmd: npm run test:e2e
        - cmd: npm run lint
```

### Mixed Execution Patterns

Combine serial and parallel for sophisticated workflows by defining separate executables and referencing them:

```yaml
executables:
  # Individual build steps
  - verb: build
    name: api-image
    exec:
      cmd: docker build -t api ./api
      
  - verb: build
    name: web-image
    exec:
      cmd: docker build -t web ./web
      
  - verb: build
    name: worker-image
    exec:
      cmd: docker build -t worker ./worker

  # Parallel builds
  - verb: build
    name: all-images
    parallel:
      execs:
        - ref: build api-image
        - ref: build web-image
        - ref: build worker-image

  # Full deployment workflow
  - verb: deploy
    name: microservices
    serial:
      execs:
        # Parallel preparation
        - ref: build all-images
        
        # Serial deployment (order matters)
        - cmd: kubectl apply -f database.yaml
        - cmd: kubectl wait --for=condition=ready pod -l app=database
        - cmd: kubectl apply -f api.yaml
        - cmd: kubectl apply -f web.yaml
        - cmd: kubectl apply -f worker.yaml
```

> [!NOTE]
> **Cross-workspace references**: To reference executables from other workspaces, they must have `visibility: public` in their configuration. Private, internal, and hidden executables cannot be referenced from other workspaces.

### Error Handling and Retries

Build resilient workflows that handle failures gracefully:

```yaml
executables:
  - verb: deploy
    name: resilient
    serial:
      failFast: false  # Continue on failures
      execs:
        # Retry flaky operations
        - cmd: curl -f https://api.example.com/health
          retries: 3
        
        # Continue even if optional steps fail
        - cmd: ./optional-cleanup.sh || true
        
        # Critical step with custom error handling
        - cmd: |
            if ! kubectl apply -f deployment.yaml; then
              echo "Deployment failed, rolling back..."
              kubectl rollout undo deployment/app
              exit 1
            fi
```

### Review Gates

Add human approval steps for critical operations:

```yaml
executables:
  - verb: deploy
    name: production
    serial:
      execs:
        - cmd: ./build-production.sh
        - cmd: ./run-smoke-tests.sh
        
        # Pause for human review
        - reviewRequired: true
          cmd: echo "Review deployment artifacts before continuing"
        
        - cmd: ./deploy-to-production.sh
        - cmd: ./notify-team.sh
```

## Environment Variable Handling

Understanding how environment variables are resolved and prioritized in flow executables.

### Resolution Order

Environment variables are resolved in this order (highest to lowest priority):

1. **Command-line overrides** (`--param`)
2. **Executable `params`** (secretRef, prompt, text)
3. **Executable `args`** (positional and flag arguments)
4. **Shell environment** (inherited from your terminal)

```yaml
executables:
  - verb: deploy
    name: app
    exec:
      params:
        - text: "staging"
          envKey: ENVIRONMENT
        - secretRef: api-key
          envKey: API_KEY
      args:
        - flag: verbose
          envKey: VERBOSE
          type: bool
          default: false
      cmd: ./deploy.sh
```

> [!NOTE]
> In serial and parallel executables, `params` and `args` that are defined at the parent level will apply to all 
> child executables. Argument from the parent -> child executable should use matching `EnvKey` to ensure proper resolution.

**Resolution example:**
```shell
# Shell environment
export API_KEY=shell-key
export VERBOSE=true
export ENVIRONMENT=development

# Command execution
flow deploy app verbose=false --param ENVIRONMENT=production

# Final environment variables:
# API_KEY=<secret-value>    (params wins over shell)
# VERBOSE=false             (args wins over shell and params)
# ENVIRONMENT=production    (--param override wins over all)
```

### Environment Variable Expansion

Environment variables are expanded in certain contexts:

**Directory paths:**
```yaml
executables:
  - verb: backup
    name: logs
    exec:
      dir: "$HOME/backups"  # Expands to /home/user/backups
      cmd: cp /var/log/app.log .
```

**Command strings:**
```yaml
executables:
  - verb: deploy
    name: app
    exec:
      params:
        - prompt: "Environment?"
          envKey: ENV
      cmd: |
        echo "Deploying to $ENV"
        kubectl config use-context $ENV-cluster
        kubectl apply -f k8s/$ENV/
```

### Special Environment Variables

flow provides special environment variables automatically:

- `FLOW_CURRENT_WORKSPACE` - Current workspace name
- `FLOW_CURRENT_NAMESPACE` - Current namespace
- `FLOW_WORKSPACE_PATH` - Full path to workspace
- `FLOW_EXECUTABLE_NAME` - Name of current executable
- `FLOW_DEFINITION_DIR` - Directory containing the current flow file
- `FLOW_TMP_DIR` - Temporary directory for current execution, if `f:tmp` is set

### Environment Inheritance

Child executables inherit environment variables from their parents:

```yaml
executables:
  - verb: deploy
    name: full-stack
    serial:
      params:
        - prompt: "Environment?"
          envKey: DEPLOY_ENV
        - secretRef: "${DEPLOY_ENV}/api-key"
          envKey: API_KEY
      execs:
        # These inherit DEPLOY_ENV and API_KEY
        - ref: deploy backend
        - ref: deploy frontend
        - cmd: echo "Deployed to $DEPLOY_ENV"

  - verb: deploy
    name: backend
    exec:
      # Automatically has access to DEPLOY_ENV and API_KEY
      cmd: |
        echo "Deploying backend to $DEPLOY_ENV"
        ./deploy-backend.sh --env $DEPLOY_ENV --key $API_KEY
```
