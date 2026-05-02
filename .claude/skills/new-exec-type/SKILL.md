---
name: new-exec-type
description: Add a new executable type to the flow runner (a new kind of automation block users can define in .flow files).
disable-model-invocation: true
argument-hint: "<type-name> — description of what this executable type does"
allowed-tools: Bash(flow generate:*) Bash(flow validate:*) Bash(flow build:*) Bash(go test:*) Read
---

Add a new executable type to the flow runner for: $ARGUMENTS

An "executable type" is a new automation block users can declare in `.flow` files (like `exec`, `serial`, `parallel`, `render`). Follow these steps in order:

1. **Schema first** — add the new type's fields to `types/executable/schema.yaml`.
   Read the existing schema to match the structure. Run `flow generate` to regenerate `types/executable/generated.go`.

2. **Runner handler** — create `internal/runner/<type>.go` implementing the runner interface.
   Read `internal/runner/exec.go` or `internal/runner/serial.go` as reference for the exact interface and pattern.

3. **Register the type** — wire the new handler into `internal/runner/runner.go` (the dispatch table).

4. **Parser support** — update `internal/fileparser/` if needed to recognize and validate the new type during YAML parsing.

5. **Tests** — add unit tests in `internal/runner/<type>_test.go` using Ginkgo.
   Use `Describe`/`It`/`Entry` — never `FDescribe`/`FIt`. Cover happy path and error cases.

6. **Validate** — run `flow validate` to confirm generate, lint, and tests all pass.

Do not skip the schema step — editing generated files directly will cause CI to fail.
