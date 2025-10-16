---
title: Executables
---

# Executables

Executables are the building blocks of flow automation. They can be simple commands, complex multi-step workflows, HTTP requests, or even GUI applications. 
This guide covers all executable types and configuration options.

## Finding Executables

Use the `flow browse` command to discover executables across your workspaces:

```shell
flow browse            # Interactive multi-pane browser
flow browse --list     # Simple list view
flow browse VERB ID    # View specific executable details
```

Filter executables by workspace, namespace, verb, or tag:

```shell
flow browse --workspace api --namespace v1 --verb deploy --tag production
flow browse --all --filter "database"  # Search names and descriptions
```

## Executable Configuration

### Basic Structure

Every executable needs a verb and optionally a name:

```yaml
executables:
  - verb: run
    name: my-task
    description: "Does something useful"
    tags: [development, automation]
    aliases: [task, job]
    timeout: 5m
    visibility: public
    exec:
      cmd: echo "Hello, world!"
```

### Common Fields

- **verb**: Action type (run, build, test, deploy, etc.)
- **verbAliases**: Alternative names for the verb
- **name**: Unique identifier within the namespace
- **description**: Markdown documentation for the executable
- **tags**: Labels for categorization and filtering
- **aliases**: Alternative names for the executable
- **timeout**: Maximum execution time (e.g., 30s, 5m, 1h)
- **visibility**: Access control (public, private, internal, hidden)

### Visibility Levels

- **public**: Available from any workspace
- **private**: Only available within the same workspace but shown in browse lists (default)
- **internal**: Available within workspace but hidden from browse lists
- **hidden**: Cannot be run or listed

## Environment Variables

Customize executable behavior with environment variables or temporary files using `params` or `args`.

> [!INFO]
> Executables inherit environment variables from their parent executable, workspace, and system.
> 
> By default, values defined in the `.env` file at the workspace root are automatically loaded. This can be overriden
> in the workspace configuration file with the `envFiles` field.

### Parameters (`params`)

Set environment data from various sources:

```yaml
executables:
  - verb: deploy
    name: app
    exec:
      file: deploy.sh
      params:
        # From secrets
        - secretRef: api-token
          envKey: API_TOKEN
        - secretRef: production/database-url
          envKey: DATABASE_URL
        
        # Interactive prompts
        - prompt: "Which environment?"
          envKey: ENVIRONMENT
        
        # Static values
        - text: "production"
          envKey: DEPLOY_ENV
          
        # Env File (key=value format)
        - envFile: "development.env"
        - envFile: "staging.env"
          envKey: SHARED_KEYS  # Only load specific keys

        # Saved to a file
        - secretRef: tls-cert
          outputFile: cert.pem
```

**Parameter types:**
- `secretRef`: Reference to vault secret
- `prompt`: Interactive user input
- `text`: Static value
- `envFile`: Load environment variables from a file

### Arguments (`args`)

Handle command-line arguments:

```yaml
executables:
  - verb: build
    name: container
    exec:
      file: build.sh
      args:
        # Positional argument
        - pos: 1
          envKey: IMAGE_TAG
          required: true
        
        # Flag arguments
        - flag: publish
          envKey: PUBLISH
          type: bool
          default: false
        
        - flag: registry
          envKey: REGISTRY
          default: "docker.io"

        # Saved to a file
        - flag: version
          outputFile: //version.txt
```

**Run with arguments:**
```shell
flow build container v1.2.3 publish=true registry=my-registry.com
```

**Argument types:**
- `pos`: Positional argument (by position number, starting from 1)
- `flag`: Named flag argument

### Command-Line Overrides

Override any environment variable with `--param`:

```shell
flow deploy app --param API_TOKEN=override --param ENVIRONMENT=staging
```

> [!NOTE]
> If the `outputFile` field is used to save a value, it will automatically be cleaned up after the executable finishes running.

## Working Directories

Control where executables run with the `dir` field:

