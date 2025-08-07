# Integrations

flow integrates with popular CI/CD platforms, AI assistants, and containerized environments to bring your automation anywhere.


## AI Assistant Integration

### Model Context Protocol (MCP) <!-- {docsify-ignore} -->

Connect flow to AI assistants through the local Model Context Protocol server for natural language workflow management.
The flow MCP server enables AI assistants to discover, understand, and execute your flow workflows through conversational interfaces.

#### Basic Usage <!-- {docsify-ignore} -->

Add the MCP server command to your favorite MCP client:

```shell
flow mcp
```

The server uses stdio transport and provides AI assistants with:

**Available Tools:**
- `get_info` - Get flow information, schemas, and current context
- `execute` - Execute flow workflows
- `list_workspaces` - List all registered workspaces
- `get_workspace` - Get details about a specific workspace
- `switch_workspace` - Change the current workspace
- `list_executables` - List and filter executables across workspaces
- `get_executable` - Get detailed information about an executable
- `get_execution_logs` - Retrieve recent execution logs
- `sync_executables` - Sync workspace and executable state

**Available Prompts:**
- `generate_executable` - Generate flow executable configurations
- `generate_project_executables` - Generate complete project automation sets
- `debug_executable` - Debug failing executables
- `migrate_automation` - Convert existing automation to flow
- `explain_flow` - Explain flow concepts and usage

> [!NOTE]
> **Learn more about MCP**: Visit the [Model Context Protocol](https://modelcontextprotocol.io) documentation for client setup and integration details.

## CI/CD & Deployment <!-- {docsify-ignore} -->

### GitHub Actions

Execute flow workflows directly in your GitHub Actions pipelines with the official action.

```yaml
name: Build and Deploy
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: flowexec/action@v1
        with:
          executable: 'build app'
```

> **Complete documentation**: Visit the [Flow Execute Action](https://github.com/marketplace/actions/flow-execute) on GitHub Marketplace.

### Docker

Run flow in containerized environments for CI/CD pipelines or isolated execution.

### Basic Usage <!-- {docsify-ignore} -->

```shell
# Run with default workspace
docker run -it --rm ghcr.io/flowexec/flow

# Execute specific executable
docker run -it --rm ghcr.io/flowexec/flow validate
```

**Environment Variables**
- `REPO`: Repository URL to clone (defaults to flow's repo)
- `BRANCH`: Git branch to checkout (optional)
- `WORKSPACE`: Workspace name to use (defaults to "flow")


### Workspace from Git <!-- {docsify-ignore} -->

Automatically clone and configure a workspace:

```shell
docker run -it --rm \
  -e REPO=https://github.com/your-org/your-workspace \
  -e BRANCH=main \
  -e WORKSPACE=my-workspace \
  ghcr.io/flowexec/flow exec "deploy app"
```

### Local Workspace <!-- {docsify-ignore} -->

Mount your local workspace:

```shell
docker run -it --rm \
  -v $(pwd):/workspaces/my-workspace \
  -w /workspaces/my-workspace \
  -e WORKSPACE=my-workspace \
  ghcr.io/flowexec/flow exec "build app"
```

### In CI/CD Pipelines <!-- {docsify-ignore} -->

Any CI/CD platform that supports Docker can run flow. The key is:

1. **Use the Docker image**: `ghcr.io/flowexec/flow`
2. **Set environment variables**: `REPO`, `WORKSPACE`, `BRANCH` as needed
3. **Execute your flow commands**: `flow exec "your-executable"`

> **Note**: While this should work, the Docker integration hasn't been extensively tested. If you try flow with other CI/CD platforms, we'd love to hear about your experience!
