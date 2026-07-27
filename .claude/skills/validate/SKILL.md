---
name: validate
description: Run flow validate and fix any failures. Invoke after completing a feature or bug fix to confirm the codebase is clean before committing.
allowed-tools: mcp__flow__execute, mcp__flow__run_command, mcp__flow__list_executables, mcp__flow__get_execution_logs, Bash(flow validate:*), Bash(flow generate:*), Bash(flow lint:*), Bash(flow test:*), Bash(go test:*), Read
---

Run the `validate` executable via `mcp__flow__execute` (ref: `validate`) rather than a raw shell
call, so the run inherits workspace env/secrets and is captured in flow's history.

It runs these steps in order: `generate` → `lint` → `test` → `validate generated` (checks for uncommitted generated diffs in CI).

For each failure, diagnose and fix before moving on:

- **generate fails**: Schema syntax error in the source schema — `types/executable/*_schema.yaml`, or `types/{config,workspace,common}/schema.yaml`. Read and fix the schema, never the `.gen.go` output.
- **lint fails**: Read the golangci-lint output, fix each violation, re-run
- **test fails**: Read the Ginkgo output, identify the failing spec, fix the root cause — do not skip or comment out tests
- **validate generated fails**: Generated files are out of sync — re-run the `generate` executable and stage the regenerated files; this is always the fix

Use `mcp__flow__get_execution_logs` with `mine: true` to re-read output from a run instead of
re-running it.

Do not report done until `validate` exits 0 with all steps passing.