```yaml
executables:
  - verb: build
    name: frontend
    exec:
      cmd: npm run build
      dir: "./frontend"  # Relative to flowfile
  
  - verb: clean
    name: downloads
    exec:
      cmd: rm -rf downloads/*
      dir: "~/Downloads"  # User home directory
  
  - verb: deploy
    name: from-root
    exec:
      cmd: kubectl apply -f k8s/
      dir: "//"  # Workspace root
  
  - verb: test
    name: isolated
    exec:
      cmd: |
        echo "Running in temporary directory"
        ls -la
      dir: "f:tmp"  # Temporary directory (auto-cleaned)
```

**Directory prefixes:**
- `//`: Workspace root directory
- `~/`: User home directory
- `./`: Current working directory
- `f:tmp`: Temporary directory (auto-cleaned)
- `$VAR`: Environment variable expansion

## Executable Types

### exec - Shell Commands

Run commands or scripts directly:

```yaml
executables:
  - verb: build
    name: app
    exec:
      cmd: npm run build && npm test
  
  - verb: deploy
    name: app
    exec:
      file: deploy.sh
      logMode: json  # text, logfmt, json, or hidden
```

**Options:**
- `cmd`: Inline command to run
- `file`: Script file to execute
- `logMode`: How to format command output

### serial - Sequential Execution

Run multiple steps in order:

```yaml
executables:
  - verb: deploy
    name: full-stack
    serial:
      failFast: true  # Stop on first failure
      execs:
        - cmd: docker build -t api .
        - cmd: docker build -t web ./frontend
        - ref: test api
        - cmd: kubectl apply -f k8s/
          retries: 3
        - cmd: kubectl rollout status deployment/api
          reviewRequired: true  # Pause for user confirmation
```

