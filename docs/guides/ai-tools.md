---
title: AI Tools
description: "Expose your workflows to Claude Code, Cursor, and any MCP client with flow's MCP server, plus llms.txt, JSON schemas, and a committed skill file."
---

# AI Tools

`flow mcp` exposes your automation over the [Model Context Protocol](https://modelcontextprotocol.io).
A connected assistant discovers, runs, and writes executables through the same engine you use by
hand — with the workspace's environment, secrets injected at run time, and every run recorded in
flow's history.

flow provides the tools and stops there. It makes no model calls of its own and holds no vendor
keys; the assistant is yours, and flow stays a deterministic task runner underneath it.

## Setup

Put this in `.mcp.json` at the root of your repository and commit it:

```json title=".mcp.json"
{
  "mcpServers": {
    "flow": {
      "type": "stdio",
      "command": "flow",
      "args": ["mcp"]
    }
  }
}
```

That is the entire setup. Claude Code and Cursor both read `.mcp.json` from the repository root, so
committing it means every teammate and every fresh clone gets flow's tools with no per-person
configuration. Each user still approves the server on first use, so the file grants no access on its
own — it only removes the setup step. Other MCP clients take the same object under their own config
path.

This repo does exactly that: see
[`.mcp.json`](https://github.com/flowexec/flow/blob/main/.mcp.json) in the flow source.

Connecting the server gives your assistant the tools; it does not make it *reach* for them.
[Teach it to prefer flow](#teaching-your-assistant-to-use-it) with a committed instruction file.

## What the server exposes

**Tools**

| Tool | Description |
|------|-------------|
| `get_info` | Current workspace, schema URLs, and documentation index |
| `list_workspaces` | All registered workspaces |
| `get_workspace` | Details and config for a specific workspace |
| `switch_workspace` | Change the active workspace |
| `list_executables` | Browse executables — filterable by tag, verb, workspace, and resolvable from a `dir` |
| `get_executable` | Full definition and metadata for a specific executable |
| `execute` | Run a **named** executable by ref, in a given `dir` or `workspace` |
| `run_command` | Run one or more **arbitrary** shell commands through flow (with a `label`, working `dir`, and optional `workspace`) — captured in history like any executable |
| `run_python` | Run **Python** code through flow (with `code`, a `label`, working `dir`, and optional `workspace`) — uses the workspace's virtualenv when there is one |
| `run_executable` | Run a **transient executable of any type** from an inline `spec` — a serial/parallel batch, an HTTP `request`, a `render`, or a `launch` — without saving a file |
| `get_execution_logs` | Output from recent runs, filterable by `source`/`session`/`status`, or `mine` for this session's own runs |
| `sync_executables` | Refresh cached workspace and executable state |
| `write_flowfile` | Create or update a `.flow` file, validated before writing |

The run tools form a ladder, closest-fit first: **`execute`** for a task you've already named, **`run_command`** for a one-off shell command, **`run_python`** when the one-off is Python rather than shell, **`run_executable`** for something richer than a single command. Reaching for flow before a raw shell tool means every run inherits the workspace's environment and secrets and is recorded — see [Observability](#observability) below.

**Working in a worktree or a fresh clone**

The MCP server inherits whatever directory it was started in, which is often not where you are
working. Pass `dir` on `execute`, `run_command`, `run_python`, `run_executable`, or `list_executables` and flow
resolves the workspace by walking up from *that* directory to the nearest `flow.yaml` — so a git
worktree or a just-cloned repo works without being registered first. `get_info` reports
`workspaceRegistered` and `workspaceSource` so you can tell which case you are in; an
unregistered workspace runs normally but cannot be switched to. Exporting `FLOW_WORKSPACE` in the
server's environment pins it for every call instead. See
[Unregistered workspaces](workspaces.md#unregistered-workspaces).

**Prompts**

Structured prompts the assistant can invoke for common tasks:

| Prompt | Description |
|--------|-------------|
| `generate_executable` | Generate a new executable for a described task |
| `generate_project_executables` | Generate a full automation set for a project |
| `debug_executable` | Debug a failing executable |
| `migrate_automation` | Convert existing Makefile, npm scripts, or shell scripts to flow |
| `explain_flow` | Explain flow concepts and configuration |

## Teaching your assistant to use it

Ask an assistant to run the tests and it will happily type `go test ./...` instead of calling your
`test` executable — the tools being available is not the same as the tools being used.

The fix is a small instruction file committed alongside the `.mcp.json`. Claude Code calls these
[background skills](https://agentskills.io): they load on every session with no manual invocation.

Create `.claude/skills/flow-context/SKILL.md`:

```markdown
---
name: flow-context
description: This project uses flow for automation. When asked to build, test, run, deploy, lint, generate, or perform any dev task — or to run any one-off shell command — prefer the flow MCP tools over raw Bash so the work runs with workspace env/secrets and is recorded in flow's history.
user-invocable: false
---

This repository uses **flow** for all development automation. The `mcp__flow__*` MCP tools are
available — prefer them over raw `Bash` for anything runnable, so it executes with the workspace's
environment and secrets and is captured in flow's execution history.

## Running work — pick the closest tool

1. **Named task?** (build, test, lint, validate, generate, deploy, …) → `mcp__flow__list_executables`
   to find it, then `mcp__flow__execute` with its verb + name. Don't hand-roll a shell command a
   flow executable already covers.
2. **Arbitrary one-off command?** (a `git ...`, a script) → `mcp__flow__run_command` with the
   command and a short `label`. It runs with workspace env/secrets and lands in `flow logs` with
   provenance. Pass `commands` (array) + `mode` (`serial`/`parallel`) to run several in one call.
3. **One-off is Python?** → `mcp__flow__run_python` with `code` and a short `label`, rather than
   `python -c` or a scratch `.py` file. flow resolves the workspace's virtualenv, so imports see
   the project's dependencies and tracebacks report real line numbers.
4. **Something richer than one command?** (a serial/parallel batch, an HTTP `request`) →
   `mcp__flow__run_executable` with an inline `spec`.
5. Only fall back to `Bash` when a command genuinely shouldn't be recorded or flow isn't the right
   tool (e.g. interactive/TTY programs).

## Key executables in this repo

Common refs: `test unit`, `lint`, `build app`. Use `mcp__flow__list_executables` to discover
current names — don't assume them.

## Context & authoring

- Call `mcp__flow__get_info` at the start of a session, or when you need schema URLs to author
  `.flow` files.
- Author or edit flow files with `mcp__flow__write_flowfile` (validated server-side) rather than
  writing YAML by hand.
- To review what you've run this session, call `mcp__flow__get_execution_logs` with `mine: true`;
  `source`/`session`/`status` filter history more broadly.
```

`user-invocable: false` keeps it out of the `/` command menu — it is not something you invoke, it is
just always in context.

The **Key executables** section is the part worth tailoring: name the handful of refs your project
actually uses, and tell the assistant to discover the rest rather than guess. flow's own copy lives
at [`.claude/skills/flow-context/SKILL.md`](https://github.com/flowexec/flow/blob/main/.claude/skills/flow-context/SKILL.md)
if you want a working reference.

## Observability

Every run flow launches — whether a named `execute`, a `run_command`, a `run_python`, or a `run_executable` — is recorded as one lifecycle-aware history entry: written as `running` when it starts and updated to `completed` or `failed` when it finishes. Runs launched over MCP also capture **provenance**: which tool called (`claude`, `cursor`, …) and its session ID. That turns flow's history into an audit trail of what your assistant did.

Query it from the CLI:

```shell
flow logs                              # everything, most recent first
flow logs --status running             # what's in flight right now
flow logs --source mcp                 # only runs launched by an AI client
flow logs --session <id>               # everything one agent session ran
flow logs --client cursor --status failed
```

The interactive `flow logs` view shows an **Origin** column (the client, or `cli`/`mcp`) so you can see at a glance which runs came from an assistant; opening a record reveals its full source, client, and session.

The assistant can review its own activity too: `get_execution_logs` accepts the same `source`/`session`/`status` filters, plus `mine: true` — a shortcut that scopes results to the calling session's own runs, so it can check "what have I run so far?" without guessing a session ID.

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

## Understanding flow itself

Everything above is about running *your* automation. A different question — how flow itself works,
why a reference resolves the way it does, what the runner does on a failed step — is answered by the
Go source rather than by these guides.

[DeepWiki](https://deepwiki.com/flowexec/flow) indexes the `flowexec/flow` repository and answers
questions against it.

### From this site

Every page here has an **Ask AI** button in the nav. It sends your question to DeepWiki and renders
the answer inline; expect roughly fifteen seconds.

### From your assistant

DeepWiki runs a public MCP server. Adding it alongside flow's own server lets an assistant read how
flow works *and* run your workflows in the same session:

```json title=".mcp.json"
{
  "mcpServers": {
    "flow": {
      "type": "stdio",
      "command": "flow",
      "args": ["mcp"]
    },
    "deepwiki": {
      "type": "http",
      "url": "https://mcp.deepwiki.com/mcp"
    }
  }
}
```

It exposes three tools — `ask_question`, `read_wiki_structure`, and `read_wiki_contents` — and takes
any `owner/repo` it has indexed, so the same entry answers for your own dependencies too.

This one is genuinely optional. It is a third-party service, it needs network access, and unlike
flow's server it does put a model in the loop. Leave it out if either of those is a problem.

> [!NOTE]
> DeepWiki's answers are model-generated. They are grounded in real source, but they can be wrong or
> lag behind `main`. Treat them as a fast way to find the right file, and confirm anything
> load-bearing against the code or these guides.
