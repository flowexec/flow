// Package cli provides a public API for extending and customizing the Flow CLI.
//
// This package allows external projects to:
//   - Build custom CLIs that include Flow's commands
//   - Add cross-cutting hooks (PreRun/PostRun) to commands
//   - Override existing commands with custom implementations
//   - Add new commands alongside Flow's built-in commands
//
// # Basic Usage
//
// Build a custom CLI with all Flow commands:
//
//	ctx := context.NewContext(...)
//	rootCmd := cli.BuildRootCommand(ctx)
//	cli.RegisterAllCommands(ctx, rootCmd)
//	if err := cli.Execute(ctx, rootCmd); err != nil {
//	    log.Fatal(err)
//	}
//
// # Customizing the Root Command
//
// Use functional options to customize the root command:
//
//	rootCmd := cli.BuildRootCommand(ctx,
//	    cli.WithVersion("1.0.0-custom"),
//	    cli.WithShort("My custom Flow CLI"),
//	    cli.WithPersistentPreRun(myPreRunHook),
//	)
//
// # Adding Hooks
//
// Add hooks to individual commands or recursively to all commands:
//
//	// Add to a single command
//	cli.AddPreRunHook(cmd, telemetryHook)
//
//	// Add to all commands recursively
//	cli.ApplyHooksRecursive(rootCmd, preHook, postHook)
//
// # Command Registry Operations
//
// Manipulate the command tree:
//
//	// Find a command
//	wsCmd := cli.FindCommand(rootCmd, "workspace")
//
//	// Replace a command
//	cli.ReplaceCommand(rootCmd, "exec", customExecCmd)
//
//	// Walk all commands
//	cli.WalkCommands(rootCmd, func(cmd *cobra.Command) {
//	    // Do something with each command
//	})
//
// # Thread Safety
//
// This package is not thread-safe. Command building and modification should be
// done during CLI initialization, not concurrently.
//
// # Versioning
//
// This package follows semantic versioning. The API is considered stable but may
// receive additions in minor version updates. Breaking changes will only occur
// in major version updates.
package cli