The [executable environment variables](#environment-variables) and [executable directory](#working-directories)
of the parent executable are inherited by the child executables.

**Options:**
- `failFast`: Stop execution on first failure (default: true)
- `retries`: Number of times to retry failed steps
- `reviewRequired`: Pause for user confirmation

### parallel - Concurrent Execution

Run multiple steps simultaneously:

```yaml
executables:
  - verb: test
    name: all-suites
    parallel:
      maxThreads: 4  # Limit concurrent operations
      failFast: false  # Run all tests even if some fail
      execs:
        - cmd: npm run test:unit
        - cmd: npm run test:integration
        - cmd: npm run test:e2e
        - ref: lint code
          retries: 1
```

The [executable environment variables](#environment-variables) and [executable directory](#working-directories)
of the parent executable are inherited by the child executables.

**Options:**
- `maxThreads`: Maximum concurrent operations (default: 5)
- `failFast`: Stop all operations on first failure (default: true)
- `retries`: Number of times to retry failed operations

### launch - Open Applications

Open files, URLs, or applications:

```yaml
executables:
  - verb: open
    name: workspace
    launch:
      uri: "$FLOW_WORKSPACE_PATH"
      app: "Visual Studio Code"
  
  - verb: open
    name: docs
    launch:
      uri: "https://flowexec.io"
  
  - verb: open
    name: note
    launch:
      uri: "./note.md"
      app: "Obsidian"
```

**Options:**
- `uri`: File path or URL to open (required)
- `app`: Specific application to use

### request - HTTP Requests

Make HTTP requests to APIs:

```yaml
executables:
  - verb: deploy
    name: webhook
    request:
      method: POST
      url: "https://api.example.com/deploy"
      headers:
        Authorization: "Bearer $API_TOKEN"
        Content-Type: "application/json"
      body: |
        {
          "environment": "$ENVIRONMENT",
          "version": "$VERSION"
        }
      timeout: 30s
      validStatusCodes: [200, 201]
      logResponse: true
      transformResponse: |
        "Deployment " + fromJSON(data)["status"]
      responseFile:
        filename: "deploy-response.json"
```

**Options:**
- `method`: HTTP method (GET, POST, PUT, PATCH, DELETE)
- `url`: Request URL (required)
- `headers`: Custom headers
- `body`: Request body with Expr templating
- `timeout`: Request timeout
- `validStatusCodes`: Acceptable status codes
- `logResponse`: Log response body
- `transformResponse`: Transform response with Expr templating
- `responseFile`: Save response to file

### render - Dynamic Documentation

Generate and display markdown with templates:

```yaml
executables:
  - verb: show
    name: status
    render:
      templateFile: "status-template.md"
      templateDataFile: "status-data.json"
```

**Template file example:**
```markdown
# System Status

Current time: {{ data["timestamp"] }}

## Services
{{- range .services }}
- **{{ .name }}**: {{ data["status"] }}
{{- end }}

## Metrics
- CPU: {{ data["cpu"] }}%
- Memory: {{ data["memory"] }}%
```

**Options:**
- `templateFile`: Markdown template file (required)
- `templateDataFile`: JSON/YAML data file

## Importing Executables

Generate executables from shell scripts, Makefiles, package.json scripts, or docker-compose services:

```yaml
# In flowfile
imports:
  - "scripts/deploy.sh"
  - "scripts/backup.sh"
  - "Makefile"
  - "frontend/package.json"
  - "docker-compose.yaml"
```

All imported executables are automatically tagged with `generated` and their file type (e.g., `docker-compose`, `makefile`, `package.json`).




#### **Shell Scripts (.sh files)**

Shell scripts are imported as single executables with the script's filename as the name and `exec` as the default verb.

You can use special comments to override executable metadata:

```bash
#!/bin/bash
# scripts/deploy.sh

# f:name=production f:verb=deploy
# f:description="Deploy to production environment"
# f:tag=production f:tag=critical
# f:visibility=internal
# f:timeout=10m

echo "Deploying to production..."
kubectl apply -f k8s/
```

See the [generated configuration reference](generated-config.md) for more details.

#### **Makefiles**

Makefile targets are imported as executables with a verb and name that best represents the target.

```makefile
# Makefile

# f:name=app f:verb=build f:description="Build the application"
build:
	go build -o bin/app ./cmd/app

# Run all tests
test:
	go test ./...

# f:visibility=internal
clean:
	rm -rf bin/
```

See the [generated configuration reference](generated-config.md) for more details on overriding executable configuration.

#### **Package.json Scripts**

NPM scripts from package.json are imported as executables with a verb and name that best represents the script name.

```json
{
  "scripts": {
    "build": "webpack --mode production",
    "test": "jest",
    "dev": "webpack-dev-server --mode development",
    "lint": "eslint src/"
  }
}
```

This creates executables like:
- `build` - Runs the build script
- `test` - Runs the test script
- `start dev` - Runs the development server
- `lint` - Runs the linter

#### **Docker Compose Services**

Docker Compose files are imported to create executables for managing services:

```yaml
# docker-compose.yml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "3000:3000"
  
  db:
    image: postgres:13
    environment:
      POSTGRES_DB: myapp
  
  redis:
    image: redis:6
```

This creates executables like:
- `start app` - Start the app service
- `start db` - Start the database service
- `start redis` - Start the Redis service
- `start` (alias: all, services) - Start all services
- `stop` (alias: all, services) - Stop all services
- `build app` - Build the app service (if build config exists)



## Executable References

Reference other executables to build modular workflows:

```yaml
executables:
  # Reusable components
  - verb: build
    name: api
    exec:
      cmd: docker build -t api .
  
  - verb: test
    name: api
    exec:
      cmd: npm test
  
  # Composite workflows
  - verb: deploy
    name: full
    serial:
      execs:
        - ref: build api
        - ref: test api
        - cmd: kubectl apply -f api.yaml
  
  # Cross-workspace references (requires public visibility)
  - verb: deploy
    name: with-monitoring
    serial:
      execs:
        - ref: deploy full
        - ref: trigger monitoring/slack:deployment-complete
```

**Reference formats:**
- `ref: build api` - Current workspace/namespace
- `ref: build workspace/namespace:api` - Full reference
- `ref: build workspace/api` - Specific workspace
- `ref: build namespace:api` - Specific namespace

**Cross-workspace requirements:**
- Referenced executables must have `visibility: public`
- Private, internal, and hidden executables cannot be cross-referenced

## What's Next?

Now that you understand all executable types and options:

- **Build complex workflows** → [Advanced workflows](advanced.md)
- **Secure your automation** → [Working with secrets](secrets.md)
- **Generate project templates** → [Templates & code generation](templating.md)