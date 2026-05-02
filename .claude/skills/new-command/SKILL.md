---
name: new-command
description: Scaffold a new Cobra CLI command following the project's patterns.
disable-model-invocation: true
argument-hint: "<verb> [noun] — what the command does"
allowed-tools: Bash(flow build:*) Bash(go build:*) Read
---

Scaffold a new Cobra CLI command for: $ARGUMENTS

Before writing any code, read a similar existing command to match the exact style:
- Simple verb commands: `cmd/exec.go`
- Noun/verb subcommands: any file under `cmd/workspace/` or `cmd/vault/`

Then follow these patterns:

1. **File location**: `cmd/<verb>.go` for top-level, or `cmd/<noun>/<verb>.go` for subcommands
2. **Command registration**: register in the parent command's `init()` or `cmd/root.go`
3. **Error handling**:
   - Runtime errors → `errhandler.HandleFatal(ctx, cmd, err)`
   - Flag/arg misuse → `errhandler.HandleUsage(ctx, cmd, "message", args...)`
   - Never use `log.Fatal`, `os.Exit`, or `logger.Log().FatalErr()` in `cmd/`
4. **Context**: resolve workspace context via `pkg/context` before delegating to `internal/services`
5. **Output**: respect `--output` flag (text/json/yaml) for structured responses

After scaffolding, verify it builds: `flow build binary ./bin/flow`
