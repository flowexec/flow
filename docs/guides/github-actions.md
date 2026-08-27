---
title: GitHub Actions
description: "Run your flow executables in GitHub Actions with flowexec/action: inputs, outputs, vault secrets, multi-workspace setups, and matrix builds."
---

# GitHub Actions

The [`flowexec/action`](https://github.com/marketplace/actions/flow-execute) action installs the
flow CLI, registers your repository as a workspace, and runs one executable. Anything you can run
locally runs the same way in CI — same flow file, same reference, same environment resolution.

## Quickstart

```yaml
name: CI
on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: flowexec/action@v1
        with:
          executable: 'build app'
```

`executable` takes the same `VERB NAME` reference you would type after `flow exec`. This is the
whole point of running flow in CI: the pipeline stops being a second copy of your build logic and
becomes a one-line call into the automation you already maintain.

flow's own pipelines work this way — see
[`.github/workflows`](https://github.com/flowexec/flow/tree/main/.github/workflows) in this repo.

## Inputs

### Choosing what runs

| Input | Default | What it does |
|-------|---------|--------------|
| `executable` | *required* | The executable reference to run, e.g. `test unit` |
| `workspace` | `.` | Workspace to run in, as a path or a registered name |
| `workspace-name` | auto | Name to register the workspace under |
| `workspaces` | — | A YAML/JSON map of several workspaces, local or git-sourced |
| `working-directory` | `.` | Directory to run flow from |
| `flow-version` | `latest` | Pin the CLI version |

### Passing values in

| Input | Default | What it does |
|-------|---------|--------------|
| `params` | — | `KEY=VALUE` pairs bound to the executable's parameters |
| `env` | — | `KEY=VALUE` pairs exported for the run |
| `secrets` | — | `KEY=VALUE` pairs written into a vault before the run |
| `vault-key` | — | Encryption key for an existing vault |

### Controlling the run

| Input | Default | What it does |
|-------|---------|--------------|
| `timeout` | `30m` | Abort the executable after this long |
| `upload` | `false` | Upload flow's logs as a workflow artifact |
| `continue-on-error` | `false` | Keep the job green when the executable fails |
| `clone-token` | — | Token used to clone private git workspaces |
| `clone-depth` | `1` | Clone depth for git workspaces; `0` for full history |

## Secrets

Secrets go into a flow vault for the duration of the job, then are injected into the executable
exactly as they would be on your machine — so the executable does not need a CI-specific branch.

```yaml
      - uses: flowexec/action@v1
        with:
          executable: 'deploy app'
          params: |
            ENVIRONMENT=production
          secrets: |
            API_KEY=${{ secrets.API_KEY }}
            DB_PASSWORD=${{ secrets.DB_PASSWORD }}
```

If you do not pass `vault-key`, the action generates one and returns it as an output, so a later
step in the same job can reuse the vault.

> [!IMPORTANT]
> Pass secrets through the `secrets` input rather than `env`. Values arriving through `secrets` are
> stored in the vault and referenced by name, which keeps them out of the executable's rendered
> command line.

## Multiple workspaces

Composed automation often spans repositories. `workspaces` registers several at once, including
ones cloned from git:

```yaml
      - uses: flowexec/action@v1
        with:
          executable: 'deploy platform'
          workspaces: |
            app: .
            infra: https://github.com/your-org/infra.git
          clone-token: ${{ secrets.GITHUB_TOKEN }}
```

Cross-workspace references (`infra/k8s:apply`) then resolve the same way they do locally. See
[Workspaces](workspaces.md) for how references are addressed.

## Outputs

| Output | What it carries |
|--------|-----------------|
| `exit-code` | Exit code of the executable |
| `output` | Captured stdout |
| `error-code` | Machine-readable failure code, e.g. `EXECUTION_FAILED`, `TIMEOUT`, `NOT_FOUND` |
| `vault-key` | The generated vault key, when secrets were configured without one |

`error-code` is the same stable code the CLI reports in its
[error envelope](/guides/execution-history), so a job can branch on *why* something failed rather
than parsing log text:

```yaml
      - uses: flowexec/action@v1
        id: run
        with:
          executable: 'test e2e'
          continue-on-error: 'true'
      - if: steps.run.outputs.error-code == 'TIMEOUT'
        run: echo "::warning::e2e timed out; retrying nightly"
```

## Matrix builds

The action runs on Linux, macOS, and Windows runners, so one matrix covers every platform your
workflow supports:

```yaml
jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - uses: flowexec/action@v1
        with:
          executable: 'test unit'
          upload: 'true'
```

## Other CI platforms

There is no first-party action for GitLab CI, CircleCI, or Jenkins. Run flow through its container
image instead — see [Containers](containers.md#in-ci-cd-pipelines).
