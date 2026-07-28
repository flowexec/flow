---
title: Run Provenance
---

# Run Provenance

Every execution flow records answers *what ran*. Provenance answers **who ran it** — a terminal,
a GUI, or an AI assistant, and which one.

This page is the contract for setting it. For reading it back, see
[History & Logs](execution-history.md).

## What gets recorded

| Field | Meaning |
|---|---|
| `source` | How the run reached flow — `cli`, `desktop`, or `mcp` |
| `clientName` | Who drove it — e.g. `claude-code`, `cursor` |
| `sessionId` | What it groups with, so one assistant's related runs stay together |
| `workingDir` | Where it executed — the directory, not just the workspace |

`source` is compared as a plain string throughout, so the set is open. The three values above
are what flow itself produces; anything embedding flow can record its own without waiting for
a release.

## MCP runs: automatic

Nothing to configure. When an assistant runs something through the
[MCP server](ai-tools.md) — a named `execute`, a `run_command`, or a `run_executable` — the
server tags it before handing off to the CLI:

```shell
flow logs --source mcp --limit 1 -o json
```
```json
{
  "ref": "exec myproject/adhoc-run-tests",
  "source": "mcp",
  "clientName": "claude-code",
  "sessionId": "4ce5bfea-7ada-4b20-b59e-bcc0df500181",
  "command": "go test ./...",
  "label": "run the test suite"
}
```

### Reading your own session ID

A client that wants to group its runs later needs the session ID it will be tagged with.
Over stdio nothing in the transport carries it, so the server reports it in `get_info`:

```json
{
  "currentContext": {
    "workspace": "myproject",
    "sessionId": "4ce5bfea-7ada-4b20-b59e-bcc0df500181"
  }
}
```

Call it once when you connect, keep the value, and query with it later:

```shell
flow logs --session 4ce5bfea-7ada-4b20-b59e-bcc0df500181
```

### What a session is

**Whatever the caller says belongs together.** flow resolves it in this order:

1. a per-connection ID from the transport, when one exists
2. otherwise `FLOW_RUN_SESSION` from the environment — the caller told us
3. otherwise a UUID minted once per `flow mcp` process

Step 2 is what keeps a single conversation intact. An assistant does not only use MCP; it
also shells out to `flow` directly. If the server invented an ID while the shell inherited a
different one, the same conversation would land in two groups. A caller that tags its
environment gets both paths under one ID.

The minted UUID is a last resort, and it approximates a conversation rather than being one —
a client that reconnects mid-conversation gets a new one. If you know your own conversation
identity, supply it and flow will use it instead.

## Everything else: three environment variables

Runs that arrive any other way — an assistant shelling out, a script, an app embedding flow —
would otherwise be indistinguishable from you typing at a prompt. flow reads provenance from
the environment for **every** run, so anything that can export a variable can attribute itself:

| Variable | Sets | Default |
|---|---|---|
| `FLOW_RUN_SOURCE` | `source` | `cli` |
| `FLOW_RUN_CLIENT` | `clientName` | empty |
| `FLOW_RUN_SESSION` | `sessionId` | empty |

```shell
FLOW_RUN_CLIENT=claude-code FLOW_RUN_SESSION=$MY_SESSION_ID flow run build
```

```shell
flow logs --client claude-code    # finds it, even though source is still "cli"
```

Leaving `FLOW_RUN_SOURCE` alone is usually right: the run genuinely did arrive through the
CLI, and `clientName` is what says an assistant drove it. Set it when the surface itself is
distinct — a desktop app sets `desktop` so its runs are not confused with a terminal's.

> **Export it, don't pass it.** Environment beats a flag the assistant must remember on every
> call. A model can silently omit an argument; it cannot omit a variable it never sees. This
> is also why identity is not a tool parameter — parameters are for intent, which only the
> model knows.

### Referencing another variable

A session ID is only known once the assistant is running, so it cannot be written into a
config file as a constant. Most harnesses store that file's values **verbatim** — no `${...}`
interpolation — which would otherwise leave you recording a literal `${...}` on every run.

So flow resolves the reference itself. Give it `${NAME}` and it reads `NAME` from the
environment:

```shell
FLOW_RUN_SESSION='${MY_HARNESS_SESSION_ID}' flow run build
# records the value of MY_HARNESS_SESSION_ID
```

If `NAME` is unset the session is recorded empty, never as the literal — one shared constant
across every run would group unrelated work together, which is worse than recording nothing.

### Example: an editor or agent harness

Most agent tools let you define environment variables for the commands they run. Point flow at
the variable the tool already exports for its own session:

```json
{
  "env": {
    "FLOW_RUN_CLIENT": "claude-code",
    "FLOW_RUN_SESSION": "${CLAUDE_CODE_SESSION_ID}"
  }
}
```

That works whether or not the tool expands `${...}` itself — flow handles it either way. The
variable's name lives in your config rather than in flow, so when a vendor renames theirs you
edit one line here instead of waiting for a flow release.

### Example: embedding flow behind a UI

Set the variables on the process you spawn:

```go
cmd := exec.Command("flow", "run", "build")
cmd.Env = append(os.Environ(),
    "FLOW_RUN_SOURCE=desktop",
    "FLOW_RUN_CLIENT=my-app",
)
```

A wrapper that also runs an MCP server can set `FLOW_RUN_SESSION` on it too, and the server
will use that instead of minting its own — so runs the assistant makes over MCP and runs the
wrapper makes directly share one session.

## What flow deliberately does not do

**No client registry.** flow does not sniff for `CLAUDE_CODE_SESSION_ID`, `CURSOR_*`, or any
other vendor's variables. Those are undocumented internals that get renamed, and detection
built on them fails silently — history quietly stops grouping and nobody notices. Each tool
maps its own variables onto the contract above; flow stays neutral.

**No conversation identity of its own.** A session is a grouping key, not a claim about what a
conversation is. flow will happily record a conversation ID you hand it — that is what step 2
above is for — but it never goes looking for one, and it stores nothing that points back at a
chat. When nobody supplies an identity, the minted per-process UUID is an approximation, and
flow treats it as one. Deciding what a conversation *is* stays with the client, which is the
only thing that can know.

## Querying it

```shell
flow logs --source mcp                    # only assistant-launched runs
flow logs --client cursor                 # one client
flow logs --session <id>                  # one connection's runs
flow logs --source mcp --status failed    # what an assistant broke
```

Assistants can review their own activity with `get_execution_logs`, which takes the same
filters plus `mine: true` — scoped to the calling session without needing to know its ID.

## What's Next?

- [History & Logs](execution-history.md) — reading, filtering, and clearing history
- [AI Tools](ai-tools.md) — setting up the MCP server
