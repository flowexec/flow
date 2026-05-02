---
name: pr-ready
description: Run a pre-PR readiness check and report READY or NOT READY.
disable-model-invocation: true
allowed-tools: Bash(git *) Bash(flow validate:*) Bash(flow generate:*) Bash(go test:*) Read
---

Check whether the current branch is ready to open a PR. Work through each item and report PASS or FAIL:

1. **No focus markers** — `grep -rn "FDescribe\|FIt\|FEntry\|FContext\|FWhen" --include="*.go" .`
   Any match is a FAIL — these silently exclude all other tests in the suite.

2. **Validation passes** — run `flow validate`. All steps must pass.

3. **No debug artifacts** — grep for `fmt.Println`, `spew.Dump` in `cmd/`, `internal/`, `pkg/`.
   Flag anything that looks like leftover debug output (not legitimate logging).

4. **Commit message format** — `git log main..HEAD --oneline`.
   Each commit should be: imperative, lowercase, ≤72 chars (`fix: ...`, `feat: ...`, `refactor: ...`).

5. **No direct edits to generated files** — `git diff main...HEAD --name-only` should not show changes to `types/**/*.go`, `docs/cli/*.md`, or `docs/types/*.md` in isolation (without a corresponding schema change).

After all checks: report overall **READY** or **NOT READY** with a concise summary of what needs fixing. Do not open a PR automatically — let the user decide.
