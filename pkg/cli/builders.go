package cli

import (
	"github.com/spf13/cobra"

	"github.com/flowexec/flow/cmd"
	"github.com/flowexec/flow/pkg/context"
)

// BuildRootCommand creates a new root command with optional configuration.
// The root command is the main entry point for the CLI.
//
// By default, this creates a root command with Flow's standard configuration.
// Use RootOption functions to customize the command.
//
// Example:
//
//	rootCmd := cli.BuildRootCommand(ctx,
//	    cli.WithVersion("1.0.0-custom"),
//	    cli.WithShort("My custom Flow CLI"),
//	)
func BuildRootCommand(ctx *context.Context, opts ...RootOption) *cobra.Command {
	// Create the root command using Flow's standard builder
	rootCmd := cmd.NewRootCmd(ctx)

	// Apply any custom configuration
	if len(opts) > 0 {
		config := &RootConfig{}
		for _, opt := range opts {
			opt(config)
		}

		if config.Use != "" {
			rootCmd.Use = config.Use
		}
		if config.Short != "" {
			rootCmd.Short = config.Short
		}
		if config.Long != "" {
			rootCmd.Long = config.Long
		}
		if config.Version != "" {
			rootCmd.Version = config.Version
		}
		if config.PersistentPreRun != nil {
			rootCmd.PersistentPreRun = config.PersistentPreRun
		}
		if config.PersistentPostRun != nil {
			rootCmd.PersistentPostRun = config.PersistentPostRun
		}
	}

	return rootCmd
}

// RegisterAllCommands registers all Flow commands to the root command.
// This includes: exec, browse, config, secret, vault, cache, workspace, template, logs, sync, mcp.
//
// This is a convenience function that calls all the individual Register*Command functions.
//
// Example:
//
//	rootCmd := cli.BuildRootCommand(ctx)
//	cli.RegisterAllCommands(ctx, rootCmd)
func RegisterAllCommands(ctx *context.Context, rootCmd *cobra.Command) {
	cmd.RegisterSubCommands(ctx, rootCmd)
}

// Execute runs the root command. This is a convenience wrapper around cobra's Execute.
// It configures the command's IO streams and executes it.
//
// Example:
//
//	rootCmd := cli.BuildRootCommand(ctx)
//	cli.RegisterAllCommands(ctx, rootCmd)
//	if err := cli.Execute(ctx, rootCmd); err != nil {
//	    log.Fatal(err)
//	}
func Execute(ctx *context.Context, rootCmd *cobra.Command) error {
	return cmd.Execute(ctx, rootCmd)
}

// BuildExecCommand creates the exec command with all its configuration.
// The exec command is used to execute workflows by their verb and reference.
//
// This command is typically registered automatically via RegisterAllCommands,
// but can be built separately for customization or replacement.
//
// Note: Individual command builders require commands to be registered to a parent
// first, so this creates a temporary parent and returns the registered command.
//
// Example:
//
//	execCmd := cli.BuildExecCommand(ctx)
//	// Customize execCmd...
//	rootCmd.AddCommand(execCmd)
func BuildExecCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "exec")
}

// BuildBrowseCommand creates the browse command for browsing executables.
//
// Example:
//
//	browseCmd := cli.BuildBrowseCommand(ctx)
//	rootCmd.AddCommand(browseCmd)
func BuildBrowseCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "browse")
}

// BuildConfigCommand creates the config command for managing Flow configuration.
//
// Example:
//
//	configCmd := cli.BuildConfigCommand(ctx)
//	rootCmd.AddCommand(configCmd)
func BuildConfigCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "config")
}

// BuildSecretCommand creates the secret command for managing secrets.
//
// Example:
//
//	secretCmd := cli.BuildSecretCommand(ctx)
//	rootCmd.AddCommand(secretCmd)
func BuildSecretCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "secret")
}

// BuildVaultCommand creates the vault command for managing the secrets vault.
//
// Example:
//
//	vaultCmd := cli.BuildVaultCommand(ctx)
//	rootCmd.AddCommand(vaultCmd)
func BuildVaultCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "vault")
}

// BuildCacheCommand creates the cache command for managing Flow's cache.
//
// Example:
//
//	cacheCmd := cli.BuildCacheCommand(ctx)
//	rootCmd.AddCommand(cacheCmd)
func BuildCacheCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "cache")
}

// BuildWorkspaceCommand creates the workspace command for managing workspaces.
//
// Example:
//
//	wsCmd := cli.BuildWorkspaceCommand(ctx)
//	rootCmd.AddCommand(wsCmd)
func BuildWorkspaceCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "workspace")
}

// BuildTemplateCommand creates the template command for managing templates.
//
// Example:
//
//	templateCmd := cli.BuildTemplateCommand(ctx)
//	rootCmd.AddCommand(templateCmd)
func BuildTemplateCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "template")
}

// BuildLogsCommand creates the logs command for viewing execution logs.
//
// Example:
//
//	logsCmd := cli.BuildLogsCommand(ctx)
//	rootCmd.AddCommand(logsCmd)
func BuildLogsCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "logs")
}

// BuildSyncCommand creates the sync command for synchronizing caches.
//
// Example:
//
//	syncCmd := cli.BuildSyncCommand(ctx)
//	rootCmd.AddCommand(syncCmd)
func BuildSyncCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "sync")
}

// BuildMCPCommand creates the mcp command for MCP server management.
//
// Example:
//
//	mcpCmd := cli.BuildMCPCommand(ctx)
//	rootCmd.AddCommand(mcpCmd)
func BuildMCPCommand(ctx *context.Context) *cobra.Command {
	return buildCommand(ctx, "mcp")
}

// buildCommand is a helper that creates a temporary root, registers all commands,
// and returns the requested command. This is necessary because the cmd package
// uses internal registration functions that we can't access directly.
func buildCommand(ctx *context.Context, name string) *cobra.Command {
	tempRoot := &cobra.Command{}
	cmd.RegisterSubCommands(ctx, tempRoot)

	// Find and return the command with the matching name
	for _, c := range tempRoot.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
