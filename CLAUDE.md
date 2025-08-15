# flow repo LLM Context

## Project Overview

**flow** is a workflow automation hub that helps with organizing automation across multiple projects (workspaces) with built-in secrets, templates, and cross-workspace composition. Users define workflows in YAML flow files, discover them visually, and run them anywhere.

This is the main repository for the flow CLI tool and desktop application, written in Go with additional desktop components in Rust/React/TypeScript.

## Repository Structure

```
flow/
├── cmd/                # CLI entry point and command handlers
├── internal/           # Core application logic
│   ├── cache/          # Executable and workspace caching logic  
│   ├── context/        # Global application context
│   ├── io/             # Terminal user interface and I/O
│   ├── runner/         # Executable execution engine
│   ├── services/       # Business logic services
│   ├── templates/      # Templating system for workflows
│   └── vault/          # Secret management
├── types/              # Generated Go types from YAML schemas
├── tests/              # CLI end-to-end test suite
├── docs/               # Documentation (hosted at flowexec.io)
├── desktop/            # Desktop application (Tauri + React + TypeScript)
│   ├── src/            # React frontend code
│   ├── src-tauri/      # Rust backend code
│   └── scripts/        # Build and type generation scripts
└── tools/              # Code generation and build tools
```

## Key Technologies & Frameworks

### Go CLI Application
- **Language**: Go 1.24+ (see go.mod:3)
- **CLI Framework**: Cobra (github.com/spf13/cobra)
- **TUI Framework**: Custom tuikit (github.com/flowexec/tuikit) based on github.com/charmbracelet/bubbletea 
- **Testing**: Ginkgo BDD framework (github.com/onsi/ginkgo/v2)

### Desktop Application  
- **Frontend**: React 18 with TypeScript
- **UI Library**: Mantine v8 (@mantine/core)
- **Backend**: Rust with Tauri v2
- **Build Tool**: Vite
- **Testing**: Vitest with Storybook for component development

## Development Workflow

The project uses **flow itself** for development automation. Key commands:

```bash
# Build the CLI binary
flow build binary ./bin/flow

# Run all validation (tests, linting, code generation)
flow validate

# Run specific checks
flow test                 # All tests
flow generate             # Code generation  
flow lint                 # Linting only
flow install tools        # Install/update Go tools
```

These "executables" are defined in the flow files in the `.execs` directory.

## Go Testing Strategy

- **Go Tests**: Uses Ginkgo BDD framework for both unit and e2e tests
- **Location**: Unit tests in `internal/*_test.go`, e2e tests in `tests/`
- **Run Command**: `flow test` or standard `go test`
- **Focusing test**: `FDescribe`, `FIt`, `FEntry`, etc. should be used temporarily to filter when troubleshooting / writing tests

## Code Generation

The project heavily uses code generation:

1. **Go Types**: Generated from YAML schemas in `types/*/schema.yaml` using go-jsonschema
2. **Documentation**: CLI and type docs auto-generated from the go schema definitions
3. **TypeScript and Rust Types**: Generated for desktop app from JSON schemas that docgen creates

**Important**: Always edit schema files, not generated code!

## Configuration Files

- **flow.yaml**: Workspace configuration for the flow repo itself
- **go.mod**: Go dependencies and version (Go 1.24+)
- **desktop/package.json**: Node.js dependencies for desktop app
- **desktop/src-tauri/tauri.conf.json**: Tauri desktop app configuration
- **.execs/**: flow development workflows (executables)

## Development Setup

1. **Prerequisites**: Go 1.24+, flow CLI installed
2. **Setup**: `flow workspace add flow . --set`
3. **Dependencies**: `flow install tools`
4. **Verification**: `flow validate`

## Important Context for Claude Code Sessions

- **Testing**: Always use `flow test` or `go test` - Ginkgo is the testing framework
- **Linting**: Use `flow lint` for Go linting
- **Code Gen**: Run `flow generate` after schema changes
- **Desktop**: The `desktop/` directory is a separate Tauri/React app
  - Use `flow build desktop` to build the Tauri app
  - Code should be written in a way that would enable easy testing in the future
- **Documentation**: Lives in `docs/` and is hosted at flowexec.io
- **Build**: Use `flow build binary ./bin/flow` for development CLI builds
