---
name: new-command
description: Scaffold a new Cobra CLI command following the project's patterns.
disable-model-invocation: true
argument-hint: "<verb> [noun] — what the command does"
allowed-tools: mcp__flow__execute, mcp__flow__run_command, mcp__flow__list_executables, Bash(flow build:*), Bash(go build:*), Read
---

Scaffold a new Cobra CLI command for: $ARGUMENTS

Before writing any code, read a similar existing command to match the exact style:
- Simple verb commands: `cmd/internal/exec.go`
- Noun/verb subcommands: `cmd/internal/workspace.go` or `cmd/internal/vault.go` — each groups its
  subcommands in a single file rather than a directory

Then follow these patterns:

1. **File location**: `cmd/internal/<noun>.go`. Only `root.go` lives directly in `cmd/`; every
   command handler is under `cmd/internal/`. Shared helpers go in `cmd/internal/helpers.go`,
   flags in `cmd/internal/flags/`, output shaping in `cmd/internal/response/`.
2. **Command registration**: register on the parent command, or add to `rootCmd` in `cmd/root.go`
3. **Error handling**:
   - Runtime errors → `errhandler.HandleFatal(ctx, cmd, err)`
   - Flag/arg misuse → `errhandler.HandleUsage(ctx, cmd, "message", args...)`
   - Never use `log.Fatal`, `os.Exit`, or `logger.Log().FatalErr()` in `cmd/`
4. **Context**: resolve workspace context via `pkg/context` before delegating to `internal/services`
5. **Output**: respect `--output` flag (text/json/yaml) for structured responses

After scaffolding, verify it builds — prefer `mcp__flow__execute` with ref `build binary` and
argument `./bin/flow` over a raw shell call.
