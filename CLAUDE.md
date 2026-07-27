# flow repo — Claude Code Context

**flow** is a workflow automation hub: automation organized across projects (workspaces), with
built-in secrets, templates, and cross-workspace composition. Users define workflows in YAML flow
files and run them anywhere. This repo is the flow CLI (Go), and flow runs its own dev automation
via `.execs/*.flow`.

---

## Critical Rules

1. **Only `*.gen.go` is generated — the rest of `types/` is hand-written.** Never edit
   `types/**/*.gen.go`, `docs/cli/`, `docs/types/`, or `docs/public/schemas/`; edit the source
   schema and run the `generate` executable. Sources are `types/executable/*_schema.yaml` and
   `types/{config,workspace,common}/schema.yaml`. CI's `validate generated` fails on uncommitted
   diffs. (Permission rules block these writes, so a failed edit here is expected, not a bug.)
2. **Remove `FDescribe`/`FIt`/`FEntry` before committing** — they silently exclude every other
   spec in the suite.
3. **In `cmd/`, never `log.Fatal`, `os.Exit`, or `logger.Log().FatalErr()`** — use
   `errhandler.HandleFatal(ctx, cmd, err)`, or `HandleUsage` for flag/arg misuse.
4. **`pkg/` is importable API surface; `internal/` is not.**
5. **Run work through flow's MCP tools, not raw Bash** — see Development Workflow.
6. **`go test ./...` without build tags silently skips most tests.** Use the `test` refs.
7. **Never commit directly on `main`.** Before the first commit of any change, create a branch
   (`git switch -c <type>/<slug>`) so the work lands as a PR. If you're already on `main` with
   commits, move them onto a branch and reset `main` back to `origin/main`.

---

## Repository Structure

```
flow/
├── cmd/                 # Cobra CLI. Only root.go here; handlers in cmd/internal/
├── pkg/                 # Importable: cache, cli, context, errors, filesystem,
│                        #   imports, logger, store
├── internal/            # Not importable outside the binary
│   ├── io/              # Terminal UI and output rendering (wraps tuikit)
│   ├── mcp/             # MCP server (tools, resources)
│   ├── runner/          # Execution engine; one subpackage per executable type
│   ├── services/        # Business logic orchestration
│   ├── templates/       # Workflow template expansion
│   ├── vault/           # Thin wrapper over the vault module
│   └── ...              # fileparser, updater, utils, validation, version
├── types/               # Schemas + generated types (*.gen.go) + hand-written helpers
├── tests/               # E2E suite (Ginkgo, -tags=e2e)
├── docs/                # flowexec.io source; docs/cli and docs/types are generated
└── .execs/              # flow's own dev automation
```

---

## Architecture

```
cmd/internal (Cobra) → pkg/context (workspace + config) → internal/services
  → internal/runner → type-specific subpackage
```

`internal/mcp` exposes that same pipeline over the Model Context Protocol; `flow mcp` starts the
server.

**Scope boundary — flow is an AI *tool provider*, not an AI *consumer*.** The core exposes
deterministic capabilities (MCP server, published JSON schemas, `llms.txt`). Do not add LLM calls,
natural-language command parsing, or AI generation *into* the CLI — that puts vendor keys,
per-call cost, and non-determinism in a task runner's critical path. Anything applying a model to
flow does so from outside, via the MCP surface. Treat "add AI features to the CLI" as out of scope.

---

## Sibling Repositories

Two first-party modules carry much of this repo's behavior, and their seam is where most breakage
happens:

- **`flowexec/tuikit`** — all TUI rendering (`flow browse`, logs view, prompts). Apparent
  rendering or input bugs usually live here, not in `internal/io`.
- **`flowexec/vault`** — secret storage providers (AES-256, age, keyring, env). `internal/vault`
  is a thin type-alias wrapper.

**Read the pinned version, not a local checkout.** There are no `replace` directives, so builds
use the `go.mod` versions from `$(go env GOMODCACHE)/github.com/flowexec/<mod>@<version>`. A
sibling working copy is often on a feature branch at a *different* version than what compiles, so
answering an integration question from it yields confidently wrong results. Use a checkout only
when deliberately co-developing upstream. Upgrades here are usually breaking-change adaptations,
not version bumps — check release notes before assuming an API is unchanged.

Same org, not imported: `action` (GitHub Action), `examples`, `homebrew-tap`.

---

## Development Workflow

flow's MCP server is wired up in `.mcp.json`. **Prefer `mcp__flow__*` over raw Bash** so runs get
workspace env/secrets and land in flow's history. The `flow-context` skill has the full guidance.

`mcp__flow__execute` runs a named executable, `run_command` a one-off shell command,
`run_executable` an inline multi-step spec. Discover names with `mcp__flow__list_executables` —
don't assume them.

| Ref | What it does |
|-----|--------------|
| `build binary` | Build the CLI (pass `./bin/flow` as the output arg) |
| `test` | All tests (unit + e2e), parallel |
| `test unit` / `test e2e` | One suite |
| `lint` | golangci-lint |
| `generate` | All code generation |
| `validate` | generate → lint → test → generated-diff check |
| `install tools` | Install/update Go tools |

`flow browse` and `flow mcp` are interactive — they belong in a real terminal, not a tool call.

**Testing notes:** e2e needs the binary on PATH (the `test e2e` ref builds it first). Set
`UPDATE_GOLDEN_FILES=true` when output changes are intentional.

---

## Error Handling

The CLI emits a structured error envelope (`{"error":{"code","message","details"}}`) on stderr for
`--output json|yaml`, plain text otherwise. Both paths go through `cmd/internal/errors.HandleFatal`.

Typed errors in `pkg/errors/errors.go` implement `Code() string`. Extend that set rather than
returning bare `fmt.Errorf` when a stable machine-readable code matters. Codes: `INVALID_INPUT`,
`NOT_FOUND`, `EXECUTION_FAILED`, `TIMEOUT`, `CANCELLED`, `VALIDATION_FAILED`, `INTERNAL_ERROR`,
`PERMISSION_DENIED`.

---

## PR & Code Quality

Run the `validate` executable before marking a PR ready. Commit messages: imperative, lowercase,
≤72 chars (`fix:`, `feat:`, `refactor:`). No focus markers, debug prints, or open TODOs.

---

## Configuration Files

- **`flow.yaml`** — this repo's workspace config
- **`.execs/`** — flow's dev automation definitions
- **`.mcp.json`** — registers the flow MCP server; committed so every clone gets it
- **`.claude/settings.json`** — committed permissions. `deny` blocks generated-file writes and
  secret reads; `ask` gates pushes and releases. Keep absolute paths out — it ships to everyone.
- **`.claude/settings.local.json`** — gitignored, per-user. Where absolute paths belong, e.g.
  `permissions.additionalDirectories` for the module cache and sibling checkouts.

## Development Setup

1. Go 1.25+, flow CLI installed
2. `flow workspace add flow . --set`
3. `flow install tools`
4. `flow validate`
