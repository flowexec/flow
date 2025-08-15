# Contributing to flow

[![Go Report Card](https://goreportcard.com/badge/github.com/flowexec/flow)](https://goreportcard.com/report/github.com/flowexec/flow)
[![Go Reference](https://pkg.go.dev/badge/github.com/flowexec/flow.svg)](https://pkg.go.dev/github.com/flowexec/flow)
[![GitHub branch check runs](https://img.shields.io/github/check-runs/flowexec/flow/main)](https://github.com/flowexec/flow/actions?query=branch%3Amain)
[![Codecov](https://img.shields.io/codecov/c/github/flowexec/flow)](https://app.codecov.io/gh/flowexec/flow)

This document provides an overview of how to contribute to the flow project, including setting up your development environment, understanding the project structure, and running tests.

Before getting started, please read our [Code of Conduct](https://github.com/flowexec/flow/blob/main/.github/CODE_OF_CONDUCT.md) and [Contributing Guidelines](https://github.com/flowexec/flow/blob/main/.github/CONTRIBUTING.md).

**Ways to Contribute**

- **Report bugs** - [Open an issue](https://github.com/flowexec/flow/issues/new) with reproduction steps
- **Suggest features** - Share ideas for new functionality
- **Improve documentation** - Fix typos, add examples, or clarify explanations
- **Write code** - Fix bugs, implement features, or optimize performance
- **Share examples** - Contribute to the [examples repository](https://github.com/flowexec/examples)

## Quick Start 

**Prerequisites**

- **Go** - See [go.mod](https://github.com/flowexec/flow/blob/main/go.mod) for the required version
- **flow CLI** - Install the [latest version](installation.md) before developing

```sh
# Clone and set up the repository
git clone https://github.com/flowexec/flow.git
cd flow

# Register the repo as a flow workspace
flow workspace add flow . --set

# Install development dependencies
flow install tools

# Verify everything works
flow validate
```

## Development Executables

The flow project uses flow itself for development! Here are the key commands:

```sh
# Build the CLI binary
flow build binary ./bin/flow

# Run all validation (tests, linting, code generation)
flow validate

# Run specific checks
flow test                 # All tests
flow generate             # Code generation
flow lint                 # Linting only

# Install/update Go tools
flow install tools
```

## Project Structure

```
flow/
├── .execs/             # Development workflows
├── cmd/                # CLI entry point
├── docs/               # Documentation
├── internal/           # Core application logic
│   ├── cache/          # Executable and workspace caching logic
│   ├── context/        # Global application context
│   ├── io/             # Terminal user interface and I/O
│   ├── runner/         # Executable execution engine
│   ├── services/       # Business logic services
│   ├── templates/      # Templating system for workflows
│   └── vault/          # Secret management
├── tests/              # CLI end-to-end test suite
└── types/              # Generated types from schemas
```

_Some directories are omitted for brevity._

## Working with Generated Code

flow uses code generation extensively:

### Go CLI Type Generation <!-- {docsify-ignore} -->

Types are generated from YAML schemas using [go-jsonschema](https://github.com/atombender/go-jsonschema):

```sh
# Regenerate types after schema changes
flow generate cli
```

**Important**: When modifying types, edit the `schemas/*.yaml` files, not the generated Go files in `types/`.

### Documentation Generation <!-- {docsify-ignore} -->

CLI and type documentation is generated automatically:

```sh
# Updates both CLI docs and type reference docs
flow generate docs
```

## TUI Development

flow uses [tuikit](tuikit.md) for terminal interface development:

**Local Development**

```sh
# Link to local tuikit for development
go mod edit -replace github.com/flowexec/tuikit=../tuikit

# Test TUI changes
flow build binary ./bin/flow-dev
./bin/flow-dev browse
```

## Development Tools

### Required Tools <!-- {docsify-ignore} -->

These are installed automatically by `flow install tools`:

- [mockgen](https://github.com/uber-go/mock) - Generate test mocks
- [golangci-lint](https://golangci-lint.run/) - Code linting
- [go-jsonschema](https://github.com/atombender/go-jsonschema) - Generate Go types from YAML schemas

### Additional Tools <!-- {docsify-ignore} -->

- [goreleaser](https://goreleaser.com/) - Release automation
- [ginkgo](https://onsi.github.io/ginkgo/) - BDD testing framework
