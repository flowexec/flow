# Flow CLI Extension API

The `pkg/cli` package provides a public API for extending and customizing the Flow CLI. This allows external projects to build custom CLIs that include Flow's commands, add cross-cutting functionality, and override existing behavior.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [API Overview](#api-overview)
  - [Command Builders](#command-builders)
  - [Hook Injection](#hook-injection)
  - [Command Registry](#command-registry)
  - [Root Command Customization](#root-command-customization)
- [Examples](#examples)
- [API Reference](#api-reference)
- [Best Practices](#best-practices)

## Features

- **Command Builders**: Create Flow commands programmatically
- **Hook Injection**: Add PreRun/PostRun hooks to any command
- **Command Registry**: Manipulate the command tree (find, replace, remove)
- **Root Customization**: Customize the root command with functional options
- **Backward Compatible**: Works alongside standard Flow CLI

## Installation

```bash
go get github.com/flowexec/flow
```

Then import the package:

```go
import "github.com/flowexec/flow/pkg/cli"
```

## Quick Start

Here's a minimal example of building a custom Flow CLI:

```go
package main

import (
    stdCtx "context"

    "github.com/flowexec/flow/pkg/context"
    "github.com/flowexec/flow/pkg/filessystem"
    "github.com/flowexec/flow/internal/io"
    "github.com/flowexec/flow/pkg/logger"
    "github.com/flowexec/flow/pkg/cli"
)

func main() {
    // Setup (same as standard Flow CLI)
    cfg, _ := filesystem.LoadConfig()
    logger.Init(logger.InitOptions{
        StdOut:  io.Stdout,
        LogMode: cfg.DefaultLogMode,
        Theme:   io.Theme(cfg.Theme.String()),
    })
    defer logger.Log().Flush()

    // Create context
    bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
    ctx := context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
    defer ctx.Finalize()

    // Build custom CLI
    rootCmd := cli.BuildRootCommand(ctx,
        cli.WithVersion("1.0.0-custom"),
    )
    cli.RegisterAllCommands(ctx, rootCmd)

    // Execute
    cli.Execute(ctx, rootCmd)
}
```

## API Overview

### Command Builders

Create Flow commands programmatically:

```go
// Build root command
rootCmd := cli.BuildRootCommand(ctx,
    cli.WithVersion("1.0.0"),
    cli.WithShort("My custom CLI"),
)

// Register all commands at once
cli.RegisterAllCommands(ctx, rootCmd)

// Or build individual commands
execCmd := cli.BuildExecCommand(ctx)
wsCmd := cli.BuildWorkspaceCommand(ctx)
rootCmd.AddCommand(execCmd, wsCmd)
```

Available command builders:
- `BuildRootCommand(ctx, ...opts)` - Root command
- `BuildExecCommand(ctx)` - Exec command
- `BuildBrowseCommand(ctx)` - Browse command
- `BuildConfigCommand(ctx)` - Config command
- `BuildSecretCommand(ctx)` - Secret command
- `BuildVaultCommand(ctx)` - Vault command
- `BuildCacheCommand(ctx)` - Cache command
- `BuildWorkspaceCommand(ctx)` - Workspace command
- `BuildTemplateCommand(ctx)` - Template command
- `BuildLogsCommand(ctx)` - Logs command
- `BuildSyncCommand(ctx)` - Sync command
- `BuildMCPCommand(ctx)` - MCP command

### Hook Injection

Add cross-cutting functionality to commands:

```go
// Add hooks to a single command
cli.AddPreRunHook(cmd, func(cmd *cobra.Command, args []string) {
    log.Println("Before command:", cmd.Name())
})

cli.AddPostRunHook(cmd, func(cmd *cobra.Command, args []string) {
    log.Println("After command:", cmd.Name())
})

// Add hooks to all commands recursively
cli.ApplyHooksRecursive(rootCmd,
    // PreRun hook
    func(cmd *cobra.Command, args []string) {
        telemetry.Start(cmd.Name())
    },
    // PostRun hook
    func(cmd *cobra.Command, args []string) {
        telemetry.End(cmd.Name())
    },
)

// Add persistent hooks (inherited by subcommands)
cli.AddPersistentPreRunHook(rootCmd, initHook)
cli.AddPersistentPostRunHook(rootCmd, cleanupHook)
```

Hook functions:
- `AddPreRunHook(cmd, hook)` - Add PreRun hook
- `AddPostRunHook(cmd, hook)` - Add PostRun hook
- `AddPersistentPreRunHook(cmd, hook)` - Add PersistentPreRun hook
- `AddPersistentPostRunHook(cmd, hook)` - Add PersistentPostRun hook
- `ApplyHooksRecursive(cmd, preRun, postRun)` - Apply to entire tree
- `ApplyPersistentHooksRecursive(cmd, preRun, postRun)` - Apply persistent to tree

### Command Registry

Manipulate the command tree:

```go
// Find a command
wsCmd := cli.FindCommand(rootCmd, "workspace")

// Find by path
addCmd := cli.FindCommandPath(rootCmd, "workspace add")

// Replace a command
customExec := &cobra.Command{...}
cli.ReplaceCommand(rootCmd, "exec", customExec)

// Remove a command
cli.RemoveCommand(rootCmd, "sync")

// Walk all commands
cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
    fmt.Println("Command:", cmd.Name())
})

// List all command names
names := cli.ListCommands(rootCmd)

// Get subcommands
subCmds := cli.GetSubcommands(wsCmd)
```

Registry functions:
- `WalkCommands(root, fn)` - Traverse command tree
- `FindCommand(root, name)` - Find command by name
- `FindCommandPath(root, path)` - Find by full path
- `ReplaceCommand(root, name, newCmd)` - Replace command
- `RemoveCommand(root, name)` - Remove command
- `ListCommands(root)` - Get all command names
- `GetSubcommands(cmd)` - Get subcommands map
- `CloneCommand(cmd)` - Clone a command

### Root Command Customization

Use functional options to customize the root command:

```go
rootCmd := cli.BuildRootCommand(ctx,
    cli.WithUse("mycli"),
    cli.WithShort("My custom CLI description"),
    cli.WithLong("Detailed description..."),
    cli.WithVersion("1.0.0"),
    cli.WithPersistentPreRun(func(cmd *cobra.Command, args []string) {
        // Global initialization
    }),
    cli.WithPersistentPostRun(func(cmd *cobra.Command, args []string) {
        // Global cleanup
    }),
)
```

Available options:
- `WithUse(use)` - Set Use field
- `WithShort(short)` - Set Short description
- `WithLong(long)` - Set Long description
- `WithVersion(version)` - Set Version string
- `WithPersistentPreRun(hook)` - Set PersistentPreRun hook
- `WithPersistentPostRun(hook)` - Set PersistentPostRun hook

## Examples

See the [examples](./examples) directory for complete working examples:

- **[basic](./examples/basic/main.go)** - Basic CLI with custom version
- **[hooks](./examples/hooks/main.go)** - Adding telemetry hooks to all commands
- **[override](./examples/override/main.go)** - Overriding commands and adding new ones

## API Reference

### Types

```go
// HookFunc is a function that can be used as a PreRun or PostRun hook
type HookFunc func(cmd *cobra.Command, args []string)

// RootConfig holds configuration options for building the root command
type RootConfig struct {
    Use              string
    Short            string
    Long             string
    Version          string
    PersistentPreRun HookFunc
    PersistentPostRun HookFunc
}

// RootOption is a functional option for configuring the root command
type RootOption func(*RootConfig)
```

### Core Functions

```go
// Build and execute
BuildRootCommand(ctx *context.Context, opts ...RootOption) *cobra.Command
RegisterAllCommands(ctx *context.Context, rootCmd *cobra.Command)
Execute(ctx *context.Context, rootCmd *cobra.Command) error

// Command builders
BuildExecCommand(ctx *context.Context) *cobra.Command
BuildBrowseCommand(ctx *context.Context) *cobra.Command
BuildConfigCommand(ctx *context.Context) *cobra.Command
BuildSecretCommand(ctx *context.Context) *cobra.Command
BuildVaultCommand(ctx *context.Context) *cobra.Command
BuildCacheCommand(ctx *context.Context) *cobra.Command
BuildWorkspaceCommand(ctx *context.Context) *cobra.Command
BuildTemplateCommand(ctx *context.Context) *cobra.Command
BuildLogsCommand(ctx *context.Context) *cobra.Command
BuildSyncCommand(ctx *context.Context) *cobra.Command
BuildMCPCommand(ctx *context.Context) *cobra.Command

// Hook injection
AddPreRunHook(cmd *cobra.Command, hook HookFunc)
AddPostRunHook(cmd *cobra.Command, hook HookFunc)
AddPersistentPreRunHook(cmd *cobra.Command, hook HookFunc)
AddPersistentPostRunHook(cmd *cobra.Command, hook HookFunc)
ApplyHooksRecursive(cmd *cobra.Command, preRun, postRun HookFunc)
ApplyPersistentHooksRecursive(cmd *cobra.Command, preRun, postRun HookFunc)
WrapRunFunc(cmd *cobra.Command, before, after HookFunc)
WrapRunEFunc(cmd *cobra.Command, before, after HookFunc)

// Command registry
WalkCommands(rootCmd *cobra.Command, fn func(*cobra.Command))
FindCommand(rootCmd *cobra.Command, name string) *cobra.Command
FindCommandPath(rootCmd *cobra.Command, path string) *cobra.Command
ReplaceCommand(rootCmd *cobra.Command, oldName string, newCmd *cobra.Command) error
RemoveCommand(rootCmd *cobra.Command, name string) error
ListCommands(rootCmd *cobra.Command) []string
GetSubcommands(cmd *cobra.Command) map[string]*cobra.Command
CloneCommand(cmd *cobra.Command) *cobra.Command

// Root options
WithUse(use string) RootOption
WithShort(short string) RootOption
WithLong(long string) RootOption
WithVersion(version string) RootOption
WithPersistentPreRun(hook HookFunc) RootOption
WithPersistentPostRun(hook HookFunc) RootOption
```

## Best Practices

### 1. Hook Ordering

Hooks are chained in the order they're added:
- **PreRun**: New hooks run before existing hooks
- **PostRun**: New hooks run after existing hooks

```go
// This hook runs first
cli.AddPreRunHook(cmd, hook1)
// This hook runs second
cli.AddPreRunHook(cmd, hook2)
```

### 2. Context Management

Always create and finalize the Flow context properly:

```go
bkgCtx, cancelFunc := stdCtx.WithCancel(stdCtx.Background())
ctx := context.NewContext(bkgCtx, cancelFunc, io.Stdin, io.Stdout)
defer ctx.Finalize() // Important!
```

### 3. Command Replacement

When replacing commands, ensure the new command has the same `Use` field:

```go
// Bad: Use field doesn't match
newCmd := &cobra.Command{Use: "execute", ...}
cli.ReplaceCommand(rootCmd, "exec", newCmd) // Won't work as expected

// Good: Use field matches
newCmd := &cobra.Command{Use: "exec", ...}
cli.ReplaceCommand(rootCmd, "exec", newCmd) // Works correctly
```

### 4. Error Handling

Always check errors from registry operations:

```go
if err := cli.ReplaceCommand(rootCmd, "exec", newCmd); err != nil {
    log.Fatalf("Failed to replace command: %v", err)
}
```

### 5. Hook Safety

Hooks should be safe to call multiple times and handle nil values:

```go
// Good: Safe hook implementation
myHook := func(cmd *cobra.Command, args []string) {
    if cmd == nil {
        return
    }
    // Do work...
}
```

### 6. Building Individual Commands

When building individual commands instead of using `RegisterAllCommands`, remember that some commands may have dependencies:

```go
// Build commands with potential dependencies
execCmd := cli.BuildExecCommand(ctx)
cacheCmd := cli.BuildCacheCommand(ctx) // Exec may depend on cache

rootCmd.AddCommand(execCmd, cacheCmd)
```

## Thread Safety

This package is **not thread-safe**. Command building and modification should be done during CLI initialization, not concurrently.

## Versioning

This package follows semantic versioning:
- **Major**: Breaking changes to the public API
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

## Contributing

When contributing to this package:
1. Maintain backward compatibility
2. Add comprehensive godoc comments
3. Include examples for new features
4. Write tests for all public functions
5. Update this README with new features

## License

This package is part of the Flow CLI project and follows the same license.
