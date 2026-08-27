---
title: Containers
description: "Run a single flow executable inside a container image with exec.container, or run the whole CLI from ghcr.io/flowexec/flow in a CI pipeline."
---

# Containers

There are two separate things you might mean by "run flow in a container", and they solve different
problems:

- **[A step runs in a container](#running-a-step-in-a-container)** — flow runs on your machine and
  puts one executable inside an image. Use this when a task needs a toolchain you would rather not
  install.
- **[flow runs in a container](#running-flow-in-a-container)** — the whole CLI runs inside an image.
  Use this for CI platforms with no first-party integration, or for fully isolated execution.

## Running a step in a container

Any `exec` executable can declare a `container`. flow resolves parameters, arguments, and secrets on
the host as usual, then runs the command inside the image with the workspace mounted:

```yaml
executables:
  - verb: build
    name: app
    exec:
      container:
        image: golang:1.26-bookworm
      cmd: go build -o bin/app ./cmd/app
```

The workspace root is mounted at `/workspace` and the executable's directory becomes the working
directory inside it, so `cmd` reads exactly as it would on the host. Nothing else about the
executable changes — it still composes, still takes parameters, still appears in `flow browse`.

### Options

| Field | Default | What it does |
|-------|---------|--------------|
| `image` | *required* | Image to run in, e.g. `golang:1.26-alpine` |
| `runtime` | `auto` | `docker` or `podman`; `auto` prefers Docker and falls back to Podman |
| `workdir` | the executable's `dir` | Working directory inside the container |
| `mountWorkspace` | `/workspace` | Where the workspace root is mounted |
| `volumes` | `[]` | Extra bind mounts, `host:container` or `host:container:options` |
| `inheritEnv` | `true` | Pass flow's resolved environment in |
| `entrypoint` | matches the interpreter | `sh` for a shell command, `python3` for a Python one; set empty to use the image's own `ENTRYPOINT` |
| `user` | current host user on Linux | `uid`, `uid:gid`, or a name |
| `network` | runtime default | e.g. `host`, `none`, or a named network |

### What crosses the boundary

Two defaults are worth knowing, because they are what make containerised steps behave predictably:

**Environment.** `inheritEnv` passes flow's *resolved* environment — parameters, arguments, and
`FLOW_*` variables — into the container. The host process environment is **never** forwarded,
whatever you set it to. A container step therefore sees the values the executable declares and
nothing it happened to inherit from your shell, which is the same isolation CI gives you.

Host interpreter discovery stops at the container boundary too: `VIRTUAL_ENV`, `PYTHONPATH`,
`PYTHONHOME`, and `FLOW_PYTHON_BIN` are dropped regardless of `inheritEnv`. Those paths either do
not exist inside the image or, worse, resolve to an unrelated mounted directory.

**File ownership.** On Linux, `user` defaults to the current host user, so files the container
writes into the mounted workspace are not owned by `root`. Set `user: root` to opt out if the image
needs it.

### A fuller example

```yaml
executables:
  - verb: test
    name: integration
    exec:
      params:
        - envKey: DATABASE_URL
          secretRef: test-db-url
      container:
        image: golang:1.26-bookworm
        network: host
        volumes:
          - /var/run/docker.sock:/var/run/docker.sock
        workdir: /workspace/tests
      cmd: go test -tags=integration ./...
```

`DATABASE_URL` is decrypted from the vault on the host and handed to the container as an
environment variable. The secret never touches the image or the mounted files.

### A pinned Python toolchain

`interpreter: python` composes with `container`, which is the tidiest way to run Python against an
exact version without installing it locally:

```yaml
executables:
  - verb: run
    name: report
    exec:
      interpreter: python
      container:
        image: python:3.13-alpine
      cmd: |
        import sys
        print(sys.version)
```

flow writes the inline code to a script and bind-mounts it read only rather than passing it with
`-c`, so tracebacks report real line numbers and the code never lands in the process table. The
entrypoint defaults to `python3`, resolved on the image's `PATH` — which is what lets any Python
image work with no further configuration.

Setting `entrypoint: ""` hands execution to the image's own `ENTRYPOINT`; with Python that only
works if the image's entrypoint is itself an interpreter.

> [!TIP]
> Reach for a container when the *toolchain* is the problem — a pinned compiler, a linter you do not
> want globally installed. For steps that only need what is already on the machine, plain `exec` is
> faster and simpler.

## Running flow in a container

The published image at `ghcr.io/flowexec/flow` carries the CLI and enough scaffolding to register a
workspace on startup.

```shell
# Drop into the default workspace
docker run -it --rm ghcr.io/flowexec/flow

# Run one executable
docker run -it --rm ghcr.io/flowexec/flow validate
```

The entrypoint reads three environment variables:

| Variable | What it does |
|----------|--------------|
| `REPO` | Repository URL to clone (defaults to flow's own repo) |
| `BRANCH` | Branch to check out |
| `WORKSPACE` | Name to register the workspace under (defaults to `flow`) |

### Workspace from git

```shell
docker run -it --rm \
  -e REPO=https://github.com/your-org/your-workspace \
  -e BRANCH=main \
  -e WORKSPACE=my-workspace \
  ghcr.io/flowexec/flow exec "deploy app"
```

### Workspace from your machine

```shell
docker run -it --rm \
  -v $(pwd):/workspaces/my-workspace \
  -w /workspaces/my-workspace \
  -e WORKSPACE=my-workspace \
  ghcr.io/flowexec/flow exec "build app"
```

## In CI/CD pipelines

Any platform that can run a container can run flow. The shape is always the same:

1. Use `ghcr.io/flowexec/flow` as the job image.
2. Set `REPO`, `BRANCH`, and `WORKSPACE`, or mount the checkout directly.
3. Call `flow exec "your executable"`.

```yaml
# GitLab CI
test:
  image: ghcr.io/flowexec/flow
  variables:
    WORKSPACE: my-project
  script:
    - flow exec "test unit"
```

On GitHub Actions, prefer the [first-party action](github-actions.md) — it handles vaults, multiple
workspaces, and structured error codes for you.

> [!NOTE]
> The container image is less exercised than the CLI and the GitHub Action. It should work on any
> Docker-capable platform, but if you run flow somewhere new we would like to hear how it went —
> open an issue or say so in [Discord](https://discord.gg/CtByNKNMxM).
