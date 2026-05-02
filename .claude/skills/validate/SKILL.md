---
name: validate
description: Run flow validate and fix any failures. Invoke after completing a feature or bug fix to confirm the codebase is clean before committing.
allowed-tools: Bash(flow validate:*) Bash(flow generate:*) Bash(flow lint:*) Bash(flow test:*) Bash(go test:*) Read
---

Run `flow validate` — it runs these steps in order: `generate` → `lint` → `test` → `validate generated` (checks for uncommitted generated diffs in CI).

For each failure, diagnose and fix before moving on:

- **generate fails**: Schema syntax error in `types/*/schema.yaml` — read and fix the schema
- **lint fails**: Read the golangci-lint output, fix each violation, re-run
- **test fails**: Read the Ginkgo output, identify the failing spec, fix the root cause — do not skip or comment out tests
- **validate generated fails**: Generated files are out of sync — run `flow generate` and stage the regenerated files; this is always the fix

Do not report done until `flow validate` exits 0 with all steps passing.
