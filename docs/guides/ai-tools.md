---
title: AI Tools
---

# AI Tools

flow exposes your automation to AI coding tools through the [Model Context Protocol](https://modelcontextprotocol.io). Your assistant can discover, run, and write executables the same way you do — with full access to workspace context, secrets handling, and execution history.

## MCP Server

### Setup

Add this to your MCP client configuration (Claude Code, Cursor, Cline, or any MCP-compatible client):

```json
{
  "mcpServers": {
    "flow": {
      "command": "flow",
      "args": ["mcp"]
    }
  }
}
```

The server runs over stdio. That's the entire setup.

### What's available

**Tools**

| Tool | Description |
|------|-------------|
| `get_info` | Current workspace, schema URLs, and documentation index |
| `list_workspaces` | All registered workspaces |
| `get_workspace` | Details and config for a specific workspace |
| `switch_workspace` | Change the active workspace |
| `list_executables` | Browse executables — filterable by tag, verb, workspace |
| `get_executable` | Full definition and metadata for a specific executable |
| `execute` | Run an executable by ref |
| `get_execution_logs` | Output from recent runs |
| `sync_executables` | Refresh cached workspace and executable state |
| `write_flowfile` | Create or update a `.flow` file, validated before writing |

**Prompts**

Structured prompts the assistant can invoke for common tasks:

| Prompt | Description |
|--------|-------------|
| `generate_executable` | Generate a new executable for a described task |
| `generate_project_executables` | Generate a full automation set for a project |
| `debug_executable` | Debug a failing executable |
| `migrate_automation` | Convert existing Makefile, npm scripts, or shell scripts to flow |
| `explain_flow` | Explain flow concepts and configuration |

## Wiring It Up for Your Project

Connecting flow via MCP gives your assistant the tools — but it won't automatically reach for them over plain shell commands. If you ask it to run tests, it might just run `go test ./...` instead of your `test` executable.

The solution is a small instruction file committed to your repo. Claude Code calls these [background skills](https://agentskills.io): they load into context on every session without any manual invocation.

Create `.claude/skills/flow-context/SKILL.md`:

```markdown
---
name: flow-context
description: This project uses flow for automation. When asked to build, test, run, deploy, lint, or perform any dev task, check for a flow executable before falling back to raw shell commands.
user-invocable: false
---

This repository uses flow for automation. The `mcp__flow__*` MCP tools are available.

When the user asks to build, test, deploy, lint, generate, or run anything:

1. Call `mcp__flow__list_executables` first — check whether a flow executable already handles the task
2. If found, use `mcp__flow__execute` with the matching verb and name
3. Only fall back to shell if no flow executable exists for the task

Call `mcp__flow__get_info` at the start of a session or when you need schema URLs to author .flow files.
```

`user-invocable: false` keeps it out of the `/` command menu — it's not something you invoke, it's just always there. Update the body with the specific executables your project uses.

## llms.txt

flow publishes an [`llms.txt`](https://flowexec.io/llms.txt) index following the [llmstxt.org](https://llmstxt.org) standard — a plain-text map of all documentation pages and schemas. Tools that support it can pull the full docs in one request.

```
https://flowexec.io/llms.txt
```

The `get_info` MCP tool returns this URL as well, so a connected assistant can find it without any prior knowledge of the site.

## JSON Schemas

Every flow file type has a published JSON schema. Adding a `yaml-language-server` comment gives you validation and autocomplete in any schema-aware editor:

```yaml
# yaml-language-server: $schema=https://flowexec.io/schemas/flowfile_schema.json
executables:
  - verb: deploy
    name: staging
    serial:
      execs:
        - cmd: npm run build
        - ref: infra/k8s:apply-staging
```

| File | Schema |
|------|--------|
| `*.flow` / `*.flow.yaml` | [flowfile_schema.json](https://flowexec.io/schemas/flowfile_schema.json) |
| `flow.yaml` (workspace) | [workspace_schema.json](https://flowexec.io/schemas/workspace_schema.json) |
| `*.flow.tmpl` (template) | [template_schema.json](https://flowexec.io/schemas/template_schema.json) |
| User config | [config_schema.json](https://flowexec.io/schemas/config_schema.json) |
